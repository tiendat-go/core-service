package grpc

import (
	"context"

	"github.com/tiendat-go/core-service/internal/core/port/service"
	pbCore "github.com/tiendat-go/proto-service/gen/core/v1"
)

type coreController struct {
	pbCore.UnimplementedCoreServiceServer
	userService service.UserService
}

func NewCoreController(userService service.UserService) pbCore.CoreServiceServer {
	return &coreController{
		userService: userService,
	}
}

func (s *coreController) SayHello(ctx context.Context, req *pbCore.SayHelloRequest) (*pbCore.SayHelloResponse, error) {
	return &pbCore.SayHelloResponse{Message: "Hello " + req.Name}, nil
}
