package common

import (
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
)

// NewManagementApiResponseError
func NewManagementApiResponseError(code, msg string, details ...string) *grpc.ResponseErrorMessage {
	var det string
	if len(details) > 0 && details[0] != "" {
		det = details[0]
	} else {
		det = ""
	}
	return &grpc.ResponseErrorMessage{Code: code, Message: msg, Details: det}
}

// NewValidationError
func NewValidationError(details string) *grpc.ResponseErrorMessage {
	return NewManagementApiResponseError(ErrorValidationFailed.Code, ErrorValidationFailed.Message, details)
}

var (
	ErrorUnknown                       = NewManagementApiResponseError("ma000001", "unknown error. try request later")
	ErrorValidationFailed              = NewManagementApiResponseError("ma000002", "validation failed")
	ErrorInternal                      = NewManagementApiResponseError("ma000003", InternalErrorTemplate)
	ErrorIncorrectOrderId              = NewManagementApiResponseError("ma000008", "incorrect order identifier")
	ErrorMessageSignatureHeaderIsEmpty = NewManagementApiResponseError("ma000022", "header with request signature can't be empty")
	ErrorRequestParamsIncorrect        = NewManagementApiResponseError("ma000023", "incorrect request parameters")
	ErrorRequestDataInvalid            = NewManagementApiResponseError("ma000026", "request data invalid")
	ErrorMessageIncorrectZip           = NewManagementApiResponseError("ma000073", "incorrect zip code")

	ValidationErrors = map[string]*grpc.ResponseErrorMessage{
		ValidationParameterOrderId:   ErrorIncorrectOrderId,
		ValidationParameterOrderUuid: ErrorIncorrectOrderId,
	}
)
