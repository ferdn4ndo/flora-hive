package models

import (
	"fmt"
	"runtime"
)

// ErrorWrapped is used with Gin's error chain and Sentry middleware.
type ErrorWrapped struct {
	msg string
	err error
}

func (e ErrorWrapped) Error() string {
	return e.msg
}

func (e ErrorWrapped) Unwrap() error {
	return e.err
}

// CreateErrorWrapped wraps an error for controllers.
func CreateErrorWrapped(msg string, err error) error {
	return &ErrorWrapped{
		msg: msg,
		err: fmt.Errorf("%s - %w", getFrame(1).Function, err),
	}
}

// CreateErrorWithContext annotates an error with caller information.
func CreateErrorWithContext(err error) error {
	return fmt.Errorf("%s - %w", getFrame(1).Function, err)
}

func getFrame(skipFrames int) runtime.Frame {
	targetFrameIndex := skipFrames + 2
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)
	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}
	return frame
}

// ForbiddenError is returned when RBAC denies access.
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "forbidden"
}
