# MCP Validation and Testing Tools Design

**Date:** 2026-06-02  
**Issue:** #12  
**Goal:** Add MCP tools for policy validation and testing to enable LLM-driven validate-test-repair loops

## Problem Statement

The current ComplyPack MCP server provides catalogs and schemas as resources, enabling LLM-assisted policy generation. However, there's no way for the LLM to validate generated policies or test them without writing files to disk and running external commands.

**Missing capability:** Inline validation and testing within the MCP tool interface.

**Desired workflow:**
1. LLM reads control from catalog (MCP resource) ✅
2. LLM reads platform schema (MCP resource) ✅
3. LLM generates policy ✅
4. **LLM validates policy (syntax, contract, lint)** ❌ (no tool)
5. **LLM sees errors, repairs policy** ❌ (no feedback loop)
6. **LLM generates test data, validates against schema** ❌ (no tool)
7. **LLM runs tests, sees failures, fixes policy** ❌ (no tool)

This design adds the missing tools to close the loop.

## Architecture

### High-Level Design

**Two MCP tools extending the existing server:**

1. **`validate_policy`** — Deterministic validation of policy syntax, schema contract, and linting
2. **`test_policy`** — Schema-validated test data execution against policy

**Integration:**
- Tools registered in `internal/mcp/server.go` during server initialization
- Tool handlers implemented in new file `internal/mcp/tools.go`
- Tools consume existing evaluator registry (`evaluator.DefaultRegistry()`)
- Platform schemas already loaded in MCP server's `ResourceStore`

**Design principles:**
- **Deterministic validation** — No LLM in the validation path, only in repair
- **LLM-driven repair** — Tools return structured errors, LLM fixes code
- **Schema-validated test data** — Test data validated against platform schema before policy execution
- **Granular tools** — Separate validation and testing for iterative workflows

### Data Flow

```
┌─────────────────────────────────────────────────────────┐
│                    LLM (Claude)                         │
└─────────────────────────────────────────────────────────┘
           │                            │
           │ validate_policy            │ test_policy
           ▼                            ▼
┌──────────────────────┐      ┌─────────────────────────┐
│ handleValidatePolicy │      │   handleTestPolicy       │
│                      │      │                          │
│ 1. Get evaluator     │      │ 1. Validate test data    │
│ 2. Validate syntax   │      │    against JSON Schema   │
│ 3. Load CUE schema   │      │ 2. Get evaluator         │
│ 4. Check contract    │      │ 3. Execute tests         │
│ 5. Run lint          │      │ 4. Return results        │
└──────────────────────┘      └─────────────────────────┘
           │                            │
           ▼                            ▼
┌─────────────────────────────────────────────────────────┐
│            evaluator.DefaultRegistry()                  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  OPA Evaluator                                   │  │
│  │  - Validate() → syntax/compilation errors        │  │
│  │  - CheckContract() → contract violations         │  │
│  │  - Test() → test execution results               │  │
│  │  - Lint() → regal warnings (if available)        │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Tool Specifications

### Tool 1: `validate_policy`

**Purpose:** Validate policy syntax, contract compliance, and linting

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "policyContent": {
      "type": "string",
      "description": "The Rego policy source code to validate"
    },
    "platform": {
      "type": "string",
      "enum": ["kubernetes", "terraform", "docker", "ansible", "ci"],
      "description": "Target platform for contract validation"
    }
  },
  "required": ["policyContent", "platform"]
}
```

**Output Schema (success):**
```json
{
  "valid": true,
  "syntaxErrors": [],
  "contractViolations": [],
  "lintWarnings": [
    {
      "rule": "prefer-snake-case",
      "message": "Rule name should use snake_case",
      "location": "policy.rego:5:1"
    }
  ]
}
```

**Output Schema (validation errors):**
```json
{
  "valid": false,
  "syntaxErrors": [
    "policy.rego:10:5: expected '}', found 'EOF'"
  ],
  "contractViolations": [
    {
      "path": "input.metadata.labels",
      "location": "policy.rego:12:8"
    }
  ],
  "lintWarnings": []
}
```

