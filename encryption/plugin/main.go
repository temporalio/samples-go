package main

import (
	"github.com/hashicorp/go-plugin"
	"go.temporal.io/sdk/converter"
	cliplugin "go.temporal.io/server/tools/cli/plugin"

	"github.com/temporalio/samples-go/encryption"
)

func main() {
	var pluginMap = map[string]plugin.Plugin{
		cliplugin.DataConverterPluginType: &cliplugin.DataConverterPlugin{
			Impl: encryption.NewEncryptionDataConverter(
				converter.GetDefaultDataConverter(),
				encryption.DataConverterOptions{Compress: true},
			),
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: cliplugin.PluginHandshakeConfig,
		Plugins:         pluginMap,
	})
}
