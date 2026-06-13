# ADR 015: Comply Pipeline as Plugin Skills

**Status:** Proposed

**Date:** 2026-06-11

**Context:**

ComplyPack's MCP server provides compliance data (controls, parameters, resolved policies) but has no opinion on workflow. Users need guided, multi-stage audit preparation: scoping which controls apply, mapping parameter deltas across frameworks, and producing an adherence plan (the applicability statement).

Three approaches were considered:

1. **Hardcoded CLI workflow** — `complypack comply --stage scoping`. Rigid; can't adapt to user context or partial completion.
2. **Autonomous agent loop** — a single agent prompt that runs all stages. Prone to hallucinating control IDs and parameter values when context windows grow large.
3. **Plugin skills with MCP grounding** — decompose the pipeline into discrete skills (`/comply:pipeline`, `/comply:pack`, `/comply:setup`), each reading from MCP resources at every step.

Option 2 was rejected on safety grounds: an AI agent hallucinated a control ID in a real evidence registry during testing. The core safety property is that every stage must read control IDs, requirement IDs, and parameter values from MCP resources — never from model memory.

Option 1 was rejected because audit workflows are inherently conversational: auditors need to review scoping decisions, adjust parameter bindings, and approve the applicability statement before it's finalized.

Option 3 was chosen. Skills provide structured guidance while keeping the human in the loop. MCP grounding prevents hallucination. The pipeline router checks `.complytime/` artifact state to resume from the correct stage.

**Decision:**

Implement the comply pipeline as plugin skills:

- **`/comply:pipeline`** — router that inspects `.complytime/` directory state and dispatches to the correct stage (scoping → mapping → adherence)
- **`/comply:pack`** — generates assessment logic after pipeline completion
- **`/comply:setup`** — configures `.mcp.json` for the user's environment

Each stage reads exclusively from MCP resources. The pipeline produces Gemara Policy artifacts — functionally equivalent to an ISO 27001 Statement of Applicability or a NIST System Security Plan.

**Consequences:**

- Users interact conversationally with each stage rather than running a batch process — audit decisions are reviewed before being recorded
- The pipeline is stateless across sessions: `.complytime/` artifacts are the checkpoint, not conversation history
- Adding new stages (e.g., evidence collection, continuous monitoring) means adding new skill files and updating the router
- The plugin registers with Claude Code, Cursor, and Gemini via their respective plugin manifests — same skills, three runtimes
- MCP grounding is a hard constraint: if the MCP server is unreachable, the pipeline fails rather than proceeding with stale or generated data
