package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	httpClient "github.com/tiendat-go/core-service/internal/client/http"

	"github.com/gin-gonic/gin"
	http2 "github.com/tiendat-go/core-service/internal/controller/http"
	"github.com/tiendat-go/core-service/internal/core/config"
	"github.com/tiendat-go/core-service/internal/core/server/http"
	"github.com/tiendat-go/core-service/internal/core/service"
	infraConf "github.com/tiendat-go/core-service/internal/infra/config"
	"github.com/tiendat-go/core-service/internal/infra/repository"
)

func main() {
	httpClient.NewRegistryClient("http://localhost:9999", "core-service", "9090")

	// Create a new instance of the Gin router
	instance := gin.New()
	instance.Use(gin.Recovery())

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
	userController := http2.NewUserController(instance, userService)

	// Initialize the routes for UserController
	userController.InitRouter()

	// Create the HTTP server
	httpServer := http.NewHttpServer(
		instance,
		config.HttpServerConfig{
			Port: 8000,
		},
	)

	// Start the HTTP server
	httpServer.Start()
	defer func(httpServer http.HttpServer) {
		err := httpServer.Close()
		if err != nil {
			log.Printf("failed to close http server %v", err)
		}
	}(httpServer)

	// Listen for OS signals to perform a graceful shutdown
	log.Println("listening signals...")
	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	<-c
	log.Println("graceful shutdown...")
}
