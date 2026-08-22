# Multi-Locale Tool Params Specification

## Purpose

MCP tools and prompts SHALL advertise every discovered locale as a valid choice, so new book locales are usable without code changes. Existing defaults remain unchanged.

## Requirements

### Requirement: Dynamic locale enum on tool params

Every tool parameter that accepts a locale MUST use an Enum of the discovered locales (built from GetAvailableLocales at registration time). The enum MUST NOT be hardcoded to `es`/`en`.

Affected tools: list_books, list_chapters, read_chapter, search, build_semantic_index (locale param).

#### Scenario: Tools advertise all discovered locales

- GIVEN discovery returns 5 locales and the server registers its tools
- WHEN a client inspects the tool input schemas
- THEN each locale param's enum contains all 5 locales
- AND no locale present on disk is missing from any enum

#### Scenario: Handler accepts non-whitelisted locale

- GIVEN `harness` is a discovered locale not in the old es/en whitelist
- WHEN list_chapters is called with locale="harness"
- THEN it returns that locale's chapters without an unsupported-locale error

#### Scenario: Unknown locale rejected

- GIVEN the server is running
- WHEN a tool is called with locale "fr" (not discovered)
- THEN the handler returns a clear error naming valid locales

### Requirement: Defaults unchanged

The default locale for all tool params and prompts SHALL remain "es". Existing es/en behavior MUST NOT change.

#### Scenario: Default still es

- GIVEN tools are registered with dynamic enums
- WHEN list_chapters is called with no explicit locale
- THEN chapters come from the "es" locale

### Requirement: compare_patterns prompt uses dynamic locales

The compare_patterns prompt MUST NOT hardcode "es" as a selectable/compared locale; its locale inputs SHALL use the discovered set.

#### Scenario: Prompt offers all locales

- GIVEN 5 locales are discovered
- WHEN the compare_patterns prompt schema is listed
- THEN its locale params enumerate all 5 discovered locales
