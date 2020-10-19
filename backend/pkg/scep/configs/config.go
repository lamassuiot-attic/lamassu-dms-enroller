package configs

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Port string

	PostgresUser     string
	PostgresDB       string
	PostgresPort     string
	PostgresPassword string
	PostgresHostname string

	EnrollerUIHost     string
	EnrollerUIPort     string
	EnrollerUIProtocol string

	KeycloakHostname string
	KeycloakPort     string
	KeycloakProtocol string
	KeycloakRealm    string

	CertFile string
	KeyFile  string
}

func NewConfig() (error, Config) {
	var cfg Config
	err := envconfig.Process("scep", &cfg)
	if err != nil {
		return err, Config{}
	}
	return nil, cfg
}
