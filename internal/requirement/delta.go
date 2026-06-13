// SPDX-License-Identifier: Apache-2.0

package requirement

import (
	"fmt"
	"strings"

	"github.com/gemaraproj/go-gemara"
)

// Verdict classifies the relationship between parameter values from
// different sources in a resolved policy graph.
type Verdict string

const (
	VerdictAligned         Verdict = "aligned"
	VerdictMismatch        Verdict = "mismatch"
	VerdictOrgBindsGeneric Verdict = "org_binds_generic"
	VerdictNotCovered      Verdict = "not_covered"
)

// Specificity describes how concrete a parameter value is.
type Specificity string

const (
	SpecificityConcrete Specificity = "concrete"
	SpecificityGeneric  Specificity = "generic"
	SpecificityNone     Specificity = "none"
)

// ParameterLayer holds a parameter value from one source.
type ParameterLayer struct {
	Source      string      `json:"source"`
	Value       string      `json:"value"`
	Specificity Specificity `json:"specificity"`
}

// ParameterDelta is the result of comparing a single parameter
// across framework, org policy, and tech baseline layers.
type ParameterDelta struct {
	RequirementID string         `json:"requirement_id"`
	Label         string         `json:"label"`
	Framework     ParameterLayer `json:"framework"`
	OrgPolicy     ParameterLayer `json:"org_policy"`
	TechBaseline  ParameterLayer `json:"tech_baseline"`
	Verdict       Verdict        `json:"verdict"`
}

// DeltaReport is the full result of analyzing parameter deltas
// across a resolved policy.
type DeltaReport struct {
	PolicyID         string           `json:"policy"`
	CatalogsCompared []string         `json:"catalogs_compared"`
	Parameters       []ParameterDelta `json:"parameters"`
	Summary          DeltaSummary     `json:"summary"`
}

// DeltaSummary counts verdicts.
type DeltaSummary struct {
	Total           int `json:"total"`
	Aligned         int `json:"aligned"`
	Mismatch        int `json:"mismatch"`
	OrgBindsGeneric int `json:"org_binds_generic"`
	NotCovered      int `json:"not_covered"`
}

// CompareValues determines the verdict between a framework layer and
// an org policy layer.
func CompareValues(framework, orgPolicy ParameterLayer) Verdict {
	if framework.Specificity == SpecificityNone {
		return VerdictNotCovered
	}
	if orgPolicy.Specificity == SpecificityNone {
		return VerdictNotCovered
	}

	if framework.Specificity == SpecificityGeneric && orgPolicy.Specificity == SpecificityConcrete {
		return VerdictOrgBindsGeneric
	}

	if framework.Value == orgPolicy.Value {
		return VerdictAligned
	}

	return VerdictMismatch
}

func classifySpecificity(value string) Specificity {
	if value == "" {
		return SpecificityNone
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "per organizational") ||
		strings.Contains(lower, "per the organization") ||
		strings.Contains(lower, "as defined by") ||
		strings.Contains(lower, "according to") {
		return SpecificityGeneric
	}
	return SpecificityConcrete
}

func findGuidelineParameter(gc gemara.GuidanceCatalog, label string) (string, bool) {
	return "", false
}

func summarizeDeltas(deltas []ParameterDelta) DeltaSummary {
	s := DeltaSummary{Total: len(deltas)}
	for _, d := range deltas {
		switch d.Verdict {
		case VerdictAligned:
			s.Aligned++
		case VerdictMismatch:
			s.Mismatch++
		case VerdictOrgBindsGeneric:
			s.OrgBindsGeneric++
		case VerdictNotCovered:
			s.NotCovered++
		}
	}
	return s
}

// AnalyzeDelta compares parameters across all layers in a resolved policy.
func AnalyzeDelta(rp *ResolvedPolicy, set *ArtifactSet) (*DeltaReport, error) {
	if rp == nil {
		return nil, fmt.Errorf("resolved policy is nil")
	}

	var catalogIDs []string
	for _, cat := range rp.ControlCatalogs {
		catalogIDs = append(catalogIDs, cat.Metadata.Id)
	}

	var deltas []ParameterDelta
	for _, plan := range rp.Policy.Adherence.AssessmentPlans {
		for _, param := range plan.Parameters {
			orgValue := ""
			if len(param.AcceptedValues) > 0 {
				orgValue = param.AcceptedValues[0]
			}

			orgLayer := ParameterLayer{
				Source:      rp.Policy.Metadata.Id,
				Value:       orgValue,
				Specificity: classifySpecificity(orgValue),
			}

			fwLayer := ParameterLayer{Specificity: SpecificityNone}
			baselineLayer := ParameterLayer{Specificity: SpecificityNone}

			for _, gc := range rp.GuidanceCatalogs {
				if v, ok := findGuidelineParameter(gc, param.Label); ok {
					fwLayer = ParameterLayer{
						Source:      gc.Metadata.Id,
						Value:       v,
						Specificity: classifySpecificity(v),
					}
					break
				}
			}

			verdict := CompareValues(fwLayer, orgLayer)

			deltas = append(deltas, ParameterDelta{
				RequirementID: plan.RequirementId,
				Label:         param.Label,
				Framework:     fwLayer,
				OrgPolicy:     orgLayer,
				TechBaseline:  baselineLayer,
				Verdict:       verdict,
			})
		}
	}

	summary := summarizeDeltas(deltas)

	return &DeltaReport{
		PolicyID:         rp.Policy.Metadata.Id,
		CatalogsCompared: catalogIDs,
		Parameters:       deltas,
		Summary:          summary,
	}, nil
}
