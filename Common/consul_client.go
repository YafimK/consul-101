package Common

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
)

type ConsulClient interface {
	Connect(consulAddress string) (*consulapi.Client, error)
	RegisterService(serviceId, serviceName, hostname string, port int) error
	GetRegisteredServices(service, tag string) ([]*consulapi.ServiceEntry, *consulapi.QueryMeta, error)
}

type client struct {
	consulClient *consulapi.Client
}

func (c *client) Connect(consulAddress string) (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.Address = consulAddress
	return consulapi.NewClient(config)
}

func (c *client) RegisterService(serviceId, serviceName, hostname string, port int) error {
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = serviceId
	registration.Name = serviceName
	registration.Address = hostname
	registration.Port = port
	registration.Check = new(consulapi.AgentServiceCheck)
	registration.Check.HTTP = fmt.Sprintf("http://%s:%d/healthcheck",
		hostname, port)
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	return c.consulClient.Agent().ServiceRegister(registration)
}

func (c *client) GetRegisteredServices(service, tag string) ([]*consulapi.ServiceEntry, *consulapi.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := c.consulClient.Health().Service(service, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", service)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}
