# Spec Delta

## Purpose

Returns the exact text of a requested Bible passage, tolerating typos and informal notation in the reference, in the requested output language.

## ADDED Requirements

### Requirement: Resolve loosely written references
The system SHALL resolve a verse reference written with misspelled or abbreviated book names, or spelled-out chapter, verse and range words, to a single canonical book, chapter and verse range.

#### Scenario: Misspelled book with word range
- **WHEN** the query is "Jhon 11:25 to 26"
- **THEN** the system resolves it to John 11:25-26

#### Scenario: Spelled-out chapter and verses
- **WHEN** the query is "Jhn chapter11 verses 25 and 26"
- **THEN** the system resolves it to John 11:25-26

#### Scenario: Finnish book name
- **WHEN** the query is "Joh 11:25" or a Finnish book name such as "Johanneksen evankeliumi 11:25"
- **THEN** the system resolves it to John 11:25

### Requirement: Reject unresolvable references
The system SHALL NOT guess when no book matches within the allowed typo distance or the chapter or verse does not exist.

#### Scenario: Nonexistent verse
- **WHEN** the query is "John 99:1"
- **THEN** the system returns a short plain-text error stating the reference was not found

### Requirement: Return text in the requested language
The system SHALL return the verse text from the Finnish translation when Finnish output is requested, and from the English translation otherwise, using the same book, chapter and verse numbers.

#### Scenario: Finnish output
- **WHEN** the reference John 11:25-26 is requested with Finnish output
- **THEN** the response contains the Finnish text of John 11:25-26 and the reference label

#### Scenario: Range response
- **WHEN** a range is requested
- **THEN** all verses in the range are returned in order, labelled with the full range
