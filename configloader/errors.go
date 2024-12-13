package configloader

import (
	"fmt"
	"reflect"
	"strings"
)

// ConfigLoadError represents the errors that occurred during config loading.
type ConfigLoadError struct {
	Value  reflect.Type
	Errors []error
}

func (e *ConfigLoadError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

func (e *ConfigLoadError) Error() string {
	var msgs = make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("failed to load %s:\n%s", e.Value.String(),
		strings.Join(msgs, "\n"))
}

func (e *ConfigLoadError) Unwrap() []error {
	return e.Errors
}

// FieldError represents an error that occurred while processing a specific field.
type FieldError struct {
	Field
	Err error
}

func (e FieldError) Error() string {
	return fmt.Sprintf("error processing field %s: %v", e.Field.String(), e.Err)
}

func (e FieldError) Unwrap() error {
	return e.Err
}

// MissingEnvError represents an error when a required environment variable is not found.
type MissingEnvError struct {
	Key string
}

func (e MissingEnvError) Error() string {
	return fmt.Sprintf("environment variable %s not found", e.Key)
}

// UnsupportedTypeError represents an error when trying to process a field with an unsupported type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type %v", e.Type)
}
