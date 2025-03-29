package service

import (
	"github.com/tiendat-go/core-service/internal/core/model/request"
	"github.com/tiendat-go/core-service/internal/core/model/response"
)

type UserService interface {
	SignUp(request *request.SignUpRequest) *response.Response
}
