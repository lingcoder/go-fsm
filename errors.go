package fsm

import (
	"errors"
)

// Error constants - standard error definitions
var (
	ErrStateMachineNotFound     = errors.New("state machine not found")
	ErrStateMachineAlreadyExist = errors.New("state machine already exists")
	ErrStateNotFound            = errors.New("state not found")
	ErrTransitionNotFound       = errors.New("no transition found")
	ErrConditionNotMet          = errors.New("transition conditions not met")
	ErrActionExecutionFailed    = errors.New("action execution failed")
	ErrStateMachineNotReady     = errors.New("state machine is not ready yet")
	ErrInternalTransition       = errors.New("internal transition source and target states must be the same")
	ErrDuplicateTransition      = errors.New("duplicate transition detected")
	ErrNoTransitionsDefined     = errors.New("no transitions defined in state machine")
)
