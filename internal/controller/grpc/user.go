package grpc

import (
	"context"

	pbCore "github.com/tiendat-go/proto-service/gen/core/v1"

	"github.com/tiendat-go/core-service/internal/core/entity/error_code"
	"github.com/tiendat-go/core-service/internal/core/model/request"
	"github.com/tiendat-go/core-service/internal/core/model/response"
	"github.com/tiendat-go/core-service/internal/core/port/service"
)

var errorCodeMapper = map[error_code.ErrorCode]pbCore.ErrorCode{
	error_code.Success:        pbCore.ErrorCode_ERROR_CODE_SUCCESS,
	error_code.InternalError:  pbCore.ErrorCode_ERROR_CODE_EC_UNSPECIFIED,
	error_code.InvalidRequest: pbCore.ErrorCode_ERROR_CODE_INVALID_REQUEST,
	error_code.DuplicateUser:  pbCore.ErrorCode_ERROR_CODE_DUPLICATE_USER,
}

type userController struct {
	pbCore.UnimplementedUserServiceServer
	userService service.UserService
}

func NewUserController(userService service.UserService) pbCore.UserServiceServer {
	return &userController{
		userService: userService,
	}
}

func (u userController) SignUp(
	ctx context.Context, request *pbCore.SignUpRequest,
) (*pbCore.SignUpResponse, error) {
	resp := u.userService.SignUp(u.newSignUpRequest(request))
	return u.newSignUpResponse(resp)
}

func (u userController) newSignUpRequest(protoRequest *pbCore.SignUpRequest) *request.SignUpRequest {
	return &request.SignUpRequest{
		Username: protoRequest.GetUserName(),
		Password: protoRequest.GetPassword(),
	}
}

func (u userController) newSignUpResponse(resp *response.Response) (
	*pbCore.SignUpResponse, error,
) {
	if !resp.Status {
		return &pbCore.SignUpResponse{
			Status:       resp.Status,
			ErrorCode:    u.mapErrorCode(resp.ErrorCode),
			ErrorMessage: resp.ErrorMessage,
		}, nil
	}

	data := resp.Data.(response.SignUpDataResponse)
	return &pbCore.SignUpResponse{
		Status:       resp.Status,
		ErrorCode:    u.mapErrorCode(resp.ErrorCode),
		ErrorMessage: resp.ErrorMessage,
		DisplayName:  data.DisplayName,
	}, nil
}

func (u userController) mapErrorCode(errCode error_code.ErrorCode) pbCore.ErrorCode {
	code, existed := errorCodeMapper[errCode]
	if existed {
		return code
	}

	return pbCore.ErrorCode_ERROR_CODE_EC_UNSPECIFIED
}
