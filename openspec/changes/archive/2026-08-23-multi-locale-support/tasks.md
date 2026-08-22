# Tasks: multi-locale-support

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~120-150 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|------|------|-----------|----------------------|-----------------|-------------------|
| 1 | Dynamic locale discovery + server wiring | PR 1 | `go test ./internal/book/... ./cmd/server/...` | `go run ./cmd/server` against book repo; tools/list shows 5 locales | Revert single commit; behavior identical when only es/en exist |

## Phase 1: RED — Parser discovery tests (STRICT TDD)

- [x] 1.1 Create `internal/book/parser_test.go`: failing tests for `GetAvailableLocales` using `t.TempDir()` fixtures — GIVEN dirs `en,es,ai-agent,harness,secret-knowledge` each with ≥1 `.mdx`, THEN all 5 returned (spec: dynamic-locale-discovery "Discovery of all locale dirs")
- [x] 1.2 Add RED cases: `.git/` dir skipped even with .mdx inside ("Hidden directory excluded"); dir `images/` with only `.png` NOT a locale ("Directory without .mdx files"); nonexistent bookPath → empty slice or clear error, no panic ("Empty book path"); output sorted lexicographically ("Sorted output"). Verify: `go test ./internal/book/ -run TestGetAvailableLocales` fails
- [x] 1.3 REFACTOR gate: confirm parser coverage of new discovery branches (`go test ./internal/book/ -cover`) after GREEN in 2.x

## Phase 2: GREEN — Parser fix

- [x] 2.1 `internal/book/parser.go:365` `GetAvailableLocales`: drop whitelist at line 373; accept any non-hidden dir containing ≥1 `*.mdx`; skip names starting with "."; sort result; nonexistent path → empty slice, nil error. Verify: Phase 1 tests pass

## Phase 3: RED — Server wiring tests

- [x] 3.1 Create `cmd/server/main_test.go`: RED test `validateLocale("fr")` errors listing valid locales; `validateLocale("harness")` and `"all"` pass (spec: multi-locale-tool-params "Unknown locale rejected")
- [x] 3.2 Add RED tests: `localeParam(default)` produces schema Enum containing all discovered locales; locale extraction from resource URI `book://index/fr` → `"fr"`; `"all"` expands to discovered list not hardcoded `[]string{"es","en"}`. Verify: `go test ./cmd/server/` fails

## Phase 4: GREEN — Server wiring

- [x] 4.1 `cmd/server/main.go`: startup calls `parser.GetAvailableLocales()` once pre-`ServeStdio`, stores sorted pkg var `availableLocales`
- [x] 4.2 Add helpers: `localeParam(def string) mcp.ToolOption` = `WithString("locale", mcp.Enum(availableLocales...), DefaultString(def))`; `validateLocale(locale string) error` (nil if in availableLocales or =="all"); replace the 5× hardcoded `DefaultString("es")` param blocks
- [x] 4.3 Handlers call `validateLocale` on extracted locale instead of bare defaults; replace 2× `AddResource` with one `AddResourceTemplate("book://index/{locale}")` whose handler extracts + validates locale from URI; `compare_patterns` gains `locale` arg (drop hardcoded `"es"` ~line 423); `"all"` expansion uses `availableLocales` (~605). Verify: `go test ./cmd/server/` passes

## Phase 5: REFACTOR + Verify

- [x] 5.1 Full suite green: `go test ./...`; `gofmt -l .` and `go vet ./...` clean
- [x] 5.2 Confirm spec scenarios mapped: dynamic-locale-discovery (all 6), multi-locale-tool-params (defaults still "es", compare_patterns enums), semantic-index-all-locales (per-locale resources via template)
