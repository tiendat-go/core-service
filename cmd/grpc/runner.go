package main

import (
	"log"

	grpcClient "github.com/tiendat-go/core-service/internal/client/grpc"
	grpcCtrl "github.com/tiendat-go/core-service/internal/controller/grpc"
	"github.com/tiendat-go/core-service/internal/core/config"
	"github.com/tiendat-go/core-service/internal/core/server"
	"github.com/tiendat-go/core-service/internal/core/server/grpc"
	"github.com/tiendat-go/core-service/internal/core/service"
	infraConf "github.com/tiendat-go/core-service/internal/infra/config"
	"github.com/tiendat-go/core-service/internal/infra/repository"
	pbCore "github.com/tiendat-go/proto-service/gen/core/v1"
	pbNotification "github.com/tiendat-go/proto-service/gen/notification/v1"
	"go.uber.org/zap"
	googleGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func main() {
	cfg := config.InitConfig()
	registryClient := grpcClient.NewRegistryClient(cfg)
	notificationClient := grpcClient.NewNotificationClient(registryClient)
	res, err := notificationClient.GetNotifications(&pbNotification.GetNotificationsRequest{UserId: "1"})
	if err != nil {
		log.Fatalf("❌ Failed to get notifications: %v", err)
	}
	log.Printf("Notifications: %v", res.Notifications)

	// Initialize logger
	logger, _ := zap.NewProduction()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// Initialize the database connection
	// db, err := repository.NewDB(
	// 	infraConf.DatabaseConfig{
	// 		Driver:                  "mysql",
	// 		Url:                     "user:password@tcp(127.0.0.1:3306)/your_database_name?charset=utf8mb4&parseTime=true&loc=UTC&tls=false&readTimeout=3s&writeTimeout=3s&timeout=3s&clientFoundRows=true",
	// 		IsSQLite:                false,
	// 		ConnMaxLifetimeInMinute: 3,
	// 		MaxOpenConns:            10,
	// 		MaxIdleConns:            1,
	// 	},
	// )
	db, err := repository.NewDB(
		infraConf.DatabaseConfig{
			Driver:   "sqlite3",
			Url:      ":memory:",
			IsSQLite: true,
		},
	)
	if err != nil {
		log.Fatalf("failed to new database err=%s\n", err.Error())
	}

	// Create redis
	rdb := repository.NewRedisClient("localhost:6379", "", 0)

	// Create the UserRepository
	userRepo := repository.NewUserRepository(db, rdb)

	// Create the UserService
	userService := service.NewUserService(userRepo)

	// Create the UserController
	userController := grpcCtrl.NewUserController(userService)
	coreController := grpcCtrl.NewCoreController(userService)

	// Create the gRPC server
	grpcServer, err := grpc.NewGrpcServer(
		config.GrpcServerConfig{
			Port: 9090,
			KeepaliveParams: keepalive.ServerParameters{
				MaxConnectionIdle:     100,
				MaxConnectionAge:      7200,
				MaxConnectionAgeGrace: 60,
				Time:                  10,
				Timeout:               3,
			},
			KeepalivePolicy: keepalive.EnforcementPolicy{
				MinTime:             10,
				PermitWithoutStream: true,
			},
		},
	)
	if err != nil {
		log.Fatalf("failed to new grpc server err=%s\n", err.Error())
	}

	// Start the gRPC server
	go grpcServer.Start(
		func(server *googleGrpc.Server) {
			pbCore.RegisterCoreServiceServer(server, coreController)
			pbCore.RegisterUserServiceServer(server, userController)
		},
	)

	// Add shutdown hook to trigger closer resources of service
	server.AddShutdownHook(grpcServer, db)
}
