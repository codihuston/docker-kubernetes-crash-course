package errors

import (
	"errors"
	"net/http"

	// "gorm.io/driver/postgres" provides and uses this pg driver. Gorm's
	// primitive error `ErrDuplicatedKey` does not appear to return true
	// during type checking ( postgres 'v1.5.0' (gorm.io/gorm v1.25.1)),
	// thus the impl below...
	// See: https://github.com/go-gorm/gorm/issues/4135

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	DUPLICATED_KEY = "23505"
)

var IsNotImplementedError = errors.New("Not Implemented")

// APIError is an error intended to be consumed by the error_handler middleware.
// When an error occurs across any layer, it should contain all of the
// information necessary to inform the client (and server-side log).
type APIError struct {
	Code    int // An HTTP Status Code to represent this error.
	err     error
	message string // A clean log message for the client-side.
}

func NewAPIError(err error) APIError {
	// Init a default error
	result := APIError{
		Code:    500,
		err:     err,
		message: http.StatusText(http.StatusInternalServerError),
	}

	// Then determine if, based on the message/type of error, the status needs
	// to be changed.
	return HandleDataNotFoundError(
		HandleDuplicateError(
			HandleIsNotImplementedError(result),
		),
	)
}

// isDuplicatedKeyError (and other error checking methods) directly check the
// underlying database client used by gorm for an error code.
func isDuplicatedKeyError(err error) bool {
	var perr *pgconn.PgError
	if errors.As(err, &perr) {
		return perr.Code == DUPLICATED_KEY
	}
	return false
}

func isNotFoundError(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func isNotImplementedError(err error) bool {
	return errors.Is(err, IsNotImplementedError)
}

// Error returns the message attached to the err.
func (e *APIError) Error() string {
	return e.err.Error()
}

// Returns the sanitized message body as an error object.
func (e *APIError) GetMessage() string {
	// return errors.New(e.message)
	return e.message
}

func (e *APIError) Unwrap() error {
	return e.err
}

func HandleDuplicateError(a APIError) APIError {
	if a.err == nil {
		return a
	}

	if isDuplicatedKeyError(a.err) {
		return APIError{
			Code:    http.StatusConflict,
			err:     a.err,
			message: http.StatusText(http.StatusConflict),
		}
	}

	return a
}

func HandleDataNotFoundError(a APIError) APIError {
	if a.err == nil {
		return a
	}

	if isNotFoundError(a.err) {
		return APIError{
			Code:    http.StatusNotFound,
			err:     a.err,
			message: http.StatusText(http.StatusNotFound),
		}
	}

	return a
}

func HandleIsNotImplementedError(a APIError) APIError {
	if a.err == nil {
		return a
	}

	if isNotImplementedError(a.err) {
		return APIError{
			Code:    http.StatusNotImplemented,
			err:     a.err,
			message: http.StatusText(http.StatusNotImplemented),
		}
	}

	return a
}
