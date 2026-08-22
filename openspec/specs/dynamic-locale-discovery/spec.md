# Dynamic Locale Discovery Specification

## Purpose

Locale directories in the book repo are the single source of truth for which locales exist. The parser SHALL discover them from the filesystem at runtime instead of using a hardcoded whitelist.

## Requirements

### Requirement: Filesystem-based locale discovery

GetAvailableLocales MUST scan the configured bookPath and return every subdirectory that contains at least one `.mdx` file. It MUST NOT contain a hardcoded locale whitelist.

#### Scenario: Discovery of all locale dirs

- GIVEN a bookPath containing dirs `en`, `es`, `ai-agent`, `harness`, `secret-knowledge`, each with `.mdx` files
- WHEN GetAvailableLocales is called
- THEN it returns exactly those five locales
- AND none are filtered by any hardcoded list

#### Scenario: Sorted output

- GIVEN multiple discovered locale dirs
- WHEN GetAvailableLocales is called
- THEN results are returned sorted lexicographically

### Requirement: .mdx presence filter

A directory SHALL qualify as a locale only if it contains at least one file ending in `.mdx`. Files without content do not create locales.

#### Scenario: Directory without .mdx files

- GIVEN bookPath contains dir `images/` with only `.png` files
- WHEN GetAvailableLocales is called
- THEN `images` is NOT returned as a locale

### Requirement: Hidden and non-content dirs ignored

The discovery scan SHOULD skip hidden directories (e.g. names starting with `.`).

#### Scenario: Hidden directory excluded

- GIVEN bookPath contains `.git/` (with or without `.mdx` files)
- WHEN GetAvailableLocales is called
- THEN `.git` is not returned

### Requirement: Missing or empty bookPath handled gracefully

Discovery MUST handle an invalid or empty bookPath without crashing.

#### Scenario: Empty book path

- GIVEN bookPath points to a nonexistent or empty directory
- WHEN GetAvailableLocales is called
- THEN it returns an empty locale list (or a clear error), not a panic

#### Scenario: Single call at startup

- GIVEN the server starts successfully
- WHEN tools and resources are registered
- THEN discovery runs once and its result feeds enum params and resource registration
