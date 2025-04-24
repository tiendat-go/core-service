package grpc

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tiendat-go/core-service/internal/core/config"
	pbRegistry "github.com/tiendat-go/proto-service/gen/registry/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RegistryClient struct {
	client      pbRegistry.DiscoveryServiceClient
	addr        string
	port        string
	serviceName string
}

func NewRegistryClient(cfg *config.Config) *RegistryClient {
	conn, err := grpc.Dial(cfg.RegistryURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Could not connect to registry: %v", err)
	}

	client := &RegistryClient{
		client:      pbRegistry.NewDiscoveryServiceClient(conn),
		addr:        cfg.RegistryURL,
		port:        strconv.Itoa(cfg.ServicePort),
		serviceName: cfg.ServiceName,
	}

	go client.registerService()
	go client.sendHeartbeats()

	return client
}

func (r *RegistryClient) registerService() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := r.client.RegisterService(ctx, &pbRegistry.RegisterServiceRequest{
		ServiceName: r.serviceName,
		Address:     fmt.Sprintf("localhost:%s", r.port),
	})
	if err != nil {
		log.Fatalf("❌ Could not register %v: %v", r.serviceName, err)
	}

	log.Printf("✅ Registered %v on port %s", r.serviceName, r.port)
}

func (r *RegistryClient) sendHeartbeats() {
	for {
		time.Sleep(1 * time.Second)
		ctx, _ := context.WithTimeout(context.Background(), time.Second)

		_, err := r.client.Heartbeat(ctx, &pbRegistry.HeartbeatRequest{
			ServiceName: r.serviceName,
			Address:     fmt.Sprintf("localhost:%s", r.port),
		})
		if err != nil {
			log.Println("❌ Failed to send heartbeat:", err)
		} else {
			// log.Println("💓 Heartbeat sent successfully")
		}
	}
}
