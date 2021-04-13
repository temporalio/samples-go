package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/temporalio/samples-go/cryptconverter"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/tools/cli"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEMPORAL_CLI_PLUGIN_DATA_CONVERTER",
	MagicCookieValue: "abb3e448baf947eba1847b10a38554db",
}

func main() {
	var pluginMap = map[string]plugin.Plugin{
		"DataConverter": &cli.DataConverterPlugin{
			Impl: cryptconverter.NewCryptDataConverter(
				converter.GetDefaultDataConverter(),
			),
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
