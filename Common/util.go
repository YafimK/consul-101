package Common

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"log"
)

func RegisterServiceWithConsul(serviceId, serviceName string, hostname string, port int) error {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatalln(err)
	}
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
	return consul.Agent().ServiceRegister(registration)
}
