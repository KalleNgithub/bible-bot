# Proposal

## Why

There is no simple, cheap bot that answers a free-text life situation ("my mom died and I require consolation") with a fitting Bible quote, or returns a specific verse from a loosely typed reference. The project is greenfield; the goal is a low-cost MVP serving Finnish and English.

## What Changes

- New Go service exposing `GET /?q=...&lang=...` that returns a plain-text quote.
- Free-text topic search over one English Bible (WEB), with randomized selection among the best matches.
- Direct verse lookup from typo-tolerant references ("Jhon 11:25 to 26").
- Automatic routing between topic search and verse lookup when the caller does not say which is wanted.
- Output in Finnish (Pyha Raamattu 1933/38) or English, using verse numbers found in the English search. Any input language is accepted and translated to English.
- Simple context expansion so returned quotes complete unfinished sentences, capped in length.
- Build-time tool that downloads WEB and the Finnish text from the getBible API and embeds the WEB verses once into a static vector file shipped with the service.
- Out of scope: output languages other than Finnish and English, multiple English versions, conversation history, LLM-judged context expansion, hosted vector databases.

## Capabilities

### New Capabilities
- `verse-lookup`: Parse and typo-correct verse references (book aliases in English and Finnish, ranges, spelled-out forms) and return the verses in the requested language.
- `topic-quote-search`: Translate a free-text query to English, run semantic search over the single English Bible, and pick a varying result from the top matches above a threshold.
- `quote-context-expansion`: Extend a matched verse backward and forward to complete unfinished sentences, with a hard cap on verse count.
- `query-routing`: Decide whether an unspecified query is a verse reference or a topic query.
- `quote-http-api`: HTTP GET interface taking string parameters and returning a plain-text body.

### Modified Capabilities

None.

## Impact

- New Go module, binary, and build-time data tool; no existing code affected.
- External dependencies: getBible API (build time only), Azure OpenAI (embeddings and a small chat model at runtime).
- Hosting on a scale-to-zero Azure service; expected cost in single-digit currency units per month.
- New `infra/terraform/dev` Terraform in this monorepo: resource group, Azure OpenAI with embedding and chat deployments, budget alert, and later container hosting. Remote state reuses the existing kalevala-bot state storage account.
- Static data artifacts (WEB and Finnish verse text, WEB vectors) bundled with the binary.
