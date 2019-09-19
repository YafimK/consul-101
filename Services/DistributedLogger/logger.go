package main

import (
	"flag"
	"fmt"
	"github.com/YafimK/consul-101/Common"
	"github.com/hashicorp/consul/api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var consulDefaultAddress = "0.0.0.0:8500"
var isLeader = false

func failOnError(err error, msg string) {
	if err != nil {
		// Exit the program.
		log.Fatalf(fmt.Sprintf("%s: %s", msg, err))
	}
}
func main() {
	consulAddress := flag.String("consul", consulDefaultAddress, "service bind port")
	flag.Parse()

	go startAPI()

	// ttl in seconds
	ttl := 10
	ttlS := fmt.Sprintf("%ds", ttl)
	serviceKey := "service/logger"
	serviceName := "logger"

	consulClient, err := Common.NewClient(*consulAddress)
	failOnError(err, "failed connecting to consul")

	sEntry := &api.SessionEntry{
		Name:      serviceName,
		TTL:       ttlS,
		LockDelay: 1 * time.Millisecond,
	}
	sID, _, err := consulClient.C.Session().Create(sEntry, nil)
	if err != nil {
		log.Fatalf("session create err: %v", err)
	}

	doneCh := make(chan struct{})
	go func() {
		err = consulClient.Session().RenewPeriodic(ttlS, sID, nil, doneCh)
		if err != nil {
			log.Fatalf("session renew err: %v", err)
		}
	}()

	log.Printf("Consul session : %+v\n", sID)

	go func() {
		hostName, err := os.Hostname()
		if err != nil {
			log.Fatalf("hostname err: %v", err)
		}

		acquireKv := &api.KVPair{
			Session: sID,
			Key:     serviceKey,
			Value:   []byte(hostName),
		}

		for {
			if !isLeader {
				acquired, _, err := client.KV().Acquire(acquireKv, nil)
				if err != nil {
					log.Fatalf("kv acquire err: %v", err)
				}

				if acquired {
					isLeader = true
					log.Printf("I'm the leader !\n")
				}
			}

			time.Sleep(time.Duration(ttl/2) * time.Second)
		}
	}()

	// wait for SIGINT or SIGTERM, clean up and exit
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	close(doneCh)
	log.Printf("Destroying session and leaving ...")
	_, err = client.Session().Destroy(sID, nil)
	if err != nil {
		log.Fatalf("session destroy err: %v", err)
	}
	os.Exit(0)
}

func startAPI() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/log", func(w http.ResponseWriter, r *http.Request) {
		if !isLeader {
			http.Error(w, "Not Leader", http.StatusBadRequest)
			return
		}

		msg, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// log msg
		log.Printf("Received %v", string(msg))

		w.Write([]byte("OK"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	port := "3000"
	log.Printf("Starting API on port %s ....\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
