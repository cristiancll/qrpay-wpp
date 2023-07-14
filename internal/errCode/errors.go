package errCode

import (
	errs "github.com/cristiancll/go-errors"
	"google.golang.org/grpc/codes"
)

const (
	Unknown errs.Code = iota
	Internal
	NotFound
	NotChanged
	AccessDenied
	AlreadyExists
	Unauthorized
	Unauthenticated
	InvalidArgument
)

func ToGRPCCode(code errs.Code) codes.Code {
	switch code {
	case Internal:
		return codes.Internal
	case NotFound:
		return codes.NotFound
	case NotChanged:
		return codes.FailedPrecondition
	case AccessDenied:
		return codes.PermissionDenied
	case AlreadyExists:
		return codes.AlreadyExists
	case Unauthorized:
		return codes.Unauthenticated
	case Unauthenticated:
		return codes.Unauthenticated
	default:
		return codes.Unknown
	}
}
