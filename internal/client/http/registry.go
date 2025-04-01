package httpclient

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type RegistryClient struct {
	registryAddress string
	serviceName     string
	port            string
}

func NewRegistryClient(registryAddress, serviceName, port string) *RegistryClient {
	client := &RegistryClient{
		registryAddress: registryAddress,
		serviceName:     serviceName,
		port:            port,
	}

	go client.registerService()
	go client.sendHeartbeats()

	return client
}

func (r *RegistryClient) registerService() {
	url := fmt.Sprintf("%s/register?serviceName=%s&address=%s", r.registryAddress, r.serviceName, fmt.Sprintf("localhost:%s", r.port))
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("❌ Could not register %v: %v", r.serviceName, err)
	}
	defer resp.Body.Close()

	log.Printf("✅ Registered %v on port %s", r.serviceName, r.port)
}

func (r *RegistryClient) sendHeartbeats() {
	url := fmt.Sprintf("%s/heartbeat?serviceName=%s&address=%s", r.registryAddress, r.serviceName, fmt.Sprintf("localhost:%s", r.port))

	for {
		time.Sleep(1 * time.Second)
		_, err := http.Get(url)
		if err != nil {
			log.Println("❌ Failed to send heartbeat:", err)
		} else {
			// log.Println("💓 Heartbeat sent successfully")
		}
	}
}
