# Design

## Context

Greenfield Go project. See proposal.md for motivation. Constraints: minimal cost (scale to zero, single-digit monthly spend), no Python, one English Bible only, one data API for the MVP, Azure as cloud.

Data source: getBible API v2 (`api.getbible.net/v2/<translation>/<book_nr>/<chapter>.json`) provides both `web` (English, public domain) and `pyharaamattu1933` (Finnish 1933/38, public domain) with identical book numbering (John = 43). Verified against the live API.

## Goals / Non-Goals

**Goals:**
- Single static Go binary with data embedded; no database or vector service.
- Per-query cost well under one cent.
- Deterministic parts (reference parsing, context expansion) need no LLM.

**Non-Goals:**
- Output languages beyond Finnish and English.
- LLM-judged relevance or context expansion.
- Caching, auth, conversation memory.

## Decisions

**Request flow**
```
GET /?q&lang&mode
   |
   v
[validate] -> mode=verse|topic given? ---------------------+
   | no                                                     |
   v                                                        |
[reference parser] -- valid ref --> verse path              |
   | not a ref                                              v
   v                                                  forced path
[LLM: translate to English + detect input language + describe an answering passage]
   v
[embed "<english>. <answer description>"] -> [cosine over WEB vectors] -> top 10 >= threshold -> random pick
   v
[context expansion on WEB text] -> verse range
   v
[fetch same range from output-language Bible] -> plain text
```

1. **Static data embedded in the binary.** A build-time tool (`cmd/build-data`) downloads WEB and Finnish from getBible, writes compact verse files, embeds WEB verses once, and writes a vector file. All are committed or generated and `go:embed`ded. Alternative: query getBible at runtime; rejected because it adds latency, an availability dependency and rate-limit risk. The flat vector file (31k x 512 float32, about 64 MB, or int8-quantised about 16 MB) is searched by brute-force dot product, which takes milliseconds.
2. **Azure OpenAI for embeddings and chat.** `text-embedding-3-small` reduced to 512 dimensions for both build and query (same model is mandatory); a small chat model (`gpt-5.4-nano`; `gpt-4o-mini` was deprecating) for one call with `reasoning_effort` none that returns JSON: English translation, detected input language, and a one or two sentence description of what a passage answering the person would say. The query embedding is of "english. description". Embedding the bare query matched verses that voice the same complaint (for example "I am afraid" returned "I am afraid of your judgments"); adding the description raised hit@10 from 23/29 to 29/29 and the chance of an acceptable pick from about 0.30 to 0.46 on the evaluation set, with no extra call. Alternative: local ONNX model; rejected for cgo and package size, revisit only if cost matters.
3. **Reference parsing is deterministic and runs first.** Normalise the query, find a book token by edit-distance match against an alias table (English and Finnish names and abbreviations), then parse chapter, verse and range with number words ("to", "and", "chapter", "verses", Finnish "luku", "jae"). A match is accepted only below a typo-distance limit and with an existing chapter and verse. This is the whole of the routing evaluation: valid reference means verse, otherwise topic. Alternative: LLM classification on every query; rejected for cost and latency.
4. **Randomisation.** Take the top 10 results scoring at or above a threshold and sample with probability proportional to `exp(score / T)` (softmax). Linear score weights are rejected: related verses score in a narrow band (about 0.55-0.65), so they would be nearly uniform. With T around 0.03, a near-verbatim quote (score about 0.9) wins almost always, while a vague query with scores within a few hundredths still varies. Threshold and T are tuned together against the evaluation set (task 6.3), not guessed.
5. **Context expansion on the English text only**, then apply the resulting range to the Finnish text. Rules: extend back if the first verse starts lowercase or with a dependent-clause word (that, who, which, because, or, nor); extend forward if the last verse lacks terminal punctuation (`. ! ?` possibly followed by a closing quote). An unclosed quotation does not extend the range: it pulled whole speeches such as the Beatitudes (Matt 5:3-6) into one answer. Capitalised "And", "But" and "For" are not treated as continuations: they open complete sentences in WEB and including them pulled John 3:16 back to verse 14. A closing quote alone does not extend backward. John 11:25 and 11:26 each stand alone. Cap at 4 verses; when over the cap, keep the window containing the hit. Never cross chapter boundaries. Applying the English range to Finnish keeps verse alignment simple; versification differences between WEB and the 1933/38 text are rare and accepted for the MVP.
6. **Language default.** `lang` explicit wins. Otherwise the LLM-detected input language decides: Finnish gives `fi`, anything else `en`. Verse-path queries with no `lang` use Finnish if the book name matched a Finnish alias, otherwise English.
7. **Hosting.** One container running the Go binary on Azure Container Apps with min replicas 0 (scale to zero). Secrets via environment variables or managed identity. Alternative: Azure Functions custom handler; equivalent, container chosen for plain Go ergonomics.
8. **Infrastructure as code in this monorepo.** All Azure resources live in `infra/terraform/dev`, mirroring the kalevala-bot layout (`versions.tf`, `variables.tf`, `main.tf`, `outputs.tf`, `terraform.tfvars.example`) and the same subscription and region (`westeurope`, subject to model availability). State uses the existing `kalevalatfstate` storage account with key `bible-dev.tfstate`, so no new bootstrap. Resources: `rg-bible-dev`, an Azure OpenAI account with embedding and chat deployments, a monthly budget alert (alerts only, same as kalevala-bot), and a Container App. The image is pulled from public GHCR; Azure Container Registry Basic (about 4.6 EUR/month) was rejected for cost. Alternative: share kalevala-bot's resource group; rejected so the two projects can be destroyed independently. Rule carried over: the agent runs only `init`, `fmt`, `validate` and `plan`; the user runs `apply`. The Azure OpenAI key is fetched with `az` into an environment variable and never committed.
9. **Tiered fallback, every topic query returns a verse.** The model never rejects a query. Tier 1 is the normal search (top 10 at or above the threshold, softmax T about 0.03). If nothing qualifies, tier 2 reuses the same query vector, ignores the threshold, takes a larger pool (top 50) and samples with a much higher temperature (T about 0.1, to be tuned) so the result is close to a random pick among the nearest verses. If embedding or the chat call fails, tier 3 matches non-stopword query words against the English verse text and picks randomly among the verses with the most hits, and if none match picks any random verse. If only the chat call fails, the raw query is embedded and tiers 1 and 2 still apply. The relevance score threshold therefore only chooses between tier 1 and tier 2 and no longer needs to separate on-topic from off-topic queries precisely.

