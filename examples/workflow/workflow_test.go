package workflow

import (
	"fmt"
	"testing"
	"time"

	"github.com/lingcoder/go-fsm"
)

// Define approval workflow states
type ApprovalState string

const (
	Draft     ApprovalState = "DRAFT"
	Submitted ApprovalState = "SUBMITTED"
	InReview  ApprovalState = "IN_REVIEW"
	Approved  ApprovalState = "APPROVED"
	Rejected  ApprovalState = "REJECTED"
	Cancelled ApprovalState = "CANCELLED"
)

// Define approval workflow events
type ApprovalEvent string

const (
	Submit   ApprovalEvent = "SUBMIT"
	Review   ApprovalEvent = "REVIEW"
	Approve  ApprovalEvent = "APPROVE"
	Reject   ApprovalEvent = "REJECT"
	Cancel   ApprovalEvent = "CANCEL"
	Resubmit ApprovalEvent = "RESUBMIT"
)

// Approval workflow payload
type ApprovalPayload struct {
	DocumentID    string
	Requester     string
	Reviewer      string
	Comments      string
	SubmittedAt   time.Time
	LastUpdatedAt time.Time
}

// Submission condition
type SubmitCondition struct{}

func (c *SubmitCondition) IsSatisfied(payload ApprovalPayload) bool {
	// Check if document is complete
	return payload.DocumentID != "" && payload.Requester != ""
}

// Reviewer check condition
type ReviewerCondition struct{}

func (c *ReviewerCondition) IsSatisfied(payload ApprovalPayload) bool {
	// Check if reviewer is assigned
	return payload.Reviewer != ""
}

// Approval action
type ApprovalAction struct {
	ActionLog []string
}

func (a *ApprovalAction) Execute(from, to ApprovalState, event ApprovalEvent, payload ApprovalPayload) error {
	// Log the transition for verification in tests
	a.ActionLog = append(a.ActionLog, fmt.Sprintf("%s -> %s [%s]", from, to, event))
	return nil
}

// setupApprovalWorkflow creates and configures the approval workflow state machine
func setupApprovalWorkflow(t *testing.T) (fsm.StateMachine[ApprovalState, ApprovalEvent, ApprovalPayload], *ApprovalAction) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[ApprovalState, ApprovalEvent, ApprovalPayload]()

	// Create conditions and actions
	submitCondition := &SubmitCondition{}
	reviewerCondition := &ReviewerCondition{}
	approvalAction := &ApprovalAction{
		ActionLog: make([]string, 0),
	}

	// Draft to Submitted
	builder.ExternalTransition().
		From(Draft).
		To(Submitted).
		On(Submit).
		When(submitCondition).
		Perform(approvalAction)

	// Submitted to InReview
	builder.ExternalTransition().
		From(Submitted).
		To(InReview).
		On(Review).
		When(reviewerCondition).
		Perform(approvalAction)

	// InReview to Approved
	builder.ExternalTransition().
		From(InReview).
		To(Approved).
		On(Approve).
		When(fsm.ConditionFunc[ApprovalPayload](func(payload ApprovalPayload) bool {
			return true
		})).
		Perform(approvalAction)

	// InReview to Rejected
	builder.ExternalTransition().
		From(InReview).
		To(Rejected).
		On(Reject).
		When(fsm.ConditionFunc[ApprovalPayload](func(payload ApprovalPayload) bool {
			return true
		})).
		Perform(approvalAction)

	// Rejected to Submitted
	builder.ExternalTransition().
		From(Rejected).
		To(Submitted).
		On(Resubmit).
		When(submitCondition).
		Perform(approvalAction)

	// Cancel from multiple states
	builder.ExternalTransitions().
		FromAmong(Draft, Submitted, InReview).
		To(Cancelled).
		On(Cancel).
		When(fsm.ConditionFunc[ApprovalPayload](func(payload ApprovalPayload) bool {
			return true
		})).
		Perform(approvalAction)

	// Build the state machine
	stateMachine, err := builder.Build("ApprovalWorkflow")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	return stateMachine, approvalAction
}

// createValidPayload creates a valid payload for testing
func createValidPayload() ApprovalPayload {
	return ApprovalPayload{
		DocumentID:    "DOC-20250425-001",
		Requester:     "John Doe",
		Reviewer:      "Jane Smith",
		SubmittedAt:   time.Now(),
		LastUpdatedAt: time.Now(),
	}
}

