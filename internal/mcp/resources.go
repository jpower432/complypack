// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/gemaraproj/go-gemara"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ResourceStore manages catalogs and schemas for MCP resource handlers.
// It holds both raw YAML (for MCP resource serving) and parsed artifacts (for tool handlers).
type ResourceStore struct {
	rawCatalogs map[string][]byte                      // raw YAML for ReadResource
	catalogs    map[string]*gemara.ControlCatalog      // parsed ControlCatalogs
	policies    map[string]*gemara.Policy              // parsed Policies
	effective   map[string]*gemara.EffectivePolicy     // resolved policy graphs
	schemas     map[string][]byte                      // platform JSON schemas
}

// NewResourceStore creates a ResourceStore with raw and parsed artifacts.
func NewResourceStore(
	rawCatalogs map[string][]byte,
	catalogs map[string]*gemara.ControlCatalog,
	policies map[string]*gemara.Policy,
	effective map[string]*gemara.EffectivePolicy,
	schemas map[string][]byte,
) *ResourceStore {
	return &ResourceStore{
		rawCatalogs: rawCatalogs,
		catalogs:    catalogs,
		policies:    policies,
		effective:   effective,
		schemas:     schemas,
	}
}

// ListResources returns all available catalog and schema resources.
func (rs *ResourceStore) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	var resources []mcp.Resource

	// Add catalog resources (from raw catalogs for ReadResource)
	for name := range rs.rawCatalogs {
		resources = append(resources, mcp.Resource{
			URI:      fmt.Sprintf("%s://%s/%s", URIScheme, ResourceTypeCatalog, name),
			Name:     fmt.Sprintf("Gemara Catalog: %s", name),
			MIMEType: MIMETypeYAML,
		})
	}

	// Add schema resources
	for platform := range rs.schemas {
		resources = append(resources, mcp.Resource{
			URI:      fmt.Sprintf("%s://%s/%s", URIScheme, ResourceTypeSchema, platform),
			Name:     fmt.Sprintf("Platform Schema: %s", platform),
			MIMEType: MIMETypeJSONSchema,
		})
	}

	return resources, nil
}

// ReadResource returns the content for a specific resource URI.
func (rs *ResourceStore) ReadResource(ctx context.Context, uri string) ([]*mcp.ResourceContents, error) {
	// Parse URI: complypack://catalog/<name> or complypack://schema/<platform>
	if !strings.HasPrefix(uri, URIScheme+"://") {
		return nil, fmt.Errorf("invalid URI scheme: expected %s://", URIScheme)
	}

	path := strings.TrimPrefix(uri, URIScheme+"://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid URI format: %s", uri)
	}

	resourceType := parts[0]
	name := parts[1]

	switch resourceType {
	case ResourceTypeCatalog:
		data, ok := rs.rawCatalogs[name]
		if !ok {
			return nil, fmt.Errorf("catalog %q not found", name)
		}
		return []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: MIMETypeYAML,
			Text:     string(data),
		}}, nil

	case ResourceTypeSchema:
		data, ok := rs.schemas[name]
		if !ok {
			return nil, fmt.Errorf("schema %q not found", name)
		}
		return []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: MIMETypeJSONSchema,
			Text:     string(data),
		}}, nil

	default:
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}
