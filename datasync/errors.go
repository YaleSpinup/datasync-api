package datasync

import (
	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func ErrCode(msg string, err error) error {
	if aerr, ok := errors.Cause(err).(awserr.Error); ok {
		switch aerr.Code() {
		case
			"Forbidden":

			return apierror.New(apierror.ErrForbidden, msg, aerr)
		case
			// Limit Exceeded
			"LimitExceeded":

			return apierror.New(apierror.ErrLimitExceeded, msg, aerr)
		case
			// Not found.
			"NotFound":

			return apierror.New(apierror.ErrNotFound, msg, aerr)
		case
			// ErrCodeInvalidRequestException for service response error code
			// "InvalidRequestException".
			//
			// This exception is thrown when the client submits a malformed request.
			datasync.ErrCodeInvalidRequestException:

			return apierror.New(apierror.ErrBadRequest, msg, err)
		case
			// ErrCodeInternalException for service response error code
			// "InternalException".
			//
			// This exception is thrown when an error occurs in the DataSync service.
			datasync.ErrCodeInternalException:

			return apierror.New(apierror.ErrInternalError, msg, err)
		case
			// Service Unavailable
			"ServiceUnavailable":

			return apierror.New(apierror.ErrServiceUnavailable, msg, aerr)
		default:
			return apierror.New(apierror.ErrBadRequest, msg, aerr)
		}
	}

	log.Warnf("uncaught error: %s, returning Internal Server Error", err)
	return apierror.New(apierror.ErrInternalError, msg, err)
}
