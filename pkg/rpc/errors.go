package rpc

import (
	"github.com/creachadair/jrpc2"
)

// Standard JSON-RPC 2.0 error codes
const (
	CodeParseError     = -32700 // Invalid JSON
	CodeInvalidRequest = -32600 // Not a valid request object
	CodeMethodNotFound = -32601 // Method does not exist
	CodeInvalidParams  = -32602 // Invalid method parameters
	CodeInternal       = -32603 // Internal error
)

// Application-specific error codes (-32000 to -32099 reserved for server)
const (
	CodePrismNotFound    = -32001 // Prism does not exist
	CodePrismNotRunning  = -32002 // Prism is not running
	CodePrismAlreadyUp   = -32003 // Prism is already running
	CodePanelNotFound    = -32004 // Panel does not exist
	CodeShuttingDown     = -32005 // Service is shutting down
	CodeConfigError      = -32006 // Configuration error
	CodeResourceBusy     = -32007 // Resource is busy
	CodeOperationFailed  = -32008 // Operation failed
	CodeNotImplemented   = -32009 // Method not implemented
)

// ErrPrismNotFound returns an error for unknown prism
func ErrPrismNotFound(name string) error {
	return jrpc2.Errorf(CodePrismNotFound, "prism not found: %s", name)
}

// ErrPrismNotRunning returns an error for prism that isn't running
func ErrPrismNotRunning(name string) error {
	return jrpc2.Errorf(CodePrismNotRunning, "prism not running: %s", name)
}

// ErrPrismAlreadyUp returns an error for prism that's already running
func ErrPrismAlreadyUp(name string) error {
	return jrpc2.Errorf(CodePrismAlreadyUp, "prism already running: %s", name)
}

// ErrPanelNotFound returns an error for unknown panel
func ErrPanelNotFound(instance string) error {
	return jrpc2.Errorf(CodePanelNotFound, "panel not found: %s", instance)
}

// ErrShuttingDown returns an error when service is shutting down
func ErrShuttingDown() error {
	return jrpc2.Errorf(CodeShuttingDown, "service is shutting down")
}

// ErrConfig returns a configuration error
func ErrConfig(msg string) error {
	return jrpc2.Errorf(CodeConfigError, "config error: %s", msg)
}

// ErrResourceBusy returns an error when resource is busy
func ErrResourceBusy(resource string) error {
	return jrpc2.Errorf(CodeResourceBusy, "resource busy: %s", resource)
}

// ErrOperationFailed returns an error when operation fails
func ErrOperationFailed(op string, err error) error {
	return jrpc2.Errorf(CodeOperationFailed, "%s failed: %v", op, err)
}

// ErrInvalidParams returns an error for invalid parameters
func ErrInvalidParams(msg string) error {
	return jrpc2.Errorf(CodeInvalidParams, "invalid params: %s", msg)
}

// ErrInternal returns an internal error
func ErrInternal(err error) error {
	return jrpc2.Errorf(CodeInternal, "internal error: %v", err)
}

// ErrNotImplemented returns an error for unimplemented methods
func ErrNotImplemented(method string) error {
	return jrpc2.Errorf(CodeNotImplemented, "method not implemented: %s", method)
}
