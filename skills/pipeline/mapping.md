---
name: comply-mapping
description: Crosswalk frameworks and perform parameter harmonization. Parent Policy always sets the floor.
user-invocable: false
---

# Mapping — Delta Analysis & Parameter Harmonization

Compare parameters across Guidance Catalogs, the parent Policy, and Control Catalogs. Surface where the organization's values match or differ from each framework. When values differ, interpret which is stricter based on domain context.

**Core invariant:** The parent Policy always sets the floor. Never resolve below it without explicit user acknowledgement.

## Prerequisites

Requires `.complytime/scoping.yaml` from the scoping stage.

## Key Concepts

### Mandated vs. Under Evaluation

The parent Policy's `imports.guidance` determines which frameworks are binding:

- **Mandated:** Guidance Catalogs imported by the parent Policy. Shortfalls MUST be addressed.
- **Under evaluation:** Guidance Catalogs loaded in MCP but NOT imported. Informational only.

### Verdict Types

| Verdict | Meaning | Action |
|---------|---------|--------|
| `aligned` | Values match | No action |
| `mismatch` | Values differ | Interpret which is stricter based on domain context (e.g., TLS 1.3 > 1.2, shorter timeout = stricter). Present both values and your assessment to the user for confirmation. |
| `org_binds_generic` | Org provides concrete value for generic language | Document |
| `not_covered` | Parameter in one source only | Flag |

## Process

### Step 1: Read Scoping Artifacts

Read `.complytime/scoping.yaml`.

### Step 2: Identify Parent Policy

```
ListMcpResourcesTool(server="complypack")
```

If a Policy exists, read it. **If no parent Policy exists:** extract minimums from the target framework's Control Catalog requirements. The framework becomes the floor.

### Step 3: Classify Guidance Frameworks

If a parent Policy exists, check `imports.guidance`. If not, all Guidance Catalogs are **under evaluation** unless the user designates one as the target.

### Step 4: Load Mapping Documents

Read any Mapping Documents recorded in scoping:

```
ReadMcpResourceTool(server="complypack", uri="complypack://mapping/<id>")
```

Use them to resolve framework crosswalks.

### Step 5: Run Delta Analysis

```
CallMcpToolTool(server="complypack", tool="analyze_parameter_delta", arguments={"policyName": "<name>"})
```

### Step 6: Present Results

**Mandated:** Show each non-aligned verdict with values and recommended action.

**Under evaluation:** Same verdicts, framed as "if you pursue this certification..."

### Step 7: Resolve Decisions

Walk the user through `mismatch` items. For each, interpret which value is stricter given the parameter's domain, present your assessment, and let the user decide.

### Step 8: Write Output

Write `.complytime/delta-report.yaml`:

```yaml
version: "1"
created: YYYY-MM-DD
sources:
  parent_policy: "<policy-id>"
  guidance:
    mandated:
      - id: "<id>"
        status: imported_by_parent
    under_evaluation:
      - id: "<id>"
        status: loaded_not_imported
  catalogs: [<ids>]

parameters:
  - id: "<label>"
    requirement_id: "<req-id>"
    layers:
      framework:
        source: "<id>"
        value: "<value>"
        specificity: <concrete|generic|none>
      org_policy:
        source: "<id>"
        value: "<value>"
        specificity: <concrete|generic|none>
      tech_baseline:
        source: "<id>"
        value: "<value>"
        specificity: <concrete|generic|none>
    verdict: <verdict>
    resolved_value: "<value>"
    resolution: <resolution>
    rationale: "<why>"

summary:
  total: <N>
  aligned: <N>
  mismatch: <N>
  org_binds_generic: <N>
  not_covered: <N>
```

## MCP Resources and Tools

- `complypack://catalog/*` — Control Catalogs, Guidance Catalogs, Policies
- `complypack://mapping/*` — Mapping Documents
- `analyze_parameter_delta` — deterministic parameter crosswalk

## Red Flags

- [ ] Did you classify guidance as mandated vs. under evaluation?
- [ ] Did the user decide every `mismatch`?
- [ ] Did you ensure no resolution goes below the parent Policy floor?
- [ ] Did every value come from MCP?
