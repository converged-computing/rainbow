package memory

import (
	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
	"github.com/converged-computing/rainbow/pkg/types"
)

// Generate a new resource from a JGF node
// A resource is associated with a dominant subsystem resource
func NewResource(node jgf.Node) *types.Resource {

	// We assume the node has a type for metadata
	resourceType := "resource"
	typ, err := node.Metadata.GetStringElement("type")
	if err == nil {
		resourceType = typ
	}

	resourceSize := int32(1)
	size, err := node.Metadata.GetIntElement("size")
	if err == nil {
		resourceSize = size
	}

	resourceUnit := ""
	unit, err := node.Metadata.GetStringElement("unit")
	if err == nil {
		resourceUnit = unit
	}

	// Throw in the rest of the metadata for algorithms to parse
	return &types.Resource{
		Size:     resourceSize,
		Unit:     resourceUnit,
		Type:     resourceType,
		Metadata: node.Metadata,
	}
}

// New SubsystemResource creates a resource,
// but also adds arbitrary metadata
func NewSubsystemResource(node jgf.Node) *types.Resource {
	resourceType := "resource"
	typ, err := node.Metadata.GetStringElement("type")
	if err == nil {
		resourceType = typ
	}

	resourceSize := int32(1)
	size, err := node.Metadata.GetIntElement("size")
	if err == nil {
		resourceSize = size
	}

	resourceUnit := ""
	unit, err := node.Metadata.GetStringElement("unit")
	if err == nil {
		resourceUnit = unit
	}

	return &types.Resource{
		Size:     resourceSize,
		Unit:     resourceUnit,
		Type:     resourceType,
		Metadata: node.Metadata,
	}
}