// createInvalidPayload creates an invalid payload (missing reviewer) for testing
func createInvalidPayload() ApprovalPayload {
	return ApprovalPayload{
		DocumentID:    "DOC-20250425-001",
		Requester:     "John Doe",
		Reviewer:      "", // Missing reviewer
		SubmittedAt:   time.Now(),
		LastUpdatedAt: time.Now(),
	}
}

// TestApprovalWorkflow tests the basic functionality of an approval workflow state machine
func TestApprovalWorkflow(t *testing.T) {
	stateMachine, action := setupApprovalWorkflow(t)
	fmt.Println(stateMachine.GenerateDiagram(fsm.PlantUML, fsm.MarkdownTable, fsm.MarkdownFlowchart, fsm.MarkdownStateDiagram))
	// Define test cases
	testCases := []struct {
		name           string
		initialState   ApprovalState
		events         []ApprovalEvent
		payload        ApprovalPayload
		expectedStates []ApprovalState
		shouldError    bool
		errorIndex     int // Which event should cause an error (if shouldError is true)
	}{
		{
			name:           "Happy Path",
			initialState:   Draft,
			events:         []ApprovalEvent{Submit, Review, Approve},
			payload:        createValidPayload(),
			expectedStates: []ApprovalState{Submitted, InReview, Approved},
			shouldError:    false,
		},
		{
			name:           "Rejection Path",
			initialState:   Draft,
			events:         []ApprovalEvent{Submit, Review, Reject, Resubmit, Review, Approve},
			payload:        createValidPayload(),
			expectedStates: []ApprovalState{Submitted, InReview, Rejected, Submitted, InReview, Approved},
			shouldError:    false,
		},
		{
			name:           "Cancellation Path",
			initialState:   Draft,
			events:         []ApprovalEvent{Submit, Cancel},
			payload:        createValidPayload(),
			expectedStates: []ApprovalState{Submitted, Cancelled},
			shouldError:    false,
		},
		{
			name:           "Missing Reviewer Error",
			initialState:   Draft,
			events:         []ApprovalEvent{Submit, Review},
			payload:        createInvalidPayload(),
			expectedStates: []ApprovalState{Submitted},
			shouldError:    true,
			errorIndex:     1, // Review event should fail
		},
		{
			name:           "Invalid Transition Error",
			initialState:   Draft,
			events:         []ApprovalEvent{Approve}, // Cannot approve from Draft
			payload:        createValidPayload(),
			expectedStates: []ApprovalState{},
			shouldError:    true,
			errorIndex:     0,
		},
		{
			name:           "Multiple Cancellations",
			initialState:   Draft,
			events:         []ApprovalEvent{Cancel, Submit}, // Second event should fail
			payload:        createValidPayload(),
			expectedStates: []ApprovalState{Cancelled},
			shouldError:    true,
			errorIndex:     1,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset action log
			action.ActionLog = make([]string, 0)

			currentState := tc.initialState
			var err error

			// Process each event
			for i, event := range tc.events {
				currentState, err = stateMachine.FireEvent(currentState, event, tc.payload)

				// Check for expected errors
				if tc.shouldError && i == tc.errorIndex {
					if err == nil {
						t.Fatalf("Expected error for event %s but got none", event)
					}
					return // Stop processing events after expected error
				} else if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Verify state transition
				if i < len(tc.expectedStates) && currentState != tc.expectedStates[i] {
					t.Errorf("Expected state to be %s after event %s, got %s",
						tc.expectedStates[i], event, currentState)
				}
			}

			// Verify action log length matches successful transitions
			expectedLogLength := len(tc.events)
			if tc.shouldError {
				expectedLogLength = tc.errorIndex
			}
			if len(action.ActionLog) != expectedLogLength {
				t.Errorf("Expected %d actions to be logged, got %d",
					expectedLogLength, len(action.ActionLog))
			}
		})
	}
}

