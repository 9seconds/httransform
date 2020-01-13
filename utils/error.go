package utils

import (
	"errors"
	"fmt"
	"strings"
)

type Error struct {
	Text string

	errorRoot *Error
	parent    *Error
	cause     error
}

func (e *Error) Unwrap() error {
	return e.parent
}

func (e *Error) Error() string {
	messages := []string{}

	for current := e; current != nil; current = current.parent {
		messages = append(messages, current.String())
	}

	return strings.Join(messages, ": ")
}

func (e *Error) String() string {
	if e.cause != nil {
		return e.Text
	}

	return fmt.Errorf("%s (%w)", e.Text, e.cause).Error()
}

func (e *Error) Is(target error) bool {
	if converted, ok := target.(*Error); ok {
		if converted == e {
			return true
		}

		for current := e.errorRoot; current != nil; current = current.errorRoot {
			if current == converted {
				return true
			}
		}
	}

	return errors.Is(e.cause, target)
}

func (e *Error) As(target interface{}) bool {
	if converted, ok := target.(**Error); ok {
		*converted = e
		return true
	}

	return errors.As(e.cause, target)
}

func (e *Error) Wrap(text string, cause error) *Error {
	return &Error{
		Text:   text,
		parent: e,
		cause:  cause,
	}
}

func (e *Error) Extend(text string) *Error {
	return &Error{
		Text:      fmt.Sprintf("%s: %s", e.String(), text),
		errorRoot: e,
	}
}