**Validation steps:**
1. **Syntax validation** — `evaluator.Validate(filename, policyContent)`
   - Parses Rego with RegoV1
   - Compiles to check for compilation errors
   - Returns syntax/compilation errors
2. **Contract validation** — `evaluator.CheckContract(filename, policyContent, schema)`
   - Loads CUE schema for platform
   - Walks policy AST for `input.*` references
   - Validates each reference exists in schema
   - Returns contract violations
3. **Linting** — `evaluator.Lint(filename, policyContent)`
   - Shells out to `regal` if available
   - Returns style/quality warnings
   - Gracefully returns empty if regal not installed

**Error cases:**
- Unknown platform → return error with available platforms list
- Evaluator not found → return error (shouldn't happen with default registry)
- CUE schema load failure → return error

### Tool 2: `test_policy`

**Purpose:** Validate test data against schema, then execute policy tests

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "policyContent": {
      "type": "string",
      "description": "The Rego policy source code to test"
    },
    "testData": {
      "type": "object",
      "description": "Test data conforming to platform schema (e.g., Kubernetes manifest)"
    },
    "platform": {
      "type": "string",
      "enum": ["kubernetes", "terraform", "docker", "ansible", "ci"],
      "description": "Target platform for test data validation"
    }
  },
  "required": ["policyContent", "testData", "platform"]
}
```

**Output Schema (success):**
```json
{
  "testDataValid": true,
  "testsExecuted": true,
  "results": {
    "total": 3,
    "passed": 3,
    "failed": 0,
    "errors": []
  }
}
```

**Output Schema (invalid test data):**
```json
{
  "testDataValid": false,
  "testDataErrors": [
    "input.kind: expected string matching ^(Pod|Deployment|StatefulSet|DaemonSet|Job|CronJob)$",
    "input.metadata.name: required property missing"
  ],
  "testsExecuted": false
}
```

**Output Schema (test failures):**
```json
{
  "testDataValid": true,
  "testsExecuted": true,
  "results": {
    "total": 3,
    "passed": 1,
    "failed": 2,
    "errors": [
      "test_deny_root_container: expected denial but policy allowed",
      "test_require_labels: assertion failed at line 25"
    ]
  }
}
```

**Validation and execution steps:**
1. **Schema validation** — `validateTestDataAgainstSchema(testData, platform, store)`
   - Loads JSON Schema for platform from resource store
   - Validates testData against schema using standard JSON Schema validator
   - Returns validation errors if invalid
   - **Critical:** This happens BEFORE policy execution to catch schema mismatches early
2. **Test execution** — `evaluator.Test(ctx, files)`
   - Constructs files map: `{"policy.rego": policyContent, "input.json": testData}`
   - Calls evaluator's Test method
   - Returns test results (total, passed, failed, error messages)

**Error cases:**
- Unknown platform → return error with available platforms list
- Evaluator not found → return error
- JSON Schema load failure → return error
- Test data validation failure → return testDataValid:false with errors

## Implementation Files

### New File: `internal/mcp/tools.go`

Implements tool handlers and helpers:

**Tool handler functions:**
```go
// handleValidatePolicy handles the validate_policy MCP tool.
func handleValidatePolicy(store *ResourceStore) mcp.ToolHandler {
    return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 1. Parse input
        var input struct {
            PolicyContent string `json:"policyContent"`
            Platform      string `json:"platform"`
        }
        
        // 2. Get evaluator from registry
        registry := evaluator.DefaultRegistry()
        eval, err := registry.Get("io.complytime.opa")
        
        // 3. Validate syntax
        syntaxErrs := eval.Validate("policy.rego", input.PolicyContent)
        
        // 4. Load CUE schema and check contract
        schema, err := loadCUESchemaForPlatform(input.Platform)
        contractViolations, err := eval.CheckContract("policy.rego", input.PolicyContent, schema)
        
        // 5. Run lint
        lintWarnings, _ := eval.Lint("policy.rego", input.PolicyContent)
        
        // 6. Build response
        valid := len(syntaxErrs) == 0 && len(contractViolations) == 0
        return buildValidationResponse(valid, syntaxErrs, contractViolations, lintWarnings)
    }
}

