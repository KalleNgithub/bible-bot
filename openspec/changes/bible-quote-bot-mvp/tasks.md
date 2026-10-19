# Tasks

## 1. Project setup

- [x] 1.1 Create Go module with `cmd/server`, `cmd/build-data` and `internal/` packages; verify `go build ./...` succeeds
- [x] 1.2 Add a Makefile with `build`, `test` and `data` targets; verify `make test` runs

## 2. Infrastructure (Terraform in `infra/terraform/dev`)

- [x] 2.1 Add `infra/terraform/bootstrap` (shared state storage account, applied by the user) and scaffold `versions.tf`, `variables.tf`, `outputs.tf` and `terraform.tfvars.example` with an azurerm backend on it (key `bible-dev.tfstate`); verify `terraform init`, `fmt -check` and `validate` pass
- [x] 2.2 Add resource group `rg-bible-dev`, an Azure OpenAI account with `text-embedding-3-small` and a small chat model deployment, and outputs for endpoint and deployment names; verify `terraform plan` shows only the expected creates in a region offering both models
- [x] 2.3 Add a monthly consumption budget alert (about 5 EUR; 50%, 100% and forecast thresholds) following the kalevala-bot pattern; verify the plan includes it
- [x] 2.4 Write `infra/README.md` with the init, plan and apply flow and how to export the key via `az cognitiveservices account keys list` into `AZURE_OPENAI_KEY`; verify the documented commands run as written up to `plan`
- [x] 2.5 User reviews the plan and runs `terraform apply` (never run by the agent); verify the Azure OpenAI deployments exist and the key export works

## 3. Bible data build

- [x] 3.1 Implement `build-data` download of WEB and `pyharaamattu1933` from getBible into compact verse files; verify John 11:25-26 exists in both and verse counts are about 31k
- [x] 3.2 Embed all WEB verses with Azure OpenAI (512 dims) into a flat vector file; verify vector count equals WEB verse count and a spot-check query returns a sensible neighbour
- [x] 3.3 `go:embed` the data and add a loader with tests; verify lookup by book, chapter and verse works for both translations

## 4. Verse lookup

- [x] 4.1 Build book alias table (English and Finnish names, abbreviations); verify every one of the 66 books resolves from its canonical name in both languages
- [x] 4.2 Implement fuzzy reference parser with number words and ranges; verify unit tests cover "Jhon 11:25 to 26", "Jhn chapter11 verses 25 and 26", Finnish names and "John 99:1" rejection
- [x] 4.3 Implement verse retrieval in the requested language; verify tests for single verse, range and Finnish output

## 5. Context expansion

- [x] 5.1 Implement expansion rules with chapter bounds and 4-verse cap; verify table-driven tests for continuation start, unfinished end, self-contained verse, cap and chapter edges
- [x] 5.2 Add a test using real WEB text for John 11:25-26 showing which neighbours the rules include for each verse; verify the test passes and review the result with the user

## 6. Topic search

- [x] 6.1 Implement Azure OpenAI client for translate-and-detect and query embedding; verify with a mocked HTTP test and one manual call after user approval
- [x] 6.2 Implement brute-force cosine search and softmax-weighted random selection (temperature T) among the top 10 above threshold; verify tests for few matches, no matches, a dominant score nearly always winning, and similar scores varying with higher scores chosen more often
- [x] 6.3 Create an evaluation set of about 30 queries with acceptable verses, including near-verbatim quotes, and a script reporting hit rate; use it to set the tier 1 threshold and T, then record the chosen values (the generated description varies between runs, so average several runs; revisit how deterministic the picks are)
- [x] 6.4 Implement the tiered fallback: lenient tier 2 search (top 50, high temperature) and tier 3 word-match then random verse; verify tests for a query with nothing above the threshold, an embedding failure, a chat-only failure, a query with no word matches, and that every non-empty query returns a verse

## 7. HTTP API and routing

- [x] 7.1 Implement the GET handler with `q`, `lang`, `mode` validation, plain-text responses and 400 errors; verify with `httptest` tests
- [x] 7.2 Implement routing (explicit mode, then reference parse, then topic) and language defaulting; verify tests for each scenario in query-routing and quote-http-api specs
- [x] 7.3 Ensure full query text is not logged; verify by test or code review

## 8. Deployment and docs

- [ ] 8.1 Add Dockerfile, then publish the image to public GHCR, then extend the Terraform with a Container Apps environment and a Container App (min replicas 0, key from a secret); verify `terraform plan` is clean and, after the user applies, the deployed URL answers `curl "<url>/?q=John%2011:25"`
- [x] 8.2 Write README with usage examples, data licensing note (WEB and 1933/38 public domain), cost estimate and the archaic-Finnish caveat; verify README commands run as written
- [ ] 8.3 End-to-end check of "my mom died and I require consolation" in Finnish and English; verify varied, complete verses are returned
