package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"

	"git.blender.org/flamenco/pkg/api"
)

// This file contains error types used in the rest of the Worker code.

// ParameterInvalidError is returned by command executors when a mandatory
// command parameter is missing.
type ParameterMissingError struct {
	Parameter string
	Cmd       api.Command
}

func NewParameterMissingError(parameter string, cmd api.Command) ParameterMissingError {
	return ParameterMissingError{
		Parameter: parameter,
		Cmd:       cmd,
	}
}

func (err ParameterMissingError) Error() string {
	return fmt.Sprintf("%s: mandatory parameter %q is missing: %+v",
		err.Cmd.Name, err.Parameter, err.Cmd.Parameters)
}

// ParameterInvalidError is returned by command executors when a command
// parameter is invalid.
type ParameterInvalidError struct {
	Parameter string
	Cmd       api.Command
	Message   string
}

func NewParameterInvalidError(parameter string, cmd api.Command, message string, fmtArgs ...interface{}) ParameterInvalidError {
	if len(fmtArgs) > 0 {
		message = fmt.Sprintf(message, fmtArgs...)
	}
	return ParameterInvalidError{
		Parameter: parameter,
		Cmd:       cmd,
		Message:   message,
	}
}

func (err ParameterInvalidError) Error() string {
	return fmt.Sprintf("%s: parameter %q has invalid value %+v: %s",
		err.Cmd.Name, err.Parameter, err.ParamValue(), err.Message)
}

// ParamValue returns the value of the invalid parameter.
func (err ParameterInvalidError) ParamValue() interface{} {
	return err.Cmd.Parameters[err.Parameter]
}

// Ensure the structs above adhere to the error interface.
var _ error = ParameterMissingError{}
var _ error = ParameterInvalidError{}