// handleTestPolicy handles the test_policy MCP tool.
func handleTestPolicy(store *ResourceStore) mcp.ToolHandler {
    return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 1. Parse input
        var input struct {
            PolicyContent string                 `json:"policyContent"`
            TestData      map[string]interface{} `json:"testData"`
            Platform      string                 `json:"platform"`
        }
        
        // 2. Validate test data against platform schema
        testDataErrs := validateTestDataAgainstSchema(input.TestData, input.Platform, store)
        if len(testDataErrs) > 0 {
            return buildTestDataErrorResponse(testDataErrs)
        }
        
        // 3. Get evaluator and execute tests
        registry := evaluator.DefaultRegistry()
        eval, err := registry.Get("io.complytime.opa")
        
        // 4. Construct test files (policy + test data as input.json)
        testDataJSON, _ := json.Marshal(input.TestData)
        files := map[string]string{
            "policy.rego": input.PolicyContent,
            "input.json":  string(testDataJSON),
        }
        
        results, err := eval.Test(ctx, files)
        
        // 5. Build response
        return buildTestResultsResponse(results)
    }
}
```

**Helper functions:**
```go
// createValidatePolicyTool creates the MCP tool definition.
func createValidatePolicyTool() *mcp.Tool

// createTestPolicyTool creates the MCP tool definition.
func createTestPolicyTool() *mcp.Tool

// loadCUESchemaForPlatform loads the CUE schema for a platform.
func loadCUESchemaForPlatform(platform string) (cue.Value, error)

// validateTestDataAgainstSchema validates test data against JSON Schema.
func validateTestDataAgainstSchema(testData map[string]interface{}, platform string, store *ResourceStore) []string

// buildValidationResponse constructs the validate_policy response.
func buildValidationResponse(valid bool, syntaxErrs []error, violations []evaluator.ContractViolation, warnings []evaluator.LintWarning) (*mcp.CallToolResult, error)

// buildTestDataErrorResponse constructs response for invalid test data.
func buildTestDataErrorResponse(errors []string) (*mcp.CallToolResult, error)

// buildTestResultsResponse constructs the test_policy response.
func buildTestResultsResponse(results *evaluator.TestResults) (*mcp.CallToolResult, error)
```

### Modified File: `internal/mcp/server.go`

Register tools during server initialization (in `NewServer` function):

```go
// After registering resources (line ~118), register tools
validateTool := createValidatePolicyTool()
mcpServer.AddTool(validateTool, handleValidatePolicy(store))

