# Exploration: multi-locale-support

## Current State

- `internal/book/parser.go:365-379` — `GetAvailableLocales()` reads `p.bookPath` dir entries but whitelists `en`/`es` at line 373. **Verified: this function has ZERO callers** — it is dead code today.
- All parser methods (`ListChapters`, `GetChapter`, `GetSection`, `Search`, `GetBookIndex`) already take `locale` as a plain string param and are fully generic — they just join `bookPath/locale` and read `.mdx` files.
- Book repo (`.../data/book/gentleman-programming-book/`) contains 5 locale dirs: `ai-agent`, `en`, `es`, `harness`, `secret-knowledge` (secret-knowledge has 15 ready MDX chapters).
- The real locale hardcoding lives in `cmd/server/main.go`, not the parser:
  - Lines 57-58, 76-77, 92-93, 104-105, 124-125: five tools with static description `"Language locale: 'es' for Spanish, 'en' for English"` — no Enum constraint.
  - Line 139-140: `build_semantic_index` description `'es', 'en', or 'all'`.
  - Lines 160-178: two hardcoded MCP resources `book://index/es` and `book://index/en`.
  - Line 605-607: `handleBuildSemanticIndex` hardcodes `locales = []string{"es", "en"}` for `"all"`.
  - Prompt handlers (`explain_concept`, `summarize_chapter`) accept any locale string; `compare_patterns` hardcodes `"es"` (lines 423-424).
  - `handleBookIndexResource` parses locale from URI suffix (only es/en registered).
- No tests exist (`*_test.go`: zero files) — no regression baseline.

## Affected Areas

- `internal/book/parser.go` — `GetAvailableLocales()`: remove whitelist; discover dirs dynamically.
- `cmd/server/main.go` — all six tool registrations (add dynamic `mcp.Enum(locs...)`), `handleBuildSemanticIndex` ("all" → discovered locales), resource registration loop, prompt descriptions. Handlers themselves need no change (they pass locale through).
- `internal/embeddings/embeddings.go` — no change needed; `Search` filters by chunk.Locale generically.
- New: first test file(s) for parser locale discovery (strict TDD baseline).

## Approaches

1. **Option A — Remove whitelist, discover dirs dynamically**
   - `GetAvailableLocales()` returns all subdirectories of bookPath (optionally filtered to dirs containing `.mdx` files to exclude non-locale dirs like assets). main.go calls it once at startup, builds `mcp.Enum(locs...)` per tool, iterates locales for resources and "all" indexing.
   - Pros: single source of truth; new locale dirs appear with zero code changes; smallest diff (~40 lines total); uses existing dead function as intended.
   - Cons: a stray non-chapter directory becomes an advertised locale (mitigated by filtering on "contains .mdx").
   - Effort: Low

2. **Option B — Env var override (GSTACK-style, e.g. `BOOK_LOCALES="en,es,secret-knowledge"`)**
   - Pros: explicit opt-in per deployment; can hide drafts.
   - Cons: second source of truth that drifts from disk; config burden for every new locale; redundant when the filesystem already IS the truth.
   - Effort: Low-Medium

3. **Option C — Config file (e.g. `locales.yaml` in repo)**
   - Pros: can carry display names/ordering metadata.
   - Cons: over-engineering — nothing needs metadata today; file must be kept in sync manually.
   - Effort: Medium

## Recommendation

**Option A.** The filesystem is already the source of truth; the whitelist is the only thing blocking multi-locale. Filter discovery to directories containing at least one `.mdx` file so non-locale dirs are excluded naturally. Wire main.go's enum/description/resources/"all"-indexing off the discovered list. Skip env override until someone actually needs to hide a locale dir.

## Risks

- Non-locale directories (e.g. shared assets under data/book) would surface as locales if unfiltered → filter by presence of `.mdx`.
- `build_semantic_index` with `"all"` will index 5 locales instead of 2 → longer indexing time / more embedding cost; acceptable, user opted in.
- `secret-knowledge` chapters become readable via the public MCP server once discovered — if that's meant to be gated, gating is a separate concern (out of scope note for proposal).
- Zero existing tests: strict-TDD apply phase must create the parser test baseline first (RED before removing whitelist).
- Default locale remains `"es"` everywhere; unchanged behavior for existing clients except expanded enum options.

## Ready for Proposal

Yes — recommend proposal scope: (1) dynamic `GetAvailableLocales()` with .mdx filter, (2) main.go startup wiring for enums/resources/all-indexing, (3) parser test baseline.
