package v2

import (
	"fmt"
	"strings"
)

// ErrorCode represents the error type. The errors are serialized via strings
// and the integer format may change and should *never* be exported.
type ErrorCode int

const (
	// ErrorCodeUnknown is a catch-all for errors not defined below.
	ErrorCodeUnknown ErrorCode = iota

	// ErrorCodeUnsupported is returned when an operation is not supported.
	ErrorCodeUnsupported

	// ErrorCodeUnauthorized is returned if a request is not authorized.
	ErrorCodeUnauthorized

	// ErrorCodeDisabled is returned when a feature has been disabled
	ErrorCodeDisabled

	// ErrorCodeDigestInvalid is returned when uploading a blob if the
	// provided digest does not match the blob contents.
	ErrorCodeDigestInvalid

	// ErrorCodeSizeInvalid is returned when uploading a blob if the provided
	// size does not match the content length.
	ErrorCodeSizeInvalid

	// ErrorCodeNameInvalid is returned when the name in the manifest does not
	// match the provided name.
	ErrorCodeNameInvalid

	// ErrorCodeTagInvalid is returned when the tag in the manifest does not
	// match the provided tag.
	ErrorCodeTagInvalid

	// ErrorCodeNameUnknown when the repository name is not known.
	ErrorCodeNameUnknown

	// ErrorCodeManifestUnknown returned when image manifest is unknown.
	ErrorCodeManifestUnknown

	// ErrorCodeManifestInvalid returned when an image manifest is invalid,
	// typically during a PUT operation. This error encompasses all errors
	// encountered during manifest validation that aren't signature errors.
	ErrorCodeManifestInvalid

	// ErrorCodeManifestUnverified is returned when the manifest fails
	// signature verfication.
	ErrorCodeManifestUnverified

	// ErrorCodeBlobUnknown is returned when a blob is unknown to the
	// registry. This can happen when the manifest references a nonexistent
	// layer or the result is not found by a blob fetch.
	ErrorCodeBlobUnknown

	// ErrorCodeBlobUploadUnknown is returned when an upload is unknown.
	ErrorCodeBlobUploadUnknown

	// ErrorCodeBlobUploadInvalid is returned when an upload is invalid.
	ErrorCodeBlobUploadInvalid
)

// ParseErrorCode attempts to parse the error code string, returning
// ErrorCodeUnknown if the error is not known.
func ParseErrorCode(s string) ErrorCode {
	desc, ok := idToDescriptors[s]

	if !ok {
		return ErrorCodeUnknown
	}

	return desc.Code
}

// Descriptor returns the descriptor for the error code.
func (ec ErrorCode) Descriptor() ErrorDescriptor {
	d, ok := errorCodeToDescriptors[ec]

	if !ok {
		return ErrorCodeUnknown.Descriptor()
	}

	return d
}

// String returns the canonical identifier for this error code.
func (ec ErrorCode) String() string {
	return ec.Descriptor().Value
}

// Message returned the human-readable error message for this error code.
func (ec ErrorCode) Message() string {
	return ec.Descriptor().Message
}

// MarshalText encodes the receiver into UTF-8-encoded text and returns the
// result.
func (ec ErrorCode) MarshalText() (text []byte, err error) {
	return []byte(ec.String()), nil
}

// UnmarshalText decodes the form generated by MarshalText.
func (ec *ErrorCode) UnmarshalText(text []byte) error {
	desc, ok := idToDescriptors[string(text)]

	if !ok {
		desc = ErrorCodeUnknown.Descriptor()
	}

	*ec = desc.Code

	return nil
}

// Error provides a wrapper around ErrorCode with extra Details provided.
type Error struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message,omitempty"`
	Detail  interface{} `json:"detail,omitempty"`
}

// Error returns a human readable representation of the error.
func (e Error) Error() string {
	return fmt.Sprintf("%s: %s",
		strings.ToLower(strings.Replace(e.Code.String(), "_", " ", -1)),
		e.Message)
}

// Errors provides the envelope for multiple errors and a few sugar methods
// for use within the application.
type Errors struct {
	Errors []Error `json:"errors,omitempty"`
}

// Push pushes an error on to the error stack, with the optional detail
// argument. It is a programming error (ie panic) to push more than one
// detail at a time.
func (errs *Errors) Push(code ErrorCode, details ...interface{}) {
	if len(details) > 1 {
		panic("please specify zero or one detail items for this error")
	}

	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	if err, ok := detail.(error); ok {
		detail = err.Error()
	}

	errs.PushErr(Error{
		Code:    code,
		Message: code.Message(),
		Detail:  detail,
	})
}

// PushErr pushes an error interface onto the error stack.
func (errs *Errors) PushErr(err error) {
	switch err.(type) {
	case Error:
		errs.Errors = append(errs.Errors, err.(Error))
	default:
		errs.Errors = append(errs.Errors, Error{Message: err.Error()})
	}
}

func (errs *Errors) Error() string {
	switch errs.Len() {
	case 0:
		return "<nil>"
	case 1:
		return errs.Errors[0].Error()
	default:
		msg := "errors:\n"
		for _, err := range errs.Errors {
			msg += err.Error() + "\n"
		}
		return msg
	}
}

// Clear clears the errors.
func (errs *Errors) Clear() {
	errs.Errors = errs.Errors[:0]
}

// Len returns the current number of errors.
func (errs *Errors) Len() int {
	return len(errs.Errors)
}
