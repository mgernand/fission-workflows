package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/go-openapi/errors"
)

// TaskInvocationStatus task invocation status
// swagger:model TaskInvocationStatus
type TaskInvocationStatus struct {

	// error
	Error *Error `json:"error,omitempty"`

	// output
	Output *TypedValue `json:"output,omitempty"`

	// status
	Status TaskInvocationStatusStatus `json:"status,omitempty"`

	// updated at
	UpdatedAt strfmt.DateTime `json:"updatedAt,omitempty"`
}

// Validate validates this task invocation status
func (m *TaskInvocationStatus) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateError(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateOutput(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *TaskInvocationStatus) validateError(formats strfmt.Registry) error {

	if swag.IsZero(m.Error) { // not required
		return nil
	}

	if m.Error != nil {

		if err := m.Error.Validate(formats); err != nil {
			return err
		}
	}

	return nil
}

func (m *TaskInvocationStatus) validateOutput(formats strfmt.Registry) error {

	if swag.IsZero(m.Output) { // not required
		return nil
	}

	if m.Output != nil {

		if err := m.Output.Validate(formats); err != nil {
			return err
		}
	}

	return nil
}

func (m *TaskInvocationStatus) validateStatus(formats strfmt.Registry) error {

	if swag.IsZero(m.Status) { // not required
		return nil
	}

	if err := m.Status.Validate(formats); err != nil {
		return err
	}

	return nil
}
