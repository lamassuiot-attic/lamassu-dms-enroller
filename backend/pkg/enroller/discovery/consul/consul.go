package consul

import (
	"log"
	"strconv"

	"math/rand"

	"enroller/discovery"

	"github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"google.golang.org/api/discovery/v1"
)

type ServiceDiscovery struct {
	client    consulsd.Client
	logger    log.Logger
	registrar *consulsd.Registrar
}

func NewServiceDiscovery(consulProtocol string, consulHost string, consulPort string, logger log.Logger) (discovery.Service, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulProtocol + "://" + consulHost + ":" + consulPort
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	client := consulsd.NewClient(consulClient)
	return &ServiceDiscovery{client, logger}, nil
}

func (sd *ServiceDiscovery) Register(advProtocol string, advHost string, advPort string) error {
	check := api.AgentServiceCheck{
		HTTP:     advProtocol + "://" + advHost + ":" + advPort + "/health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, _ := strconv.Atoi(advPort)
	num := rand.Intn(100)
	asr := api.AgentServiceRegistration{
		ID:      "enroller" + strconv.Itoa(num),
		Name:    "enroller",
		Address: advProtocol + "://" + advHost,
		Port:    port,
		Tags:    []string{"enroller", "enroller"},
		Check:   &check,
	}
	sd.registrar = consulsd.NewRegistrar(sd.client, &asr, sd.logger)
	sd.registrar.Register()
	return nil
}

func (sd *ServiceDiscovery) Deregister() error {
	sd.registrar.Deregister()
	return nil
}
