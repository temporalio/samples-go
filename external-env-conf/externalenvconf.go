package externalenvconf

import (
	"fmt"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"log"
)

func LoadProfile() client.Options {
	opts := LoadFromCustomFile()

	// Some other options of
	//opts := LoadDefaultProfile()
	//opts := LoadSpecificProfile()
	//opts := OverrideAfterLoading()

	return opts
}

func LoadFromCustomFile() client.Options {
	configFilePath := "config.toml"

	opts, err := envconfig.LoadClientOptions(envconfig.LoadClientOptionsRequest{
		ConfigFilePath: configFilePath,
	})
	if err != nil {
		log.Fatalf("failed to load client config from custom file: %v", err)
	}

	fmt.Println("opts", opts)

	fmt.Printf("✅ Connected using custom config: %s\n", configFilePath)

	return opts
}

func LoadDefaultProfile() client.Options {
	opts, err := envconfig.LoadClientOptions(envconfig.LoadClientOptionsRequest{})
	if err != nil {
		log.Fatalf("failed to load default client options: %v", err)
	}

	fmt.Printf("✅ Connected to Temporal with default client options\n")

	return opts
}

// LoadSpecificProfile does not actually run due to invalid TLS config,
// but accurately demonstrates how to load a specific profile.
func LoadSpecificProfile() client.Options {
	profile := "prod"
	configFilePath := "config.toml"

	opts, err := envconfig.LoadClientOptions(envconfig.LoadClientOptionsRequest{
		ConfigFilePath:    configFilePath,
		ConfigFileProfile: profile,
	})
	if err != nil {
		log.Fatalf("failed to load 'prod' profile: %v", err)
	}

	fmt.Printf("✅ Connected to Temporal using '%v' profile\n", profile)

	return opts
}

// OverrideAfterLoading demonstrates how to override options after loading from a config.
// For this sample to work, "test-namespace" must already exist.
func OverrideAfterLoading() client.Options {
	// Load base config (e.g., default profile)
	opts, err := envconfig.LoadClientOptions(envconfig.LoadClientOptionsRequest{})
	if err != nil {
		log.Fatalf("failed to load base client options: %v", err)
	}

	// Apply overrides programmatically
	opts.HostPort = "localhost:7233"
	opts.Namespace = "test-namespace"

	fmt.Printf("✅ Connected with overridden config to: %s in namespace: %s\n", opts.HostPort, opts.Namespace)

	return opts
}