testTool := createTestPolicyTool()
mcpServer.AddTool(testTool, handleTestPolicy(store))
```

**Changes:**
- Import `internal/evaluator` package
- Import CUE packages for schema loading
- Import JSON Schema validation library (new dependency)
- Pass `store` to tool handlers (already available in NewServer scope)

### Dependencies

**New dependency:**
```go
require (
    github.com/santhosh-tekuri/jsonschema/v5 v5.3.1  // JSON Schema validation
)
```

This is the standard Go JSON Schema validator library, well-maintained and widely used.

## Testing Strategy

### Unit Tests: `internal/mcp/tools_test.go`

**Table-driven tests for `handleValidatePolicy`:**

| Test Case | Policy Content | Platform | Expected Output |
|-----------|----------------|----------|-----------------|
| Valid policy | Valid Rego | kubernetes | valid:true, no errors |
| Syntax error | Invalid Rego (missing brace) | kubernetes | valid:false, syntax error |
| Contract violation | input.invalid.field | kubernetes | valid:false, contract violation |
| Unknown platform | Valid Rego | unknown | Error response |
| Lint warnings | Valid but non-idiomatic Rego | kubernetes | valid:true, lint warnings |

**Table-driven tests for `handleTestPolicy`:**

| Test Case | Policy | Test Data | Platform | Expected Output |
|-----------|--------|-----------|----------|-----------------|
| Valid test, passing | Deny policy | Violating K8s manifest | kubernetes | testDataValid:true, tests pass |
| Valid test, failing | Deny policy | Compliant K8s manifest | kubernetes | testDataValid:true, tests fail |
| Invalid test data | Any policy | Missing required fields | kubernetes | testDataValid:false, errors |
| Unknown platform | Any policy | Valid data | unknown | Error response |
| Schema mismatch | Any policy | Wrong type for field | kubernetes | testDataValid:false, errors |

**Test fixtures:**
- `testdata/policies/valid.rego` — Valid Rego policy
- `testdata/policies/syntax-error.rego` — Rego with syntax error
- `testdata/policies/contract-violation.rego` — References invalid input.* path
- `testdata/test-data/valid-pod.json` — Valid Kubernetes Pod manifest
- `testdata/test-data/invalid-pod.json` — Pod with schema violations

### Integration Tests: `acceptance/mcp_tools_test.go`

**Full workflow tests:**

1. **Validate-repair loop:**
   - Start MCP server
   - Call `validate_policy` with invalid policy
   - Verify structured error response
   - Fix policy (simulated)
   - Call `validate_policy` again
   - Verify valid:true response

2. **Test-repair loop:**
   - Start MCP server
   - Call `test_policy` with policy and valid test data
   - Verify test execution
   - Call with invalid test data
   - Verify testDataValid:false response

3. **Full workflow (generate → validate → test):**
   - Generate policy (string)
   - Validate via MCP tool
   - Generate test data (JSON)
   - Test via MCP tool
   - Verify all steps succeed

**Error path tests:**
- Invalid MCP tool input (malformed JSON)
- Missing required fields
- Unknown platform

**Test patterns:**
- All tests use table-driven patterns
- Use Ginkgo/Gomega (matches existing acceptance tests)
- Real Kubernetes/Terraform schemas from `schemas/`

## Workflow Integration

### Updated Policy Generation Skill

The `generating-gemara-policies` skill will be updated to use these tools:

**Before (no validation loop):**
1. Read control from catalog
2. Read platform schema
3. Generate policy
4. Write to disk
5. Done

**After (validate-test-repair loop):**
1. Read control from catalog (MCP resource)
2. Read platform schema (MCP resource)
3. Generate policy (in-memory)
4. **Validate policy (MCP tool)** → syntax errors, contract violations
5. **If invalid, fix and go to step 4**
6. Generate test data (in-memory, LLM-generated)
7. **Test policy (MCP tool)** → test data validated, tests executed
8. **If tests fail, fix policy or test data, go to step 4 or 7**
9. Write policy to disk (final validated version)
10. Done

**Benefits:**
- No disk I/O until policy is valid and tested
- Fast iteration (in-memory validation/testing)
- Schema-validated test data (no manual schema checking)
- Structured error feedback (LLM can parse and fix)

## Success Criteria

**Functional:**
- ✅ `validate_policy` tool returns syntax errors for invalid Rego
- ✅ `validate_policy` tool returns contract violations for undefined input.* refs
- ✅ `validate_policy` tool returns lint warnings when available
- ✅ `test_policy` tool validates test data against platform schema
- ✅ `test_policy` tool executes policy tests and returns results
- ✅ Both tools return structured JSON responses parseable by LLM

**Non-functional:**
- ✅ Tools respond in <500ms for typical policies (<500 lines)
- ✅ Error messages are actionable (include line numbers, field names)
- ✅ All tests use table-driven patterns
- ✅ 100% test coverage for tool handlers

**Integration:**
- ✅ Tools work with existing evaluator registry
- ✅ Tools use existing platform schemas from resource store
- ✅ No changes required to existing evaluator/validator/tester packages
- ✅ MCP server startup still fails fast on misconfiguration

## Future Enhancements (Out of Scope)

**Not included in this design:**
- Auto-repair tool (Issue #12 says LLM-driven repair, not automated)
- Fixture-based integration testing via MCP (use `internal/tester` directly for that)
- Policy execution/evaluation tool (separate from validation/testing)
- Support for non-OPA policy languages (evaluator registry supports this, but no concrete use case yet)

These can be added later as separate issues if needed.
