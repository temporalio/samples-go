package common

type (
	Temporal struct {
		HostNameAndPort string `yaml:"host"`
		Namespace       string `yaml:"namespace"`

		CaFile   string `yaml:"caFile"`
		CertFile string `yaml:"certFile"`
		KeyFile  string `yaml:"keyFile"`

		CaData   string `yaml:"caData"`
		CertData string `yaml:"certData"`
		KeyData  string `yaml:"keyData"`

		EnableHostVerification bool `yaml:"enableHostVerification"`
	}
)
