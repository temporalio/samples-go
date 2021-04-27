package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/temporalio/samples-go/cryptconverter"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/tools/cli"
)

func main() {
	var pluginMap = map[string]plugin.Plugin{
		"DataConverter": &cli.DataConverterPlugin{
			Impl: cryptconverter.NewCryptDataConverter(
				converter.GetDefaultDataConverter(),
			),
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: cli.PluginHandshakeConfig,
		Plugins:         pluginMap,
	})
}
