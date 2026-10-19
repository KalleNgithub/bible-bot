# Spec Delta

## Purpose

Exposes the bot as a minimal HTTP GET interface that any client can call with string parameters and read as plain text.

## ADDED Requirements

### Requirement: GET with string parameters
The system SHALL accept `q` (required), `lang` (optional, `fi` or `en`) and `mode` (optional, `verse` or `topic`) as query-string parameters on HTTP GET.

#### Scenario: Minimal request
- **WHEN** a client sends GET /?q=my%20mom%20died
- **THEN** the response status is 200

### Requirement: Plain-text response
The system SHALL respond with `Content-Type: text/plain; charset=utf-8` and a body consisting only of the quote with its reference label, such as `Psalm 34:18: "..."`.

#### Scenario: Topic answer
- **WHEN** a topic query succeeds
- **THEN** the body is one reference label followed by the verse text, with no JSON or markup

### Requirement: Output language selection
The system SHALL answer in the language given by `lang`; when absent, it SHALL answer in Finnish if the query is Finnish and in English otherwise.

#### Scenario: Default for Finnish input
- **WHEN** `lang` is absent and the query is in Finnish
- **THEN** the response is in Finnish

#### Scenario: Default for other input
- **WHEN** `lang` is absent and the query is in a language other than Finnish
- **THEN** the response is in English

### Requirement: Input validation
The system SHALL return status 400 with a plain-text message when `q` is missing, empty, or longer than 300 characters, or when `lang` or `mode` has an unsupported value.

#### Scenario: Missing query
- **WHEN** a client sends GET / without `q`
- **THEN** the response is 400 with a plain-text explanation
