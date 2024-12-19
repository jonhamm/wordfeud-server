package store

import (
	"errors"
	"fmt"
)

type StoreError struct {
	ErrorCode int
	Err       error
}

const (
	STORE_ERROR_NONE                = 0
	STORE_ERROR_UNEXPECTED          = 1
	STORE_ERROR_USER_NAME_SHORT     = 101
	STORE_ERROR_USER_PASSWORD_SHORT = 102
	STORE_ERROR_USER_NAME_EXISTS    = 103
)

func (serr *StoreError) Error() string {
	if serr == nil {
		return ""
	}
	return fmt.Sprintf("StoreError %d : %v", serr.ErrorCode, serr.Err)
}

func StoreErrorCode(err error) int {
	serr := AsStoreError(err)
	if serr != nil {
		return serr.ErrorCode
	}
	return 0
}

func AsStoreError(err error) *StoreError {
	if serr, ok := err.(*StoreError); ok {
		return serr
	}
	return nil
}

func NewStoreError(code int, text string) *StoreError {
	return &StoreError{
		ErrorCode: code,
		Err:       errors.New(text),
	}
}
