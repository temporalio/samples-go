package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/temporalio/samples-go/cryptconverter"
	"go.temporal.io/sdk/converter"
	cliplugin "go.temporal.io/server/tools/cli/plugin"
)

func main() {
	var pluginMap = map[string]plugin.Plugin{
		cliplugin.DataConverterPluginType: &cliplugin.DataConverterPlugin{
			Impl: cryptconverter.NewCryptDataConverter(
				converter.GetDefaultDataConverter(),
			),
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: cliplugin.PluginHandshakeConfig,
		Plugins:         pluginMap,
	})
}
