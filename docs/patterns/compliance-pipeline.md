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

## Foundation-Backed Open Source Project

| Field             | Value                                                                                                                        |
|:------------------|:-----------------------------------------------------------------------------------------------------------------------------|
| **Persona**       | Foundation-backed open source project                                                                                        |
| **Driver**        | Foundation CRA steward obligations and downstream consumer demand                                                            |
| **Catalog**       | OSPS Baseline                                                                                                                |
| **Governance**    | Ecosystem governance — foundation's security policy and project oversight combined with the project's own governance process |
| **Scoping Model** | Maturity level (1/2/3)                                                                                                       |
| **Applicability** | Project's applicability statement at its assessed maturity level                                                             |
| **Evaluation**    | Fine-grained assessment logic (e.g. Rego)                                                                                    |

## Manufacturer: CRA Open Source Consumption

| Field | Value |
|:------|:------|
| **Persona** | Manufacturer performing due diligence on open source dependencies |
| **Driver** | Cyber Resilience Act |
| **Catalog** | CrabFOMA / CrabFOSC |
| **Governance** | Organization's product security policy and internal technology baselines |
| **Scoping Model** | Component risk tier (low / high) |
| **Applicability** | Product-level applicability statement for open source consumption |
| **Evaluation** | Fine-grained assessment logic (e.g. Rego) |

## Manufacturer: AI Governance

| Field | Value |
|:------|:------|
| **Persona** | Organization building or deploying AI systems |
| **Driver** | ISO 42001 |
| **Catalog** | AIGS Baseline |
| **Governance** | Organization's AI governance policy and internal technology baselines |
| **Scoping Model** | AI system architecture and risk classification |
| **Applicability** | AI system-level applicability statement |
| **Evaluation** | Fine-grained assessment logic (e.g. Rego) |

## Manufacturer: Financial Services Cloud

| Field | Value |
|:------|:------|
| **Persona** | Financial services organization deploying to public cloud |
| **Driver** | Financial regulation |
| **Catalog** | FINOS Common Cloud Controls |
| **Governance** | Organization's cloud security policy and internal technology baselines |
| **Scoping Model** | Cloud service taxonomy |
| **Applicability** | Cloud deployment applicability statement |
| **Evaluation** | Fine-grained assessment logic (e.g. Rego) |
