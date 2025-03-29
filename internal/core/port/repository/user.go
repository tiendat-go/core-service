package repository

import (
	"errors"

	"github.com/tiendat-go/core-service/internal/core/dto"
)

var (
	DuplicateUser = errors.New("duplicate user")
)

type UserRepository interface {
	Insert(user dto.UserDTO) error
}
