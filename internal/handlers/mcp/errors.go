package mcp

import "fmt"

// OllamaClientError represents an error from the Ollama client
type OllamaClientError struct {
	Operation string
	Err       error
}

func (e *OllamaClientError) Error() string {
	return fmt.Sprintf("ollama %s error: %v", e.Operation, e.Err)
}

func NewOllamaClientError(operation string, err error) *OllamaClientError {
	return &OllamaClientError{
		Operation: operation,
		Err:       err,
	}
}

// InvalidParameterError represents an invalid parameter error
type InvalidParameterError struct {
	Parameter string
	Reason    string
}

func (e *InvalidParameterError) Error() string {
	return fmt.Sprintf("invalid parameter %s: %s", e.Parameter, e.Reason)
}

func NewInvalidParameterError(parameter, reason string) *InvalidParameterError {
	return &InvalidParameterError{
		Parameter: parameter,
		Reason:    reason,
	}
}

// MissingParameterError represents a missing required parameter error
type MissingParameterError struct {
	Parameter string
}

func (e *MissingParameterError) Error() string {
	return fmt.Sprintf("missing required parameter: %s", e.Parameter)
}

func NewMissingParameterError(parameter string) *MissingParameterError {
	return &MissingParameterError{
		Parameter: parameter,
	}
}
