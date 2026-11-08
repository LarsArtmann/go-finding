# docs/status

Point-in-time session status reports (what was done, what broke, what is
next, at the moment of writing). They age immediately: each claim is
resolved inline via ~~strikethrough~~ annotations with a disposition
header, and fully-resolved reports move to `archived/`. Open work does
NOT live here — it is harvested into the root `TODO_LIST.md` and
`ROADMAP.md`. Read those for current state; read these for history.

## Archive manifest — 2026-11-08 docs-health AUDIT sweep

Filename → classification → deciding reason (all 8 files annotated inline,
gates `check-rows --strict`/`bare-rows` + `grep -L '~~'` green BEFORE move):

- `2026-09-23_14-50_docs-health-sweep...` → ARCHIVE — every §a-§g item resolved; forward tails harvested to TODO_LIST/ROADMAP
- `2026-09-23_15-58_pareto-plan-pushed...` → ARCHIVE — T1-T27 executed same evening; §f rows struck/routed
- `2026-09-23_16-12_closing-loop...` → ARCHIVE — closing-loop segment superseded by the 18:13 round; all items dispositioned
- `2026-09-23_18-13_whole-list-execution...` → ARCHIVE — v1.13.0 completed + 26 plan tasks executed; tails routed
- `2026-09-23_22-07_session-close...` → ARCHIVE — session close; every §a-§g item resolved or routed
- `2026-09-28_01-22_dependabot-train...` → ARCHIVE — dependabot train resolved; on-sight fixes landed same-day (vendorHash comment, pre-push header)
- `2026-09-30_14-31_pr41-options-channel...` → ARCHIVE — #41 merged 2026-10-01, toolsdk shipped v1.14.0→v1.16.0; external tails routed to their repos
- (`docs/planning/`) `2026-09-23_15-01_SUPERB-pareto-plan...` → ARCHIVE — plan FULLY EXECUTED: all 27 T-rows + 101 micro-rows carry inline verdicts; annotation log is the execution record

Left active (harvest source for the next session): the six 2026-10-01/10-11/10-12 train reports — fully resolved but kept as the freshest reports per the "most recent 1-3" harvest rule; annotate+archive them in a later sweep.
