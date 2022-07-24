package common

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"go.temporal.io/sdk/client"
)

func GetClientOptions() client.Options {
	configHome := getEnvOrDefaultString(EnvKeyConfigHome, EnvDefaultConfigHome)
	configEnv := getEnvOrDefaultString(EnvKeyEnvironment, EnvDefaultEnvironment)

	configRoot := filepath.Join(configHome, configEnv)
	// Use Home Directory to contruct the configRoot if it is specified as relative path
	if !filepath.IsAbs(configRoot) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		configRoot = filepath.Join(homeDir, configRoot)
	}

	if _, err := os.Stat(configRoot); err != nil {
		// Config root does not exist
		// Use default options to initialize client
		return client.Options{}
	}

	var cfg Temporal
	if err := loadConfig(configRoot, &cfg); err != nil {
		panic(err)
	}

	options := client.Options{
		HostPort:  cfg.HostNameAndPort,
		Namespace: cfg.Namespace,
	}

	tlsConfig, err := getTLSConfiguration(configRoot, cfg)
	if err != nil {
		panic(err)
	}
	if tlsConfig != nil {
		options.ConnectionOptions = client.ConnectionOptions{
			TLS: tlsConfig,
		}
	}

	return options
}

func getTLSConfiguration(configRoot string, config Temporal) (*tls.Config, error) {
	// Ignoring error as we'll fail to dial anyway, and that will produce a meaningful error
	host, _, _ := net.SplitHostPort(config.HostNameAndPort)

	caBytes, err := getCertificateData(configRoot, config.CaFile, config.CaData)
	if err != nil {
		return nil, err
	}

	clientCertBytes, err := getCertificateData(configRoot, config.CertFile, config.CertData)
	if err != nil {
		return nil, err
	}

	clientCertKeyBytes, err := getCertificateData(configRoot, config.KeyFile, config.KeyData)
	if err != nil {
		return nil, err
	}

	var cert *tls.Certificate
	var caPool *x509.CertPool

	if len(clientCertBytes) > 0 {
		clientCert, err := tls.X509KeyPair(clientCertBytes, clientCertKeyBytes)
		if err != nil {
			return nil, err
		}
		cert = &clientCert
	}

	if len(caBytes) > 0 {
		caPool = x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caBytes) {
			return nil, errors.New("unknown failure constructing cert pool for ca")
		}
	}

	// If we are given arguments to verify either server or client, configure TLS
	if caPool != nil || cert != nil {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: !config.EnableHostVerification,
			ServerName:         host,
		}
		if caPool != nil {
			tlsConfig.RootCAs = caPool
		}
		if cert != nil {
			tlsConfig.Certificates = []tls.Certificate{*cert}
		}

		return tlsConfig, nil
	}

	return nil, nil
}

func getCertificateData(
	configRoot string,
	fileSource string,
	configSource string,
) ([]byte, error) {
	if fileSource != "" && configSource != "" {
		return nil, fmt.Errorf("cannot supply both file path and base64-encoded value")
	}

	if configSource != "" {
		return base64.StdEncoding.DecodeString(configSource)
	}

	if fileSource != "" {
		return ioutil.ReadFile(filepath.Join(configRoot, fileSource))
	}

	return nil, nil
}