## Risks / Trade-offs

- [Embedding matches are loose, silly results] -> Evaluation set of about 30 queries with acceptable verses (`eval/queries.tsv`, `go run ./cmd/eval`). Chosen values: tier 1 threshold 0.50 and T=0.03. Over 3 runs of 32 queries no on-topic run fell below 0.50 (lowest 0.509), while 7 of 9 off-topic runs did (highest 0.521) and fall to tier 2. The off-topic queries that pass 0.50 still get a real verse, which is acceptable under the always-return rule. The description varies between runs and the sample is small; revisit with more queries after deploy. "bake sourdough bread" is the closest off-topic case because the Bible has bread verses.
- [Finnish 1933/38 uses archaic language] -> Accepted; it is the only open Finnish text in the chosen API. Document in the README.
- [Versification mismatch WEB vs Finnish] -> Rare; the range is applied by number and a missing verse is skipped. Revisit if observed.
- [Heuristic expansion pulls in irrelevant verses] -> Hard cap of 4 verses and chapter boundary limit.
- [Cold start on scale-to-zero] -> Loading the vector file is the main cost; keep it int8 or float32 flat and measure. Accept a few seconds on the first request.
- [Query text sent to Azure OpenAI] -> Queries can be personal ("my mom died"); do not log full queries, state this in the README.

## Open Questions

- Exact chat model and Azure region; settled when the Azure resource is created.
- Whether to commit generated data files or generate them in CI; decided during build-data task.
