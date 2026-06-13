// SPDX-License-Identifier: Apache-2.0

package requirement

import (
	"testing"

	"github.com/gemaraproj/go-gemara"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareValues_Aligned(t *testing.T) {
	fw := ParameterLayer{Source: "framework", Value: "1.3", Specificity: SpecificityConcrete}
	org := ParameterLayer{Source: "org-policy", Value: "1.3", Specificity: SpecificityConcrete}

	verdict := CompareValues(fw, org)
	assert.Equal(t, VerdictAligned, verdict)
}

func TestCompareValues_Mismatch(t *testing.T) {
	t.Run("version strings", func(t *testing.T) {
		fw := ParameterLayer{Source: "fw", Value: "1.2", Specificity: SpecificityConcrete}
		org := ParameterLayer{Source: "org", Value: "1.3", Specificity: SpecificityConcrete}
		verdict := CompareValues(fw, org)
		assert.Equal(t, VerdictMismatch, verdict)
	})

	t.Run("numeric thresholds", func(t *testing.T) {
		fw := ParameterLayer{Source: "fw", Value: "30", Specificity: SpecificityConcrete}
		org := ParameterLayer{Source: "org", Value: "60", Specificity: SpecificityConcrete}
		verdict := CompareValues(fw, org)
		assert.Equal(t, VerdictMismatch, verdict)
	})

	t.Run("algorithms", func(t *testing.T) {
		fw := ParameterLayer{Source: "fw", Value: "AES-256-GCM", Specificity: SpecificityConcrete}
		org := ParameterLayer{Source: "org", Value: "ChaCha20-Poly1305", Specificity: SpecificityConcrete}
		verdict := CompareValues(fw, org)
		assert.Equal(t, VerdictMismatch, verdict)
	})
}

func TestCompareValues_OrgBindsGeneric(t *testing.T) {
	fw := ParameterLayer{Source: "fw", Value: "per organizational requirements", Specificity: SpecificityGeneric}
	org := ParameterLayer{Source: "org", Value: "MFA + bastion host", Specificity: SpecificityConcrete}
	verdict := CompareValues(fw, org)
	assert.Equal(t, VerdictOrgBindsGeneric, verdict)
}

func TestCompareValues_NotCovered(t *testing.T) {
	t.Run("both none", func(t *testing.T) {
		fw := ParameterLayer{Specificity: SpecificityNone}
		org := ParameterLayer{Specificity: SpecificityNone}
		verdict := CompareValues(fw, org)
		assert.Equal(t, VerdictNotCovered, verdict)
	})

	t.Run("framework none", func(t *testing.T) {
		fw := ParameterLayer{Specificity: SpecificityNone}
		org := ParameterLayer{Source: "org", Value: "90", Specificity: SpecificityConcrete}
		verdict := CompareValues(fw, org)
		assert.Equal(t, VerdictNotCovered, verdict)
	})

	t.Run("org none", func(t *testing.T) {
		fw := ParameterLayer{Source: "fw", Value: "90", Specificity: SpecificityConcrete}
		org := ParameterLayer{Specificity: SpecificityNone}
		verdict := CompareValues(fw, org)
		assert.Equal(t, VerdictNotCovered, verdict)
	})
}

func testDeltaArtifactSet() *ArtifactSet {
	catalog := &gemara.ControlCatalog{
		Metadata: gemara.Metadata{Id: "container-baseline"},
		Controls: []gemara.Control{
			{
				Id:    "CTL-TLS-001",
				Title: "TLS Configuration",
				AssessmentRequirements: []gemara.AssessmentRequirement{
					{Id: "CTL-TLS-001-AR1", Text: "TLS minimum version must be enforced"},
				},
			},
			{
				Id:    "CTL-CERT-001",
				Title: "Certificate Management",
				AssessmentRequirements: []gemara.AssessmentRequirement{
					{Id: "CTL-CERT-001-AR1", Text: "Certificate validity must not exceed maximum"},
				},
			},
		},
	}

	policy := &gemara.Policy{
		Metadata: gemara.Metadata{
			Id: "org-parent-policy",
			MappingReferences: []gemara.MappingReference{
				{Id: "container-baseline"},
			},
		},
		Imports: gemara.Imports{
			Catalogs: []gemara.CatalogImport{
				{ReferenceId: "container-baseline"},
			},
		},
		Adherence: gemara.Adherence{
			AssessmentPlans: []gemara.AssessmentPlan{
				{
					RequirementId: "CTL-TLS-001-AR1",
					Parameters: []gemara.Parameter{
						{Label: "tls_minimum_version", AcceptedValues: []string{"1.3"}},
					},
				},
				{
					RequirementId: "CTL-CERT-001-AR1",
					Parameters: []gemara.Parameter{
						{Label: "max_validity_days", AcceptedValues: []string{"90"}},
					},
				},
			},
		},
	}

	return &ArtifactSet{
		Catalogs: map[string]*gemara.ControlCatalog{"container-baseline": catalog},
		Policies: map[string]*gemara.Policy{"org-parent-policy": policy},
		Guidance: make(map[string]*gemara.GuidanceCatalog),
	}
}

func TestAnalyzeDelta(t *testing.T) {
	set := testDeltaArtifactSet()
	policy := set.Policies["org-parent-policy"]

	rp, err := ResolvePolicy(*policy, set)
	require.NoError(t, err)

	report, err := AnalyzeDelta(rp, set)
	require.NoError(t, err)

	assert.Equal(t, "org-parent-policy", report.PolicyID)
	assert.Contains(t, report.CatalogsCompared, "container-baseline")
	assert.Len(t, report.Parameters, 2)
	assert.Equal(t, report.Summary.Total, 2)
}

func TestAnalyzeDelta_NilPolicy(t *testing.T) {
	set := testDeltaArtifactSet()
	_, err := AnalyzeDelta(nil, set)
	assert.Error(t, err)
}
