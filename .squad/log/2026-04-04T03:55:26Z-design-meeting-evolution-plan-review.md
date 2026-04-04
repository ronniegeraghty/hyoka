# Design Meeting — Evolution Plan Review

**Date:** 2026-04-04  
**Time:** 03:55:26 UTC  
**Participants:** Morpheus 🕶️ (facilitator), Neo 💊, Tank 📡, Trinity 🖤, Switch 🤍, Oracle 🔮  
**Requested by:** Ronnie Geraghty  
**Status:** Complete  

## Meeting Summary

Pre-Phase 0 design review converged five independent domain reviews on a clear picture: the evolution plan is structurally sound but under-specified in three critical areas — the `GraderResult` type design, the boundary between typed fields and the `Properties` map, and the testing investment required for safe migration.

Every reviewer independently identified dependencies and ordering constraints that the flat task list obscures, and several found tasks that are missing entirely (reviewer system message, report schema versioning, template extraction, security sanitization).

**Consensus:** Phase 0 must be strengthened before Phase 1 begins. CI with race detection, pidfile tests, and serve handler security fix should all land before model-change work starts.

## Key Decisions

- 20 D-AUTO decisions made (coordinator authority)
- 4 escalations to Ronnie for approval
- 10 consensus points across all reviews
- Security issue identified: serve handler path traversal
- Testing investment in Phase 1 underestimated (~44 tests break, not "mostly mechanical")

## Consensus Points

1. **CI from day one** — `go build`, `go vet`, `go test -race` on every PR
2. **Convenience getters essential** — `p.Language()` instead of `p.Properties["language"]`
3. **System prompt removal phased** — bias rules → guidance rules → path rules
4. **File grader built first** — validates `GraderResult` type design
5. **Schema versioning mandatory** — avoids breaking old reports on `rerender`
6. **Template extraction precedes display work** — extract to `.gohtml` before Phase 3.2a
7. **Testing investment significantly underestimated** — ~8 prompt, ~6 filter, ~12 config, 6 criteria, 12 main tests need updating
8. **Documentation underrepresented** — 1 task in 72-task plan; recommend 12 distributed
9. **Config big-bang migration safe** — single PR, 6 of 8 configs legacy, ~17 call sites
10. **React SPA for Phase 4 dashboard** — interactive comparison in React, static reports in Go templates

---
**Recorded by:** Scribe  
**Next Steps:** Implement D-AUTO decisions; Ronnie reviews escalations (D-AR1-4)
