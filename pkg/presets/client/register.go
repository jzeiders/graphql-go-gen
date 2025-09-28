package client

import (
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/fragment_masking"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/gql_tag_operations"
	"github.com/jzeiders/graphql-go-gen/pkg/presets"
)

func init() {
	// Register the client preset
	presets.Register("client", &ClientPreset{})

	// Register the plugins used by client preset
	// These will be registered when the plugin packages are imported
	_ = plugin.Register("gql-tag-operations", gql_tag_operations.New())
	_ = plugin.Register("fragment-masking", fragment_masking.New())
}