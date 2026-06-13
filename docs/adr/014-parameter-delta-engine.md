# ADR 014: Parameter Delta Comparison Engine

**Status:** Proposed

**Date:** 2026-06-10

**Context:**

Gemara Policies bind parameters at the org level (e.g., "session timeout = 15 minutes"), while Guidance Catalogs define framework-level expectations (e.g., "session timeout per organizational policy"). When multiple frameworks apply — say a regulatory framework and an organizational baseline — a single parameter can have values at three layers: framework guidance, org policy, and tech baseline.

Auditors need to see where the org's parameter values align with or differ from framework expectations. Doing this manually across dozens of parameters and multiple catalogs is error-prone and time-consuming.

Three approaches were considered:

1. **String equality only** — flag mismatches without any classification. Simple but loses the distinction between "framework defers to org" and "values genuinely differ."
2. **Full directional comparison** — classify whether the org exceeds or falls short of the framework by comparing values numerically. Unreliable: "higher is stricter" holds for TLS versions (1.3 > 1.2) but not for session timeouts (15 < 30 means stricter). Algorithm comparisons (AES-256-GCM vs ChaCha20-Poly1305) aren't orderable at all.
3. **Layer-based crosswalk with mismatch detection** — classify each parameter's specificity (concrete, generic, none), detect whether values match, and produce a verdict per parameter. Leave directional interpretation (which value is stricter) to the LLM in the mapping stage, which has the domain context to judge.

Option 3 was chosen. The deterministic engine handles what it's good at — detecting alignment, mismatches, generic-to-concrete bindings, and coverage gaps. The LLM handles what it's good at — interpreting whether "1.3 vs 1.2" means the org is stricter or more lenient for a given parameter.

**Decision:**

Implement a parameter delta comparison engine (`requirement.AnalyzeDelta`) that crosswalks parameters across framework, org policy, and tech baseline layers. Each parameter receives a verdict:

- `aligned` — values match
- `mismatch` — values differ; the caller interprets directionality
- `org_binds_generic` — framework defers to org; org provides a concrete value
- `not_covered` — no framework value exists for comparison

Expose this as the `analyze_parameter_delta` MCP tool so the `/comply` pipeline's mapping stage can read delta reports directly from the server. The mapping stage LLM interprets each `mismatch` verdict in domain context and presents its assessment to the user for confirmation.

**Consequences:**

- The mapping stage of the comply pipeline consumes delta reports from MCP rather than computing them in-skill — no hallucinated parameter comparisons
- Directional interpretation ("which is stricter") is the LLM's responsibility, not the engine's — this correctly handles TLS versions, timeouts, algorithms, and any future parameter types without engine changes
- The `tech_baseline` layer is structurally present but not yet populated from Gemara data — `findGuidelineParameter` is a stub awaiting Guidance Catalog parameter extraction
- The engine is intentionally simple: no version parsing, no numeric comparison, no type detection — just string equality and specificity classification