// TestInternalTransition tests internal transitions within the same state
func TestInternalTransition(t *testing.T) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[ApprovalState, ApprovalEvent, ApprovalPayload]()

	// Track transitions
	transitionLog := make([]string, 0)

	// Define internal transition
	builder.InternalTransition().
		Within(InReview).
		On(Review). // Re-review
		When(fsm.ConditionFunc[ApprovalPayload](func(payload ApprovalPayload) bool {
			return true
		})).
		PerformFunc(func(from, to ApprovalState, event ApprovalEvent, payload ApprovalPayload) error {
			transitionLog = append(transitionLog, fmt.Sprintf("Internal: %s [%s]", from, event))
			return nil
		})

	// Build the state machine
	stateMachine, err := builder.Build("InternalTransitionTest")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test internal transition
	payload := createValidPayload()
	state, err := stateMachine.FireEvent(InReview, Review, payload)
	if err != nil {
		t.Fatalf("Failed to execute internal transition: %v", err)
	}

	// State should remain the same
	if state != InReview {
		t.Errorf("Expected state to remain %s, got %s", InReview, state)
	}

	// Action should be executed
	if len(transitionLog) != 1 {
		t.Errorf("Expected 1 action to be logged, got %d", len(transitionLog))
	}
}

// TestParallelTransitions tests parallel transitions from one state to multiple states
func TestParallelTransitions(t *testing.T) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[ApprovalState, ApprovalEvent, ApprovalPayload]()

	// Track transitions
	transitionLog := make([]string, 0)

	// Define parallel transitions
	builder.ExternalParallelTransition().
		From(Draft).
		ToAmong(Submitted, InReview). // Unusual but for testing purposes
		On(Submit).
		When(fsm.ConditionFunc[ApprovalPayload](func(payload ApprovalPayload) bool {
			return true
		})).
		PerformFunc(func(from, to ApprovalState, event ApprovalEvent, payload ApprovalPayload) error {
			transitionLog = append(transitionLog, fmt.Sprintf("%s -> %s [%s]", from, to, event))
			return nil
		})

	// Build the state machine
	stateMachine, err := builder.Build("ParallelTransitionTest")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test parallel transition
	payload := createValidPayload()
	state, err := stateMachine.FireEvent(Draft, Submit, payload)
	if err != nil {
		t.Fatalf("Failed to execute parallel transition: %v", err)
	}

	// Check that we transitioned to one of the target states
	if state != Submitted && state != InReview {
		t.Errorf("Expected state to be either %s or %s, got %s", Submitted, InReview, state)
	}

	// In the current implementation, parallel transitions only log one transition
	// This is because the state machine only returns one final state
	if len(transitionLog) != 1 {
		t.Errorf("Expected 1 transition to be logged, got %d", len(transitionLog))
	}
}

// TestMultipleTransitionsBuilder tests the ExternalTransitions builder
func TestMultipleTransitionsBuilder(t *testing.T) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[ApprovalState, ApprovalEvent, ApprovalPayload]()

	// Track transitions
	transitionLog := make([]string, 0)

	// Define multiple source transitions
	builder.ExternalTransitions().
		FromAmong(Draft, Rejected, Cancelled).
		To(Submitted).
		On(Submit).
		When(fsm.ConditionFunc[ApprovalPayload](func(payload ApprovalPayload) bool {
			return true
		})).
		PerformFunc(func(from, to ApprovalState, event ApprovalEvent, payload ApprovalPayload) error {
			transitionLog = append(transitionLog, fmt.Sprintf("%s -> %s [%s]", from, to, event))
			return nil
		})

	// Build the state machine
	stateMachine, err := builder.Build("MultipleTransitionsTest")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test transitions from different source states
	payload := createValidPayload()

	// From Draft
	state, err := stateMachine.FireEvent(Draft, Submit, payload)
	if err != nil {
		t.Fatalf("Failed to transition from Draft: %v", err)
	}
	if state != Submitted {
		t.Errorf("Expected state to be %s, got %s", Submitted, state)
	}

	// From Rejected
	state, err = stateMachine.FireEvent(Rejected, Submit, payload)
	if err != nil {
		t.Fatalf("Failed to transition from Rejected: %v", err)
	}
	if state != Submitted {
		t.Errorf("Expected state to be %s, got %s", Submitted, state)
	}

	// From Cancelled
	state, err = stateMachine.FireEvent(Cancelled, Submit, payload)
	if err != nil {
		t.Fatalf("Failed to transition from Cancelled: %v", err)
	}
	if state != Submitted {
		t.Errorf("Expected state to be %s, got %s", Submitted, state)
	}

	// Check that we logged all transitions
	if len(transitionLog) != 3 {
		t.Errorf("Expected 3 transitions to be logged, got %d", len(transitionLog))
	}
}
