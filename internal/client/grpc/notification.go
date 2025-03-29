package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	pbNotification "github.com/tiendat-go/proto-service/gen/notification/v1"
	pbRegistry "github.com/tiendat-go/proto-service/gen/registry/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationClient struct {
	registryClient *RegistryClient
}

func NewNotificationClient(registryClient *RegistryClient) *NotificationClient {
	return &NotificationClient{
		registryClient: registryClient,
	}
}

func (n *NotificationClient) GetNotifications(req *pbNotification.GetNotificationsRequest) (*pbNotification.GetNotificationsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resRegistry, err := n.registryClient.client.GetRandService(ctx, &pbRegistry.GetRandServiceRequest{ServiceName: "notification-service"})
	if err != nil {
		return nil, fmt.Errorf("❌ Could not get notification service address: %w", err)
	}

	notificationConn, err := grpc.Dial(resRegistry.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("❌ Could not connect to notification service: %w", err)
	}
	defer notificationConn.Close()

	notificationClient := pbNotification.NewNotificationServiceClient(notificationConn)
	res, err := notificationClient.GetNotifications(ctx, req)
	if err != nil {
		log.Printf("❌ Failed to get notifications: %v", err)
		return nil, err
	}

	return res, nil
}
