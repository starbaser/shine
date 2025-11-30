package rpc

import (
	"github.com/creachadair/jrpc2"
)

const (
	CodeParseError     = -32700 // Invalid JSON
	CodeInvalidRequest = -32600 // Not a valid request object
	CodeMethodNotFound = -32601 // Method does not exist
	CodeInvalidParams  = -32602 // Invalid method parameters
	CodeInternal       = -32603 // Internal error
)

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

func ErrPrismNotFound(name string) error {
	return jrpc2.Errorf(CodePrismNotFound, "prism not found: %s", name)
}

func ErrPrismNotRunning(name string) error {
	return jrpc2.Errorf(CodePrismNotRunning, "prism not running: %s", name)
}

func ErrPrismAlreadyUp(name string) error {
	return jrpc2.Errorf(CodePrismAlreadyUp, "prism already running: %s", name)
}

func ErrPanelNotFound(instance string) error {
	return jrpc2.Errorf(CodePanelNotFound, "panel not found: %s", instance)
}

func ErrShuttingDown() error {
	return jrpc2.Errorf(CodeShuttingDown, "service is shutting down")
}

func ErrConfig(msg string) error {
	return jrpc2.Errorf(CodeConfigError, "config error: %s", msg)
}

func ErrResourceBusy(resource string) error {
	return jrpc2.Errorf(CodeResourceBusy, "resource busy: %s", resource)
}

func ErrOperationFailed(op string, err error) error {
	return jrpc2.Errorf(CodeOperationFailed, "%s failed: %v", op, err)
}

func ErrInvalidParams(msg string) error {
	return jrpc2.Errorf(CodeInvalidParams, "invalid params: %s", msg)
}

func ErrInternal(err error) error {
	return jrpc2.Errorf(CodeInternal, "internal error: %v", err)
}

func ErrNotImplemented(method string) error {
	return jrpc2.Errorf(CodeNotImplemented, "method not implemented: %s", method)
}
