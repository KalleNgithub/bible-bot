# bible-bot

HTTP GET service that answers with a Bible quote as plain text. Ask with free text ("my mom died and I require consolation") or with a verse reference, typos allowed ("Jhon 11:25 to 26").

## API

```
GET /?q=<text>[&lang=fi|en][&mode=verse|topic]
```

- `q`: required, at most 300 characters.
- `lang`: output language. Default: Finnish if the query is Finnish, otherwise English.
- `mode`: force `verse` lookup or `topic` search. Default: a query that parses as a reference is a lookup, anything else a topic search.

```bash
curl "$URL/?q=Jhon%2011:25%20to%2026"
curl "$URL/?q=my%20mom%20died%20and%20I%20require%20consolation"
curl "$URL/?q=%C3%A4itini%20kuoli%20ja%20kaipaan%20lohtua"
```

Errors are plain text with status 400 (bad parameters) or 404 (reference not found). Any free-text query returns some verse.

## How it works

Topic search translates the query to English with a small chat model, embeds it with `text-embedding-3-small` and compares it to precomputed vectors of every WEB verse. A softmax-weighted random pick among the top matches gives varied answers. Short or unfinished verses are widened to neighbouring verses (at most 4). The same verse numbers are then read from the output-language text. If nothing scores well, a looser search runs; if Azure OpenAI is unavailable, word matching is used.

## Data

- English: World English Bible (public domain).
- Finnish: Finnish Bible 1933/38 (public domain). The language is archaic.

Rebuild the data and vectors (needs Azure OpenAI env vars):

```bash
make data
go run ./cmd/build-data embed
```

## Run locally

```bash
export AZURE_OPENAI_ENDPOINT=... AZURE_OPENAI_KEY=... AZURE_OPENAI_EMBED_DEPLOYMENT=text-embedding-3-small
export AZURE_OPENAI_CHAT_DEPLOYMENT=gpt-5.4-nano AZURE_OPENAI_REASONING_EFFORT=none
go run ./cmd/server        # PORT defaults to 8080
make test
```

Reference lookups work without a valid Azure key. Quality evaluation: `go run ./cmd/eval -runs 3`.

## Deploy and cost

See `infra/README.md`. Azure Container Apps scales to zero and the image is on public GHCR, so idle cost is near zero. Each topic query costs a fraction of a cent (one chat and one embedding call). Verse lookups cost nothing beyond compute.
