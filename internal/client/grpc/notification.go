package grpc

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	pbNotification "github.com/tiendat-go/proto-service/gen/notification/v1"
	pbRegistry "github.com/tiendat-go/proto-service/gen/registry/v1"
	"google.golang.org/grpc"
)

type NotificationClient struct {
	registryClient *RegistryClient
	connectionPool map[string]*grpc.ClientConn
	mu             sync.Mutex // protects the connectionPool
}

func NewNotificationClient(registryClient *RegistryClient) *NotificationClient {
	client := &NotificationClient{
		registryClient: registryClient,
		connectionPool: make(map[string]*grpc.ClientConn),
	}
	client.refreshConnectionPool()
	client.startHealthCheck(1 * time.Second)
	return client
}

func (n *NotificationClient) startHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			n.refreshConnectionPool()
		}
	}()
}

func (n *NotificationClient) refreshConnectionPool() {
	ctx := context.Background()
	services, err := n.registryClient.client.GetServices(ctx, &pbRegistry.GetServicesRequest{ServiceName: "notification-service"})
	if err != nil {
		log.Printf("❌ Failed to refresh service list: %v", err)
		return
	}

	newAddrs := make(map[string]struct{})
	for _, addr := range services.Addresses {
		newAddrs[addr] = struct{}{}
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// Close and remove outdated connections
	for addr, conn := range n.connectionPool {
		if _, ok := newAddrs[addr]; !ok {
			log.Printf("🧹 Removing stale connection to %s", addr)
			conn.Close()
			delete(n.connectionPool, addr)
		}
	}

	// Add new connections
	for addr := range newAddrs {
		if _, exists := n.connectionPool[addr]; !exists {
			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				log.Printf("❌ Failed to dial new address %s: %v", addr, err)
				continue
			}
			n.connectionPool[addr] = conn
		}
	}
}

func (n *NotificationClient) getRandomConnection() *grpc.ClientConn {
	n.mu.Lock()
	defer n.mu.Unlock()

	numConns := len(n.connectionPool)
	if numConns == 0 {
		return nil
	}

	// Randomly select one connection
	i := rand.Intn(numConns)
	idx := 0
	for _, conn := range n.connectionPool {
		if idx == i {
			return conn
		}
		idx++
	}
	return nil
}

func (n *NotificationClient) GetNotifications(req *pbNotification.GetNotificationsRequest) (*pbNotification.GetNotificationsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Select a random connection from the pool
	conn := n.getRandomConnection()
	if conn == nil {
		return nil, fmt.Errorf("❌ No available connections to notification service")
	}

	notificationClient := pbNotification.NewNotificationServiceClient(conn)
	res, err := notificationClient.GetNotifications(ctx, req)
	if err != nil {
		log.Printf("❌ Failed to get notifications: %v", err)
		return nil, err
	}

	return res, nil
}
