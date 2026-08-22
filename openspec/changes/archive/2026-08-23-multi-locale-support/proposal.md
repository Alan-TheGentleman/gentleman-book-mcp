# Proposal: Multi-Locale Support

## Intent

The MCP server hardcodes `es`/`en` while the book repo now has 5 locale dirs (`en`, `es`, `ai-agent`, `harness`, `secret-knowledge`). `GetAvailableLocales()` exists but is dead code with its own en/es whitelist. New locales require code edits today; they should appear automatically from the filesystem.

## Scope

### In Scope
- Make `internal/book/parser.go` `GetAvailableLocales()` discover locale dirs dynamically (filter: contains `.mdx`)
- Wire `cmd/server/main.go`: dynamic `mcp.Enum(locs...)` on the 6 tool locale params, resources per discovered locale, `"all"` semantic indexing from discovered list
- Fix `compare_patterns` prompt hardcoded `"es"`
- Parser test baseline (strict TDD: RED before whitelist removal)

### Out of Scope
- Env/config locale overrides (rejected in explore: second source of truth)
- Access gating for gated content (e.g. secret-knowledge visibility) — separate change
- Embeddings changes (already locale-generic)
- Default locale stays `"es"`

## Capabilities

> openspec/specs/ is empty — all capabilities are NEW.

### New Capabilities
- `dynamic-locale-discovery`: parser discovers locales from filesystem subdirs containing `.mdx`; no hardcoded whitelist
- `multi-locale-tool-params`: tool/prompt locale params advertise all discovered locales via Enum; defaults unchanged
- `semantic-index-all-locales`: `"all"` indexing and `book://index/<locale>` resources cover every discovered locale

### Modified Capabilities
- None

## Approach

Option A (explore recommendation): filesystem is single source of truth. Remove whitelist in `GetAvailableLocales()`, filter dirs by `.mdx` presence. Call it once at startup in main.go; build enums, resources, and the `"all"` index list from that slice. Handlers pass locale through untouched (~40-line diff). Strict TDD: write parser discovery tests first (`go test ./...` RED), then implement.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/book/parser.go` | Modified | Whitelist removed; .mdx-filtered dir scan |
| `cmd/server/main.go` | Modified | Dynamic Enum params, per-locale resources, "all" list |
| `internal/book/parser_test.go` | New | TDD baseline for locale discovery |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Non-locale dirs surface as locales | Low | Filter on `.mdx` presence |
| `"all"` indexing cost grows with locales | Medium | Accepted; user opted in |
| Gated content becomes publicly readable | Medium | Explicit non-goal; flag to owner |

## Rollback Plan

Single commit revert restores the whitelist and static main.go wiring; no data migrations or persisted state changes. Semantic index rebuilt with `"all"` can be re-scoped by reverting and rebuilding for es/en.

## Dependencies

- Book repo path present at configured bookPath
- mcp-go v0.43.2 `mcp.Enum(...)` variadic support (already used elsewhere)

## Success Criteria

- [ ] `go test ./...` passes with new parser discovery tests (TDD RED→GREEN recorded)
- [ ] All 5 locale dirs discoverable; tools advertise them via Enum
- [ ] `book://index/<locale>` resources exist for every discovered locale
- [ ] `build_semantic_index` with `"all"` indexes all discovered locales
- [ ] Existing es/en behavior unchanged (defaults still "es")
