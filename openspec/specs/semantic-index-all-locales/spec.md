# Semantic Index All Locales Specification

## Purpose

Semantic indexing SHALL cover every discovered locale: `build_semantic_index` with locale="all" indexes all of them, and per-locale index resources expose each one.

## Requirements

### Requirement: "all" indexes every discovered locale

build_semantic_index MUST accept locale="all", meaning every locale returned by discovery at startup.

#### Scenario: Index all locales

- GIVEN 5 locales are discovered and embeddings are configured
- WHEN build_semantic_index is called with locale="all"
- THEN content from all 5 locales is embedded and stored
- AND search can return results in any of them

#### Scenario: Single locale indexing unchanged

- GIVEN the server is running
- WHEN build_semantic_index is called with locale="es"
- THEN only the "es" locale is indexed

### Requirement: Per-locale index resources registered

For every discovered locale, the server MUST register a resource at `book://index/<locale>` exposing that locale's semantic index metadata.

#### Scenario: Resource exists per locale

- GIVEN discovery returns en, es, ai-agent, harness, secret-knowledge
- WHEN resources/list is called
- THEN book://index/en, book://index/es, book://index/ai-agent, book://index/harness, and book://index/secret-knowledge are all present

#### Scenario: No resource for unknown locale

- GIVEN only discovered locales get resources
- WHEN reading book://index/fr
- THEN the server returns a resource-not-found error

### Requirement: New locale appears automatically

Adding a new locale dir to the repo SHOULD make it indexable via "all" and expose its index resource after restart, with no code change.

#### Scenario: Restart picks up new locale

- GIVEN a new locale dir `pt` with `.mdx` files was added and the server restarted
- WHEN build_semantic_index runs with locale="all"
- THEN pt content is included and book://index/pt exists
