# Design: multi-locale-support

## Technical Approach

Discover locales once at server startup from the book filesystem, store the sorted list as a package-level value, and use it to build every tool enum, prompt, resource, and the "all" indexing target — before `ServeStdio`. No per-request scanning; no runtime tool re-registration needed because all dynamic construction happens before the server starts serving (mcp-go v0.43.2 allows adding tools/resources any time before sessions list them — Context7: `AddTool`, `AddResourceTemplate`, session-tools docs confirm thread-safe add and that clients get notifications on later adds, which we don't need).

## Architecture Decisions

| # | Decision | Options | Tradeoff | Choice |
|---|----------|---------|----------|--------|
| 1 | Where discovery runs | a) main.go startup b) Parser init | Both fine; main.go keeps parser dumb and gives one explicit call site with error handling | **a) main.go**, calling fixed `GetAvailableLocales()` |
| 2 | Enum mechanism | a) Rebuild/re-register tools at runtime b) Static `mcp.Enum` built at startup + handler validation | (b) suffices: locales are fixed at process start; (a) adds complexity for zero benefit | **b** — build `localeEnum := append([]string{}, discovered...)` and pass `mcp.Enum(localeEnum...)` into each `WithString("locale", ...)` |
| 3 | Resource registration | a) One static resource per locale b) Single `AddResourceTemplate("book://index/{locale}", ...)` | (b) is one line, handles any discovered locale, no re-registration on restart with new dirs (Context7 README template example). Template handler validates against discovered set | **b** — replace the two hardcoded `AddResource` calls |
| 4 | "all" semantics | a) Keep literal "all" resolved at index time b) Expand at startup | (a): expand `"all"` → discovered slice inside `handleBuildSemanticIndex` loop over stored locales | **a** — replaces hardcoded `[]string{"es","en"}` (~main.go:605) |
| 5 | Unknown locale | Error vs silent fallback | Spec requires clear error | Shared `validateLocale(locale string) error` helper used by handlers |
| 6 | Discovery whitelist | Keep en/es filter vs accept any dir containing .mdx | Spec dynamic-locale-discovery mandates .mdx-filtered discovery, no whitelist | Remove whitelist (parser.go:373); skip hidden dirs; sort |

## Data Flow

```
startup (main.go)
  bookPath ──→ GetAvailableLocales()          [parser.go — one os.ReadDir]
                     │
                     ▼
        availableLocales []string (sorted pkg var)
                     │
     ┌───────────────┼─────────────────┬──────────────┐
     ▼               ▼                 ▼              ▼
 tool enums      prompts args    AddResourceTemplate   "all" expansion
 mcp.Enum(...)   descriptions    book://index/{locale} → discovered slice
     │               │                 │              │
     └─────── ServeStdio(s) ◄──────────┴──────────────┘
                     │ per request
        handler → validateLocale() → parser/embeddings
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/book/parser.go` | Modify | `GetAvailableLocales()` (line 365): drop en/es whitelist (373), skip dot-dirs, sort result. Now called by main.go — dead code becomes live |
| `cmd/server/main.go` | Modify | Startup: call discovery, store `availableLocales`. Replace 5× hardcoded `DefaultString("es")` param blocks with shared `localeParam(default)` helper using `mcp.Enum`. Handlers call `validateLocale` instead of bare `GetString("locale","es")`. Replace two `AddResource` calls with one `AddResourceTemplate`. `compare_patterns` gains `locale` arg (drop hardcoded `"es"` ~423). `"all"` expands to discovered list (~605) |
| `internal/book/parser_test.go` | Create | TDD baseline (see Testing Strategy) |
| `cmd/server/main_test.go` | Create | Tests for `validateLocale`, enum construction, locale extraction from resource URIs, "all" expansion |

## Interfaces / Contracts

```go
// parser.go (behavior change, signature unchanged)
func (p *Parser) GetAvailableLocales() ([]string, error)
// returns sorted dir names under bookPath that contain ≥1 *.mdx file,
// skipping names starting with "."; empty bookPath → empty slice, nil error

// main.go new helpers
func validateLocale(locale string) error // nil if in availableLocales or == "all" (indexing only)
func localeParam(def string) mcp.ToolOption // WithString("locale", Enum(discovered...), DefaultString(def))
```

## Testing Strategy (STRICT TDD — RED first)

| Layer | What | Order |
|-------|------|-------|
| Unit | `parser_test.go`: discovery finds dirs with .mdx; skips dirs without .mdx; skips `.hidden`; sorts output; empty path → empty,nil; whitelist removed | 1st (RED before touching parser.go) |
| Unit | `main_test.go`: `validateLocale` accepts known / rejects unknown w/ message listing valid locales; `localeParam` produces Enum schema containing discovered values; resource URI locale extraction (`book://index/fr`) | 2nd |
| Integration | none yet per config.yaml (`integration: false`) — out of scope this change | — |

Test fixtures: temp dirs via `t.TempDir()` with stub .mdx files — no dependency on real book repo.

## Threat Matrix

N/A — no routing, shell, subprocess, VCS/PR automation, executable-file classification, or process-integration boundary. Locale strings flow only into `filepath.Join(bookPath, locale)`; `validateLocale` restricts them to directory names observed by `os.ReadDir`, so no traversal surface is added.

## Migration / Rollout

No data migration. Rollback = revert commit; behavior identical when book repo has only es/en. Risk: semantic-index cost grows linearly with locales (accepted Medium risk per proposal).

## Open Questions

- [ ] None blocking.
