package repository

import (
	"errors"
	"strings"

	"github.com/tiendat-go/core-service/internal/core/dto"
	"github.com/tiendat-go/core-service/internal/core/port/repository"
)

const (
	duplicateEntryMsg = "Duplicate entry"
	numberRowInserted = 1
)

var (
	insertUserErr = errors.New("failed to insert user")
)

const (
	insertUserStatement = "INSERT INTO User ( " +
		"`username`, " +
		"`password`, " +
		"`display_name`, " +
		"`created_at`," +
		"`updated_at`) " +
		"VALUES (?, ?, ?, ?, ?)"
)

type userRepository struct {
	db  repository.Database
	rdb *Redis
}

func NewUserRepository(db repository.Database, rdb *Redis) repository.UserRepository {
	return &userRepository{
		db:  db,
		rdb: rdb,
	}
}

func (u userRepository) Insert(user dto.UserDTO) error {
	result, err := u.db.GetDB().Exec(
		insertUserStatement,
		user.UserName,
		user.Password,
		user.DisplayName,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), duplicateEntryMsg) {
			return repository.DuplicateUser
		}

		return err
	}

	numRow, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if numRow != numberRowInserted {
		return insertUserErr
	}

	return nil
}
