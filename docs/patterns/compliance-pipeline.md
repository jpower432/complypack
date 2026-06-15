# The Compliance Pipeline Pattern

The compliance pipeline pattern leverages Gemara to create a repeatable pipeline for compliance workflows. The flow is constant — profile, scope, map, commit, evaluate. The inputs change per persona: different drivers, catalogs, governance baselines, and scoping methodologies feed the same pipeline to produce applicability statements and fine-grained assessment logic.

## Pattern Card Fields

Each persona instantiates the pattern with these fields:

| Field             | Role                                                                         | Category      |
|:------------------|:-----------------------------------------------------------------------------|:--------------|
| **Persona**       | Who you are                                                                  | Identity      |
| **Driver**        | External motivation — why you act                                            | Input         |
| **Catalog**       | The standard (Guidance) you assess against                                   | Input         |
| **Governance**    | Your baseline — parent Policy, technology Control Catalogs                   | Input         |
| **Scoping Model** | Methodology for determining applicability and Mapping Document               | Input/Process |
| **Applicability** | The child Policy with `adherence` populated to express evidence requirements | Output        |
| **Evaluation**    | Fine-grained assessment logic (e.g. Rego)                                    | Output        |

**Inputs** are what you bring. **Process** is how applicability is determined. **Outputs** are artifacts the pipeline produces for different stakeholders.

Pattern cards are defined as YAML files validated against [`schema.cue`](schema.cue). Each card is a machine-readable input that skills and demo tooling consume to generate demo scripts, testdata, and recordings.
