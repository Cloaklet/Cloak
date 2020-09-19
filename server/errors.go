package server

import "fmt"

// ApiError is an custom implementation of `error` which provides simplified JSON representation.
type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (a *ApiError) Error() string {
	return fmt.Sprintf("ApiError code=%d, msg=%s", a.Code, a.Message)
}

// Reformat returns a new ApiError by formatting given values into its `Message` field.
// This is mainly for errors with placeholders in their message strings, e.g. `ErrUnknown`.
func (a *ApiError) Reformat(v ...interface{}) *ApiError {
	return &ApiError{
		Code:    a.Code,
		Message: fmt.Sprintf(a.Message, v...),
	}
}

// DataContainer is a container that wraps an ApiError with some data.
// Useful when we need to return some data alongside an error.
type DataContainer struct {
	*ApiError
	Item  interface{} `json:"item,omitempty"`
	Items interface{} `json:"items,omitempty"`
	State string      `json:"state,omitempty"`
}

func (a *ApiError) WrapList(items interface{}) *DataContainer {
	return &DataContainer{
		ApiError: a,
		Items:    items,
	}
}

func (a *ApiError) WrapItem(item interface{}) *DataContainer {
	return &DataContainer{
		ApiError: a,
		Item:     item,
	}
}

func (a *ApiError) WrapState(state string) *DataContainer {
	return &DataContainer{
		ApiError: a,
		State:    state,
	}
}

// Here is a complete list of API errors
var (
	ErrOk                     = &ApiError{Code: 0, Message: "Ok"}
	ErrListFailed             = &ApiError{Code: 1, Message: "Failed to list vaults"}
	ErrMalformedInput         = &ApiError{Code: 2, Message: "Malformed input data"}
	ErrUnknown                = &ApiError{Code: 3, Message: "Error: %v"}
	ErrPathNotExist           = &ApiError{Code: 4, Message: "Given path does not exist"}
	ErrUnsupportedOperation   = &ApiError{Code: 5, Message: "Unsupported operation"}
	ErrVaultNotExist          = &ApiError{Code: 6, Message: "Given vault ID does not exist"}
	ErrVaultAlreadyUnlocked   = &ApiError{Code: 7, Message: "This vault is already unlocked"}
	ErrVaultAlreadyLocked     = &ApiError{Code: 8, Message: "This vault is already locked"}
	ErrMountpointNotEmpty     = &ApiError{Code: 9, Message: "Mountpoint is not empty"}
	ErrWrongPassword          = &ApiError{Code: 10, Message: "Password incorrect"}
	ErrCantOpenVaultConf      = &ApiError{Code: 11, Message: "gocryptfs.conf could not be opened"}
	ErrMissingGocryptfsBinary = &ApiError{Code: 12, Message: "Cannot locate gocryptfs binary"}
	ErrMissingFuse            = &ApiError{Code: 13, Message: "FUSE is not available on this computer"}
	ErrVaultMkdirFailed       = &ApiError{Code: 14, Message: "Failed to create vault directory: %v"}
	ErrVaultDirNotEmpty       = &ApiError{Code: 15, Message: "New vault directory is not empty"}
	ErrVaultPasswordEmpty     = &ApiError{Code: 16, Message: "Password for the new vault is empty"}
	ErrVaultInitConfFailed    = &ApiError{Code: 17, Message: "Could not create gocryptfs.conf for the new vault"}
)
