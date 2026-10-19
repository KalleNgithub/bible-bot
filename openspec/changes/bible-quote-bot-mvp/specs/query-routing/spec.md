# Spec Delta

## Purpose

Chooses between direct verse lookup and topic search when the caller does not state which is wanted.

## ADDED Requirements

### Requirement: Explicit mode wins
The system SHALL use the mode named by the caller (`verse` or `topic`) without evaluating the query.

#### Scenario: Forced topic search
- **WHEN** mode is `topic` and the query is "John 11:25"
- **THEN** the query is handled as a topic search

### Requirement: Automatic mode selection
When no mode is given, the system SHALL treat a query as a verse reference if it resolves to a valid reference, and as a topic query otherwise.

#### Scenario: Reference query
- **WHEN** the query is "Jhon 11:25 to 26" and no mode is given
- **THEN** it is handled as a verse lookup

#### Scenario: Topic query
- **WHEN** the query is "what happens to people who believe in jesus" and no mode is given
- **THEN** it is handled as a topic search

#### Scenario: Reference in another language
- **WHEN** the query is a reference written with Finnish book names and no mode is given
- **THEN** it is handled as a verse lookup
