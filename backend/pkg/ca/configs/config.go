package configs

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Port string

	EnrollerUIHost     string
	EnrollerUIPort     string
	EnrollerUIProtocol string

	KeycloakHostname string
	KeycloakPort     string
	KeycloakProtocol string
	KeycloakRealm    string

	VaultAddress  string
	VaultRoleID   string
	VaultSecretID string

	CertFile string
	KeyFile  string
}

func NewConfig() (error, Config) {
	var cfg Config
	err := envconfig.Process("ca", &cfg)
	if err != nil {
		return err, Config{}
	}
	return nil, cfg
}
