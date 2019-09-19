package main

import (
	"flag"
	"fmt"
	"github.com/YafimK/consul-101/Common"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var consulDefaultAddress = "127.0.0.1:8500"
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
func failOnError(err error, msg string) {
	if err != nil {
		// Exit the program.
		log.Fatalf(fmt.Sprintf("%s: %s", msg, err))
	}
}
func main() {
	consulAddress := flag.String("consul", consulDefaultAddress, "service bind port")
	flag.Parse()
	log.Println("Starting producer ...")
	serviceKey := "logProducer"
	client, err := Common.NewClient(*consulAddress)
	failOnError(err, "failed connecting to consul")

	timer := time.NewTicker(5 * time.Second)

	for ; msgCounter <= messageLimit; msgCounter++ {
		kv, _, err := client.ConsulClient.KV().Get(serviceKey, nil)
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
