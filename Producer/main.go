package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var consulDefaultAddress = "consul:8500"
var messageLimit = 50
var msgCounter = 0

func sendMessage(targetHostname string, payload string) {
	message := fmt.Sprintf("Message %d", msgCounter)
	log.Printf("Sending message: %v\n", message)
	resp, err := http.Post(fmt.Sprintf("http://%v:3000/api/v1/log", targetHostname), "text/plain", strings.NewReader(payload))
	if err != nil {
		log.Printf("http post err: %v", err)
		return
	}

	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Printf("status not OK: %v", resp.StatusCode)
		return
	}

	log.Printf("msg sent OK")
}

func main() {
	log.Println("Starting producer ...")
	serviceKey := "service/logger"

	config := api.DefaultConfig()
	config.Address = consulDefaultAddress

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("client err: %v", err)
	}

	timer := time.NewTicker(5 * time.Second)

	for ; msgCounter <= messageLimit; msgCounter++ {
		kv, _, err := client.KV().Get(serviceKey, nil)
		if err != nil {
			log.Fatalf("kv acquire err: %v", err)
		}

		if kv != nil && kv.Session != "" {
			// there is a leader
			leaderHostname := string(kv.Value)
			sendMessage(leaderHostname, "hola")
		}
		<-timer.C
	}

	log.Println("Exiting producer ...")
}
