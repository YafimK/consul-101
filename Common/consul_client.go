package Common

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"log"
	"net"
)

type ConsulClient interface {
	//Connect(consulAddress string) (*consul.Client, error)
	RegisterService(serviceId, serviceName, hostname string, port int) error
	GetRegisteredServices(service, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error)
}

type Client struct {
	ConsulClient *consul.Client
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
	return c.ConsulClient.Agent().ServiceDeregister(id)
}

func (c *Client) RegisterService(serviceId, serviceName, hostname string, port int) error {
	registration := new(consul.AgentServiceRegistration)
	registration.ID = serviceId
	registration.Name = serviceName
	ip, err := getIpAddress()
	if err != nil {
		return err
	}
	registration.Address = ip
	registration.Port = port
	registration.Check = new(consul.AgentServiceCheck)
	registration.Check.HTTP = fmt.Sprintf("http://%s:%d/healthcheck",
		hostname, port)
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	log.Printf("Registering service [%v], under host [%v], with ip: [%v]\n", serviceName, hostname, ip)
	return c.ConsulClient.Agent().ServiceRegister(registration)
}

func (c *Client) GetRegisteredServices(service, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := c.ConsulClient.Health().Service(service, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", service)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}

func getIpAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("didn't find internal ip address" +
		"" +
		"")
}
