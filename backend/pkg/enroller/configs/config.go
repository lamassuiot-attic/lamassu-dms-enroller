package configs

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Port string

	PostgresUser     string
	PostgresDB       string
	PostgresPassword string
	PostgresHostname string
	PostgresPort     string

	HomePath string

	EnrollerUIHost     string
	EnrollerUIPort     string
	EnrollerUIProtocol string

	KeycloakHostname string
	KeycloakPort     string
	KeycloakProtocol string
	KeycloakRealm    string
	KeycloakCA       string

	CACertFile string
	CAKeyFile  string

	CertFile string
	KeyFile  string

	OCSPServer string
}

func NewConfig(prefix string) (error, Config) {
	var cfg Config
	err := envconfig.Process(prefix, &cfg)
	if err != nil {
		return err, Config{}
	}
	return nil, cfg
}
