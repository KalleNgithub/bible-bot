# Spec Delta

## Purpose

Prevents returned quotes from starting or ending mid-sentence so they keep their meaning, while keeping responses short.

## ADDED Requirements

### Requirement: Complete unfinished sentences
The system SHALL extend a selected verse range backward while its first verse begins as a continuation of the previous verse, and forward while its last verse does not end a sentence.

#### Scenario: Verse ends mid-sentence
- **WHEN** the selected verse ends with a comma, colon or no terminal punctuation
- **THEN** the following verse is included

#### Scenario: Verse begins as continuation
- **WHEN** the selected verse begins with a lowercase letter or a word that opens a dependent clause ("that", "who", "which", "because", "or", "nor") continuing the previous verse
- **THEN** the previous verse is included

#### Scenario: Sentence-opening conjunction stands alone
- **WHEN** the selected verse begins with a capitalised "And", "But" or "For" and ends with terminal punctuation
- **THEN** only that verse is returned

#### Scenario: Self-contained verse
- **WHEN** the selected verse begins a sentence and ends with terminal punctuation
- **THEN** only that verse is returned

### Requirement: Cap on length
The system SHALL NOT return more than a fixed maximum number of verses, regardless of how far the sentence extends.

#### Scenario: Long sentence chain
- **WHEN** expansion would exceed the maximum verse count
- **THEN** expansion stops at the maximum
- **AND** the result still contains the originally selected verse

### Requirement: Expansion stays within the chapter
The system SHALL NOT extend a range past the first or last verse of a chapter.

#### Scenario: First verse of a chapter
- **WHEN** the selected verse is verse 1 and is a continuation
- **THEN** no previous verse is added
