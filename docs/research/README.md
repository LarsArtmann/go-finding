# docs/research/

**What lives here:** one-off investigations that answer a question and
then stay done — dependency assessments (go-error-family v0.4.0,
samber-oops v1.22.0), performance analyses, split-brain hunts, consumer
audits, and pro/contra writeups that fed an architecture decision. The
`archived/` subdirectory holds superseded rounds; nothing there is
current advice.

**What does NOT live here:** decisions (the ADR log owns those —
research documents EVIDENCE a decision, it does not make one), features
(`FEATURES.md`), feedback received from consumers
(`docs/feedback/README.md`), and anything still actionable
(`TODO_LIST.md`).

**Lifecycle:** a research document is finished when its question is
answered. Do not keep editing it afterwards — cite it from the ADR or
status report that used it instead. Claims about EXTERNAL tools (library
versions, upstream behavior) carry a verify date and go stale; re-verify
against the live source before building on them
(`~/.config/crush/skills/verify-external-claims` discipline).
