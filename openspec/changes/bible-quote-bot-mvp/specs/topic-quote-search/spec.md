# Spec Delta

## Purpose

Finds a fitting Bible quote for a free-text query in any input language and returns it in the output language, with some variety between repeated queries.

## ADDED Requirements

### Requirement: Search a single English Bible
The system SHALL translate the query to English and search only one English Bible version; no other English version SHALL be stored or indexed. The search input SHALL combine the English query with a short generated description of what a passage answering the person would say, so that results answer the query rather than echo its words.

#### Scenario: Answering rather than echoing
- **WHEN** the query is "I am afraid"
- **THEN** the search input includes a description such as "do not be afraid, God is with you"
- **AND** results are not limited to verses that merely contain the words "I am afraid"

#### Scenario: Finnish query
- **WHEN** the query is "äitini kuoli ja kaipaan lohtua"
- **THEN** the system searches the English index using an English translation of the query

#### Scenario: Consolation query
- **WHEN** the query is "my mom died and I require consolation"
- **THEN** the selected verse comes from verses relevant to comfort in grief

### Requirement: Score-weighted varied results
The system SHALL choose the result randomly among the top 10 matches whose score is at or above a threshold, with each match's probability increasing with its score (softmax over scores with a temperature), so a near-verbatim quote returns its own verse almost always while vague queries return varied verses.

#### Scenario: Vague query varies
- **WHEN** a query matches several verses with similar scores and is sent many times
- **THEN** more than one distinct verse is returned across the responses
- **AND** higher-scoring verses are returned more often than lower-scoring ones

#### Scenario: Near-verbatim quote
- **WHEN** the query is a near-verbatim quote of one verse whose score is far above all other matches
- **THEN** that verse is returned in almost every response

#### Scenario: Few qualifying matches
- **WHEN** fewer than 10 matches meet the threshold
- **THEN** the choice is made only among those that qualify

### Requirement: Always return a verse
The system SHALL return a verse for every non-empty topic query. When no match meets the threshold it SHALL run a lenient search over the nearest matches regardless of score, choosing from a larger pool with a much higher temperature so the choice is far more random.

#### Scenario: Unrelated query
- **WHEN** the query is "how do I bake sourdough bread" and no match meets the threshold
- **THEN** a verse is still returned, chosen from the nearest matches by the lenient search

#### Scenario: Nonsense query
- **WHEN** the query is "asdfgh"
- **THEN** a verse is still returned

### Requirement: Work without the search service
If the search cannot run (for example Azure OpenAI is unavailable), the system SHALL match words from the query against verse text, and failing that return a random verse.

#### Scenario: Search unavailable
- **WHEN** the embedding call fails
- **THEN** a verse containing the most query words is returned, or a random verse if no word matches

### Requirement: Output language by verse number
The system SHALL return the selected verses from the output-language Bible by book, chapter and verse number, not by translating the English text.

#### Scenario: Finnish output
- **WHEN** Finnish output is requested and John 11:25-26 is selected
- **THEN** the response contains the Finnish text of John 11:25-26
