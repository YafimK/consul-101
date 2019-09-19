package Common

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
)

type ConsulClient interface {
	//Connect(consulAddress string) (*consul.Client, error)
	RegisterService(serviceId, serviceName, hostname string, port int) error
	GetRegisteredServices(service, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error)
}

type Client struct {
	consulClient *consul.Client
}

func NewClient(addr string) (*Client, error) {
	consulApiClient, err := connectConsulClient(addr)
	if err != nil {
		return nil, err
	}
	return &Client{consulApiClient}, nil
}

func connectConsulClient(consulAddress string) (*consul.Client, error) {
	config := consul.DefaultConfig()
	config.Address = consulAddress
	return consul.NewClient(config)
}

func (c *Client) DeRegisterService(id string) error {
	return c.consulClient.Agent().ServiceDeregister(id)
}

func (c *Client) RegisterService(serviceId, serviceName, hostname string, port int) error {
	registration := new(consul.AgentServiceRegistration)
	registration.ID = serviceId
	registration.Name = serviceName
	registration.Address = hostname
	registration.Port = port
	registration.Check = new(consul.AgentServiceCheck)
	registration.Check.HTTP = fmt.Sprintf("http://%s:%d/healthcheck",
		hostname, port)
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	return c.consulClient.Agent().ServiceRegister(registration)
}

func (c *Client) GetRegisteredServices(service, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
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
