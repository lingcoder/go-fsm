package fsm

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// Define test states, events and payload
type testState string
type testEvent string
type testPayload struct {
	Value string
}

const (
	StateA testState = "A"
	StateB testState = "B"
	StateC testState = "C"
	StateD testState = "D"
)

const (
	Event1 testEvent = "Event1"
	Event2 testEvent = "Event2"
	Event3 testEvent = "Event3"
)

// Simple condition that always returns true
type alwaysTrueCondition struct{}

func (c *alwaysTrueCondition) IsSatisfied(payload testPayload) bool {
	return true
}

// Simple action that does nothing
type noopAction struct{}

func (a *noopAction) Execute(from, to testState, event testEvent, payload testPayload) error {
	return nil
}

// Create a test state machine
func createTestStateMachine(tb testing.TB) StateMachine[testState, testEvent, testPayload] {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	// Define state transitions
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateB).
		To(StateC).
		On(Event2).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateC).
		To(StateA).
		On(Event3).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	// Build the state machine
	sm, err := builder.Build("TestStateMachine")
	if err != nil {
		tb.Fatalf("Failed to build state machine: %v", err)
	}

	return sm
}

// Benchmark: Single-thread state transition
func BenchmarkSingleThreadTransition(b *testing.B) {
	sm := createTestStateMachine(b)
	payload := testPayload{Value: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentState := StateA

		newState, err := sm.FireEvent(currentState, Event1, payload)
		if err != nil {
			b.Fatalf("Failed to fire event: %v", err)
		}
		currentState = newState

		newState, err = sm.FireEvent(currentState, Event2, payload)
		if err != nil {
			b.Fatalf("Failed to fire event: %v", err)
		}
		currentState = newState

		newState, err = sm.FireEvent(currentState, Event3, payload)
		if err != nil {
			b.Fatalf("Failed to fire event: %v", err)
		}
	}
}

// Benchmark: Multi-thread state transition
func BenchmarkMultiThreadTransition(b *testing.B) {
	sm := createTestStateMachine(b)
	payload := testPayload{Value: "test"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			currentState := StateA

			newState, err := sm.FireEvent(currentState, Event1, payload)
			if err != nil {
				b.Fatalf("Failed to fire event: %v", err)
			}
			currentState = newState

			newState, err = sm.FireEvent(currentState, Event2, payload)
			if err != nil {
				b.Fatalf("Failed to fire event: %v", err)
			}
			currentState = newState

			newState, err = sm.FireEvent(currentState, Event3, payload)
			if err != nil {
				b.Fatalf("Failed to fire event: %v", err)
			}
		}
	})
}

// Benchmark: State machine creation
func BenchmarkStateMachineCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

		builder.ExternalTransition().
			From(StateA).
			To(StateB).
			On(Event1).
			When(&alwaysTrueCondition{}).
			Perform(&noopAction{})

		_, err := builder.Build("TestMachine")
		if err != nil {
			b.Fatalf("Failed to build state machine: %v", err)
		}
	}
}

// Benchmark: Large state machine with many states and transitions
func BenchmarkLargeStateMachine(b *testing.B) {
	// Create a state machine with 100 states and transitions
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	const numStates = 100
	states := make([]testState, numStates)
	for i := 0; i < numStates; i++ {
		states[i] = testState(fmt.Sprintf("State%d", i))
	}

	// Create chain transitions: State0 -> State1 -> ... -> State99 -> State0
	for i := 0; i < numStates; i++ {
		from := states[i]
		to := states[(i+1)%numStates]

		builder.ExternalTransition().
			From(from).
			To(to).
			On(Event1).
			When(&alwaysTrueCondition{}).
			Perform(&noopAction{})
	}

	sm, err := builder.Build("LargeStateMachine")
	if err != nil {
		b.Fatalf("Failed to build large state machine: %v", err)
	}

	payload := testPayload{Value: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentState := states[i%numStates]
		newState, err := sm.FireEvent(currentState, Event1, payload)
		if err != nil {
			b.Fatalf("Failed to fire event: %v", err)
		}
		_ = newState
	}
}

// Benchmark: Concurrent state machine access
func BenchmarkConcurrentStateMachineAccess(b *testing.B) {
	sm := createTestStateMachine(b)
	payload := testPayload{Value: "test"}

	var wg sync.WaitGroup
	numGoroutines := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(numGoroutines)

		for j := 0; j < numGoroutines; j++ {
			go func(j int) {
				defer wg.Done()

				// Each goroutine executes a different state transition
				currentState := StateA
				event := Event1

				if j%3 == 1 {
					currentState = StateB
					event = Event2
				} else if j%3 == 2 {
					currentState = StateC
					event = Event3
				}

				_, err := sm.FireEvent(currentState, event, payload)
				if err != nil {
					// Ignore errors, as some are expected in concurrent testing
					// For example, when state doesn't match the event
				}
			}(j)
		}

		wg.Wait()
	}
}

// ==================== Unit Tests ====================

// Condition that checks payload value
type valueCondition struct {
	expectedValue string
}

func (c *valueCondition) IsSatisfied(payload testPayload) bool {
	return payload.Value == c.expectedValue
}

// Action that returns an error
type errorAction struct{}

func (a *errorAction) Execute(from, to testState, event testEvent, payload testPayload) error {
	return errors.New("action failed")
}

// Action that records execution
type recordAction struct {
	executed bool
	from     testState
	to       testState
	event    testEvent
}

func (a *recordAction) Execute(from, to testState, event testEvent, payload testPayload) error {
	a.executed = true
	a.from = from
	a.to = to
	a.event = event
	return nil
}

// TestFireEvent_Success tests successful state transition
func TestFireEvent_Success(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()
	action := &recordAction{}

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(action)

	sm, err := builder.Build("TestFireEvent_Success")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newState, err := sm.FireEvent(StateA, Event1, testPayload{Value: "test"})
	if err != nil {
		t.Fatalf("FireEvent failed: %v", err)
	}

	if newState != StateB {
		t.Errorf("Expected state %v, got %v", StateB, newState)
	}

	if !action.executed {
		t.Error("Action was not executed")
	}

	if action.from != StateA || action.to != StateB || action.event != Event1 {
		t.Errorf("Action received wrong parameters: from=%v, to=%v, event=%v",
			action.from, action.to, action.event)
	}
}

// TestFireEvent_StateNotFound tests transition from non-existent state
func TestFireEvent_StateNotFound(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireEvent_StateNotFound")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	_, err = sm.FireEvent(StateD, Event1, testPayload{})
	if !errors.Is(err, ErrStateNotFound) {
		t.Errorf("Expected ErrStateNotFound, got %v", err)
	}
}

// TestFireEvent_TransitionNotFound tests firing event with no matching transition
func TestFireEvent_TransitionNotFound(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireEvent_TransitionNotFound")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Fire Event2 from StateA, but only Event1 is defined
	_, err = sm.FireEvent(StateA, Event2, testPayload{})
	if !errors.Is(err, ErrTransitionNotFound) {
		t.Errorf("Expected ErrTransitionNotFound, got %v", err)
	}
}

// TestFireEvent_ConditionNotMet tests transition when condition is not satisfied
func TestFireEvent_ConditionNotMet(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&valueCondition{expectedValue: "expected"}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireEvent_ConditionNotMet")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Payload value doesn't match condition
	_, err = sm.FireEvent(StateA, Event1, testPayload{Value: "wrong"})
	if !errors.Is(err, ErrConditionNotMet) {
		t.Errorf("Expected ErrConditionNotMet, got %v", err)
	}
}

// TestFireEvent_ActionError tests transition when action returns error
func TestFireEvent_ActionError(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&errorAction{})

	sm, err := builder.Build("TestFireEvent_ActionError")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	_, err = sm.FireEvent(StateA, Event1, testPayload{})
	if !errors.Is(err, ErrActionExecutionFailed) {
		t.Errorf("Expected ErrActionExecutionFailed, got %v", err)
	}
}

// TestFireEvent_NilCondition tests transition with nil condition (always allowed)
func TestFireEvent_NilCondition(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	// Use WhenFunc with a function that always returns true (simulating no condition)
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		WhenFunc(func(payload testPayload) bool { return true }).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireEvent_NilCondition")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newState, err := sm.FireEvent(StateA, Event1, testPayload{})
	if err != nil {
		t.Fatalf("FireEvent failed: %v", err)
	}

	if newState != StateB {
		t.Errorf("Expected state %v, got %v", StateB, newState)
	}
}

// TestFireEvent_StateMachineNotReady tests firing event before machine is ready
func TestFireEvent_StateMachineNotReady(t *testing.T) {
	sm := newStateMachine[testState, testEvent, testPayload]("NotReady")
	// Not created by Build, so it is never marked ready

	_, err := sm.FireEvent(StateA, Event1, testPayload{})
	if !errors.Is(err, ErrStateMachineNotReady) {
		t.Errorf("Expected ErrStateMachineNotReady, got %v", err)
	}
}

// TestFireParallelEvent_Success tests successful parallel transitions
func TestFireParallelEvent_Success(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateC).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireParallelEvent_Success")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newStates, err := sm.FireParallelEvent(StateA, Event1, testPayload{})
	if err != nil {
		t.Fatalf("FireParallelEvent failed: %v", err)
	}

	if len(newStates) != 2 {
		t.Errorf("Expected 2 states, got %d", len(newStates))
	}

	// Check that both StateB and StateC are in results
	hasB, hasC := false, false
	for _, s := range newStates {
		if s == StateB {
			hasB = true
		}
		if s == StateC {
			hasC = true
		}
	}
	if !hasB || !hasC {
		t.Errorf("Expected states [B, C], got %v", newStates)
	}
}

// TestFireParallelEvent_PartialCondition tests parallel transitions with partial condition match
func TestFireParallelEvent_PartialCondition(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	// Two transitions: one always true, one requires specific value
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateA).
		To(StateC).
		On(Event1).
		When(&valueCondition{expectedValue: "special"}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireParallelEvent_PartialCondition")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// With wrong value, only StateB transition should execute
	newStates, err := sm.FireParallelEvent(StateA, Event1, testPayload{Value: "wrong"})
	if err != nil {
		t.Fatalf("FireParallelEvent failed: %v", err)
	}

	if len(newStates) != 1 || newStates[0] != StateB {
		t.Errorf("Expected [B], got %v", newStates)
	}

	// With correct value, both transitions should execute
	newStates, err = sm.FireParallelEvent(StateA, Event1, testPayload{Value: "special"})
	if err != nil {
		t.Fatalf("FireParallelEvent failed: %v", err)
	}

	if len(newStates) != 2 {
		t.Errorf("Expected 2 states, got %v", newStates)
	}
}

// TestFireParallelEvent_ConditionNotMet tests parallel transitions when no condition is met
func TestFireParallelEvent_ConditionNotMet(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&valueCondition{expectedValue: "value1"}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateA).
		To(StateC).
		On(Event1).
		When(&valueCondition{expectedValue: "value2"}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireParallelEvent_ConditionNotMet")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	_, err = sm.FireParallelEvent(StateA, Event1, testPayload{Value: "wrong"})
	if !errors.Is(err, ErrConditionNotMet) {
		t.Errorf("Expected ErrConditionNotMet, got %v", err)
	}
}

// TestVerify_ValidTransition tests Verify returns true for valid transition
func TestVerify_ValidTransition(t *testing.T) {
	sm := createTestStateMachine(t)

	if !sm.Verify(StateA, Event1) {
		t.Error("Expected Verify to return true for valid transition")
	}
}

// TestVerify_InvalidState tests Verify returns false for non-existent state
func TestVerify_InvalidState(t *testing.T) {
	sm := createTestStateMachine(t)

	if sm.Verify(StateD, Event1) {
		t.Error("Expected Verify to return false for non-existent state")
	}
}

// TestVerify_InvalidEvent tests Verify returns false for non-existent event
func TestVerify_InvalidEvent(t *testing.T) {
	sm := createTestStateMachine(t)

	if sm.Verify(StateA, Event3) {
		t.Error("Expected Verify to return false for non-existent event from this state")
	}
}

// TestInternalTransition tests internal transition stays in same state
func TestInternalTransition(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()
	action := &recordAction{}

	builder.InternalTransition().
		Within(StateA).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(action)

	sm, err := builder.Build("TestInternalTransition")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newState, err := sm.FireEvent(StateA, Event1, testPayload{})
	if err != nil {
		t.Fatalf("FireEvent failed: %v", err)
	}

	if newState != StateA {
		t.Errorf("Expected to stay in state %v, got %v", StateA, newState)
	}

	if !action.executed {
		t.Error("Action was not executed for internal transition")
	}
}

// TestConcurrentFireEvent tests concurrent access to FireEvent
func TestConcurrentFireEvent(t *testing.T) {
	sm := createTestStateMachine(t)
	payload := testPayload{Value: "test"}

	var wg sync.WaitGroup
	numGoroutines := 100
	errChan := make(chan error, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := sm.FireEvent(StateA, Event1, payload)
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent FireEvent failed: %v", err)
	}
}

// TestShowStateMachine tests string representation
func TestShowStateMachine(t *testing.T) {
	sm := createTestStateMachine(t)

	output := sm.ShowStateMachine()
	if output == "" {
		t.Error("ShowStateMachine returned empty string")
	}

	// Check that output contains expected content
	if !contains(output, "StateMachine") {
		t.Error("ShowStateMachine output missing 'StateMachine' header")
	}
}

// TestGenerateDiagram_AllFormats tests diagram generation for all formats
func TestGenerateDiagram_AllFormats(t *testing.T) {
	sm := createTestStateMachine(t)

	tests := []struct {
		format   DiagramFormat
		expected string
	}{
		{PlantUML, "@startuml"},
		{MarkdownTable, "# State Machine"},
		{MarkdownFlowchart, "```mermaid\nflowchart"},
		{MarkdownStateDiagram, "```mermaid\nstateDiagram-v2"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Format_%d", tt.format), func(t *testing.T) {
			output := sm.GenerateDiagram(tt.format)
			if !contains(output, tt.expected) {
				t.Errorf("GenerateDiagram(%d) missing expected content %q", tt.format, tt.expected)
			}
		})
	}
}

// TestGenerateDiagram_DefaultFormat tests default format is PlantUML
func TestGenerateDiagram_DefaultFormat(t *testing.T) {
	sm := createTestStateMachine(t)

	output := sm.GenerateDiagram()
	if !contains(output, "@startuml") {
		t.Error("GenerateDiagram default format should be PlantUML")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ==================== Validation Tests ====================

// TestBuild_NoTransitions tests that building without transitions fails
func TestBuild_NoTransitions(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	_, err := builder.Build("TestBuild_NoTransitions")
	if !errors.Is(err, ErrNoTransitionsDefined) {
		t.Errorf("Expected ErrNoTransitionsDefined, got %v", err)
	}
}

// TestBuild_DuplicateTransition tests that duplicate transitions are detected
func TestBuild_DuplicateTransition(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	// Add the same transition twice
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	_, err := builder.Build("TestBuild_DuplicateTransition")
	if !errors.Is(err, ErrDuplicateTransition) {
		t.Errorf("Expected ErrDuplicateTransition, got %v", err)
	}
}

// TestBuild_SameSourceEventDifferentTarget tests that same source+event to different targets is allowed
func TestBuild_SameSourceEventDifferentTarget(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	// Same source and event, but different targets - should be allowed
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&valueCondition{expectedValue: "toB"}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateA).
		To(StateC).
		On(Event1).
		When(&valueCondition{expectedValue: "toC"}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestBuild_SameSourceEventDifferentTarget")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify both transitions work
	newState, err := sm.FireEvent(StateA, Event1, testPayload{Value: "toB"})
	if err != nil || newState != StateB {
		t.Errorf("Expected StateB, got %v (err: %v)", newState, err)
	}

	newState, err = sm.FireEvent(StateA, Event1, testPayload{Value: "toC"})
	if err != nil || newState != StateC {
		t.Errorf("Expected StateC, got %v (err: %v)", newState, err)
	}
}

// ==================== Factory Tests ====================

// TestGetStateMachine_Success tests retrieving an existing state machine
func TestGetStateMachine_Success(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	built, err := builder.Build("TestGetStateMachine_Success")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}
	if err := RegisterStateMachine("TestGetStateMachine_Success", built); err != nil {
		t.Fatalf("Failed to register state machine: %v", err)
	}
	defer RemoveStateMachine("TestGetStateMachine_Success")

	// Retrieve the state machine
	sm, err := GetStateMachine[testState, testEvent, testPayload]("TestGetStateMachine_Success")
	if err != nil {
		t.Fatalf("Failed to get state machine: %v", err)
	}

	// Verify it works
	newState, err := sm.FireEvent(StateA, Event1, testPayload{})
	if err != nil || newState != StateB {
		t.Errorf("Expected StateB, got %v (err: %v)", newState, err)
	}
}

// TestGetStateMachine_NotFound tests retrieving a non-existent state machine
func TestGetStateMachine_NotFound(t *testing.T) {
	_, err := GetStateMachine[testState, testEvent, testPayload]("NonExistentMachine")
	if !errors.Is(err, ErrStateMachineNotFound) {
		t.Errorf("Expected ErrStateMachineNotFound, got %v", err)
	}
}

// TestGetStateMachine_TypeMismatch tests retrieving with wrong type parameters
func TestGetStateMachine_TypeMismatch(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	built, err := builder.Build("TestGetStateMachine_TypeMismatch")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}
	if err := RegisterStateMachine("TestGetStateMachine_TypeMismatch", built); err != nil {
		t.Fatalf("Failed to register state machine: %v", err)
	}
	defer RemoveStateMachine("TestGetStateMachine_TypeMismatch")

	// Try to retrieve with different type parameters
	_, err = GetStateMachine[string, string, string]("TestGetStateMachine_TypeMismatch")
	if !errors.Is(err, ErrStateMachineNotFound) {
		t.Errorf("Expected ErrStateMachineNotFound for type mismatch, got %v", err)
	}
	if !strings.Contains(err.Error(), "different type parameters") {
		t.Errorf("Expected type mismatch to be explained, got %v", err)
	}
}

// TestListStateMachines tests listing all registered state machines
func TestListStateMachines(t *testing.T) {
	// Create multiple state machines
	ids := []string{"TestList_Machine1", "TestList_Machine2", "TestList_Machine3"}

	for _, id := range ids {
		builder := NewStateMachineBuilder[testState, testEvent, testPayload]()
		builder.ExternalTransition().
			From(StateA).
			To(StateB).
			On(Event1).
			When(&alwaysTrueCondition{}).
			Perform(&noopAction{})

		built, err := builder.Build(id)
		if err != nil {
			t.Fatalf("Failed to build state machine %s: %v", id, err)
		}
		if err := RegisterStateMachine(id, built); err != nil {
			t.Fatalf("Failed to register state machine %s: %v", id, err)
		}
	}

	// Cleanup
	defer func() {
		for _, id := range ids {
			RemoveStateMachine(id)
		}
	}()

	// List all state machines
	list := ListStateMachines()

	// Check that our machines are in the list
	for _, id := range ids {
		found := false
		for _, listed := range list {
			if listed == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find %s in list", id)
		}
	}
}

// TestRegisterStateMachine_AlreadyExists tests registering duplicate ID
func TestRegisterStateMachine_AlreadyExists(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	first, err := builder.Build("TestRegister_Duplicate")
	if err != nil {
		t.Fatalf("Failed to build first state machine: %v", err)
	}
	if err := RegisterStateMachine("TestRegister_Duplicate", first); err != nil {
		t.Fatalf("Failed to register first state machine: %v", err)
	}
	defer RemoveStateMachine("TestRegister_Duplicate")

	// Try to register another with the same ID
	builder2 := NewStateMachineBuilder[testState, testEvent, testPayload]()
	builder2.ExternalTransition().
		From(StateA).
		To(StateC).
		On(Event2).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	second, err := builder2.Build("TestRegister_Duplicate")
	if err != nil {
		t.Fatalf("Failed to build second state machine: %v", err)
	}
	err = RegisterStateMachine("TestRegister_Duplicate", second)
	if !errors.Is(err, ErrStateMachineAlreadyExist) {
		t.Errorf("Expected ErrStateMachineAlreadyExist, got %v", err)
	}
}

// ==================== Builder Interface Tests ====================

// TestExternalTransitions_WithFromBuilder tests ExternalTransitions using FromBuilder methods
func TestExternalTransitions_WithFromBuilder(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	// Use When (interface) and Perform (interface) instead of WhenFunc/PerformFunc
	builder.ExternalTransitions().
		FromAmong(StateA, StateB).
		To(StateC).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestExternalTransitions_WithFromBuilder")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test transition from StateA
	newState, err := sm.FireEvent(StateA, Event1, testPayload{})
	if err != nil || newState != StateC {
		t.Errorf("Expected StateC from StateA, got %v (err: %v)", newState, err)
	}

	// Test transition from StateB
	newState, err = sm.FireEvent(StateB, Event1, testPayload{})
	if err != nil || newState != StateC {
		t.Errorf("Expected StateC from StateB, got %v (err: %v)", newState, err)
	}
}

// TestExternalTransitions_WithWhenFunc tests ExternalTransitions using WhenFunc
func TestExternalTransitions_WithWhenFunc(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	actionExecuted := false
	builder.ExternalTransitions().
		FromAmong(StateA, StateB).
		To(StateC).
		On(Event1).
		WhenFunc(func(payload testPayload) bool {
			return payload.Value == "go"
		}).
		PerformFunc(func(from, to testState, event testEvent, payload testPayload) error {
			actionExecuted = true
			return nil
		})

	sm, err := builder.Build("TestExternalTransitions_WithWhenFunc")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test with matching condition
	newState, err := sm.FireEvent(StateA, Event1, testPayload{Value: "go"})
	if err != nil || newState != StateC {
		t.Errorf("Expected StateC, got %v (err: %v)", newState, err)
	}
	if !actionExecuted {
		t.Error("Action was not executed")
	}
}

// TestExternalParallelTransition_WithWhenAndPerform tests parallel transition with interface methods
func TestExternalParallelTransition_WithWhenAndPerform(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	action := &recordAction{}
	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateC).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(action)

	sm, err := builder.Build("TestExternalParallelTransition_WithWhenAndPerform")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newStates, err := sm.FireParallelEvent(StateA, Event1, testPayload{})
	if err != nil {
		t.Fatalf("FireParallelEvent failed: %v", err)
	}

	if len(newStates) != 2 {
		t.Errorf("Expected 2 states, got %d", len(newStates))
	}
}

// TestExternalParallelTransition_WithWhenFunc tests parallel transition with WhenFunc
func TestExternalParallelTransition_WithWhenFunc(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateC).
		On(Event1).
		WhenFunc(func(payload testPayload) bool {
			return true
		}).
		PerformFunc(func(from, to testState, event testEvent, payload testPayload) error {
			return nil
		})

	sm, err := builder.Build("TestExternalParallelTransition_WithWhenFunc")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newStates, err := sm.FireParallelEvent(StateA, Event1, testPayload{})
	if err != nil {
		t.Fatalf("FireParallelEvent failed: %v", err)
	}

	if len(newStates) != 2 {
		t.Errorf("Expected 2 states, got %d", len(newStates))
	}
}

// TestExternalTransition_WithPerformFunc tests single external transition with PerformFunc
func TestExternalTransition_WithPerformFunc(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	executed := false
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		WhenFunc(func(payload testPayload) bool { return true }).
		PerformFunc(func(from, to testState, event testEvent, payload testPayload) error {
			executed = true
			return nil
		})

	sm, err := builder.Build("TestExternalTransition_WithPerformFunc")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newState, err := sm.FireEvent(StateA, Event1, testPayload{})
	if err != nil || newState != StateB {
		t.Errorf("Expected StateB, got %v (err: %v)", newState, err)
	}
	if !executed {
		t.Error("PerformFunc was not executed")
	}
}

// TestInternalTransition_WithWhenFunc tests internal transition with WhenFunc
func TestInternalTransition_WithWhenFunc(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	executed := false
	builder.InternalTransition().
		Within(StateA).
		On(Event1).
		WhenFunc(func(payload testPayload) bool { return true }).
		PerformFunc(func(from, to testState, event testEvent, payload testPayload) error {
			executed = true
			return nil
		})

	sm, err := builder.Build("TestInternalTransition_WithWhenFunc")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newState, err := sm.FireEvent(StateA, Event1, testPayload{})
	if err != nil || newState != StateA {
		t.Errorf("Expected to stay in StateA, got %v (err: %v)", newState, err)
	}
	if !executed {
		t.Error("PerformFunc was not executed")
	}
}

// ==================== Diagram Tests ====================

// TestGenerateDiagram_MultipleFormats tests generating multiple formats at once
func TestGenerateDiagram_MultipleFormats(t *testing.T) {
	sm := createTestStateMachine(t)

	output := sm.GenerateDiagram(PlantUML, MarkdownTable)

	// Should contain both formats
	if !contains(output, "@startuml") {
		t.Error("Missing PlantUML output")
	}
	if !contains(output, "# State Machine") {
		t.Error("Missing MarkdownTable output")
	}
}

// TestGenerateDiagram_DefaultCase tests the default case in switch
func TestGenerateDiagram_DefaultCase(t *testing.T) {
	sm := createTestStateMachine(t)

	// Use an invalid format value to trigger default case
	output := sm.GenerateDiagram(DiagramFormat(999))

	// Default should be PlantUML
	if !contains(output, "@startuml") {
		t.Error("Default case should generate PlantUML")
	}
}

// TestGenerateDiagram_InternalTransitionInTable tests internal transition display in markdown table
func TestGenerateDiagram_InternalTransitionInTable(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.InternalTransition().
		Within(StateA).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestGenerateDiagram_InternalTransitionInTable")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	output := sm.GenerateDiagram(MarkdownTable)

	if !contains(output, "Internal") {
		t.Error("MarkdownTable should show Internal transition type")
	}
}

// TestGenerateDiagram_InternalTransitionInStateDiagram tests internal transition in state diagram
func TestGenerateDiagram_InternalTransitionInStateDiagram(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.InternalTransition().
		Within(StateA).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestGenerateDiagram_InternalTransitionInStateDiagram")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	output := sm.GenerateDiagram(MarkdownStateDiagram)

	if !contains(output, "[internal]") {
		t.Error("MarkdownStateDiagram should show [internal] marker")
	}
}

// ==================== FSM Edge Case Tests ====================

// TestFireParallelEvent_StateNotFound tests parallel event with non-existent state
func TestFireParallelEvent_StateNotFound(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateC).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireParallelEvent_StateNotFound")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	_, err = sm.FireParallelEvent(StateD, Event1, testPayload{})
	if !errors.Is(err, ErrStateNotFound) {
		t.Errorf("Expected ErrStateNotFound, got %v", err)
	}
}

// TestFireParallelEvent_TransitionNotFound tests parallel event with no matching transition
func TestFireParallelEvent_TransitionNotFound(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateC).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireParallelEvent_TransitionNotFound")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	_, err = sm.FireParallelEvent(StateA, Event2, testPayload{})
	if !errors.Is(err, ErrTransitionNotFound) {
		t.Errorf("Expected ErrTransitionNotFound, got %v", err)
	}
}

// TestFireParallelEvent_NotReady tests parallel event when machine is not ready
func TestFireParallelEvent_NotReady(t *testing.T) {
	sm := newStateMachine[testState, testEvent, testPayload]("NotReadyParallel")
	// Not created by Build, so it is never marked ready

	_, err := sm.FireParallelEvent(StateA, Event1, testPayload{})
	if !errors.Is(err, ErrStateMachineNotReady) {
		t.Errorf("Expected ErrStateMachineNotReady, got %v", err)
	}
}

// TestFireParallelEvent_ActionError tests parallel event when action fails
func TestFireParallelEvent_ActionError(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateC).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&errorAction{})

	sm, err := builder.Build("TestFireParallelEvent_ActionError")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	_, err = sm.FireParallelEvent(StateA, Event1, testPayload{})
	if !errors.Is(err, ErrActionExecutionFailed) {
		t.Errorf("Expected ErrActionExecutionFailed, got %v", err)
	}
}

// TestVerify_NotReady tests Verify when machine is not ready
func TestVerify_NotReady(t *testing.T) {
	sm := newStateMachine[testState, testEvent, testPayload]("NotReadyVerify")
	// Not created by Build, so it is never marked ready

	if sm.Verify(StateA, Event1) {
		t.Error("Expected Verify to return false when machine is not ready")
	}
}

// TestRemoveStateMachine_NotFound tests removing non-existent state machine
func TestRemoveStateMachine_NotFound(t *testing.T) {
	result := RemoveStateMachine("NonExistentMachineToRemove")
	if result {
		t.Error("Expected RemoveStateMachine to return false for non-existent machine")
	}
}

// ==================== Low-level API Tests ====================

// TestAddParallelTransitions tests the AddParallelTransitions method directly
func TestAddParallelTransitions(t *testing.T) {
	stateA := NewState[testState, testEvent, testPayload](StateA)
	stateB := NewState[testState, testEvent, testPayload](StateB)
	stateC := NewState[testState, testEvent, testPayload](StateC)

	transitions := stateA.AddParallelTransitions(Event1, []*State[testState, testEvent, testPayload]{stateB, stateC}, External)

	if len(transitions) != 2 {
		t.Errorf("Expected 2 transitions, got %d", len(transitions))
	}

	// Verify transitions are correctly set up
	for _, tr := range transitions {
		if tr.Source != stateA {
			t.Error("Source should be StateA")
		}
		if tr.Event != Event1 {
			t.Error("Event should be Event1")
		}
	}
}

// TestTransit_InvalidInternalTransition tests Transit with mismatched source/target for internal transition
func TestTransit_InvalidInternalTransition(t *testing.T) {
	stateA := NewState[testState, testEvent, testPayload](StateA)
	stateB := NewState[testState, testEvent, testPayload](StateB)

	// Manually create an invalid internal transition (source != target)
	transition := &Transition[testState, testEvent, testPayload]{
		Source:    stateA,
		Target:    stateB, // Different from source - invalid for Internal type
		Event:     Event1,
		TransType: Internal,
	}

	_, err := transition.Transit(testPayload{}, false)
	if !errors.Is(err, ErrInternalTransition) {
		t.Errorf("Expected ErrInternalTransition, got %v", err)
	}
}

// TestShowStateMachine_WithInternalTransition tests ShowStateMachine output for internal transitions
func TestShowStateMachine_WithInternalTransition(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.InternalTransition().
		Within(StateA).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestShowStateMachine_WithInternalTransition")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	output := sm.ShowStateMachine()

	if !contains(output, "INTERNAL") {
		t.Error("ShowStateMachine should display INTERNAL for internal transitions")
	}
}

// TestTransit_NoAction tests Transit when there is no action defined
func TestTransit_NoAction(t *testing.T) {
	stateA := NewState[testState, testEvent, testPayload](StateA)
	stateB := NewState[testState, testEvent, testPayload](StateB)

	// Create transition without action
	transition := &Transition[testState, testEvent, testPayload]{
		Source:    stateA,
		Target:    stateB,
		Event:     Event1,
		TransType: External,
		Condition: nil,
		Action:    nil, // No action
	}

	result, err := transition.Transit(testPayload{}, true)
	if err != nil {
		t.Errorf("Transit should succeed without action, got error: %v", err)
	}
	if result != stateB {
		t.Errorf("Expected target state %v, got %v", stateB, result)
	}
}

// TestTransit_ConditionNotSatisfied tests Transit when condition returns false
func TestTransit_ConditionNotSatisfied(t *testing.T) {
	stateA := NewState[testState, testEvent, testPayload](StateA)
	stateB := NewState[testState, testEvent, testPayload](StateB)

	// Create transition with condition that returns false
	transition := &Transition[testState, testEvent, testPayload]{
		Source:    stateA,
		Target:    stateB,
		Event:     Event1,
		TransType: External,
		Condition: ConditionFunc[testPayload](func(p testPayload) bool { return false }),
		Action:    nil,
	}

	result, err := transition.Transit(testPayload{}, true)
	if err != nil {
		t.Errorf("Transit should not error when condition not met, got: %v", err)
	}
	// Should stay at source state
	if result != stateA {
		t.Errorf("Expected to stay at source state %v, got %v", stateA, result)
	}
}

// ==================== Regression Tests ====================

// TestFireEvent_ConditionEvaluatedOnce tests that a condition is evaluated only once per FireEvent.
// A condition that changed its answer between two evaluations used to return the source state
// with a nil error and skip the action.
func TestFireEvent_ConditionEvaluatedOnce(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	calls := 0
	action := &recordAction{}
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		WhenFunc(func(payload testPayload) bool {
			calls++
			return calls == 1
		}).
		Perform(action)

	sm, err := builder.Build("TestFireEvent_ConditionEvaluatedOnce")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newState, err := sm.FireEvent(StateA, Event1, testPayload{})
	if err != nil {
		t.Fatalf("FireEvent failed: %v", err)
	}
	if newState != StateB {
		t.Errorf("Expected state %v, got %v", StateB, newState)
	}
	if !action.executed {
		t.Error("Expected action to be executed")
	}
	if calls != 1 {
		t.Errorf("Expected condition to be evaluated once, got %d", calls)
	}
}

// TestBuild_BuilderReuseDoesNotAffectBuiltMachine tests that each Build returns an independent machine
func TestBuild_BuilderReuseDoesNotAffectBuiltMachine(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(nil).
		Perform(&noopAction{})

	sm1, err := builder.Build("TestBuild_BuilderReuse_1")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	builder.ExternalTransition().
		From(StateB).
		To(StateC).
		On(Event2).
		When(nil).
		Perform(&noopAction{})

	sm2, err := builder.Build("TestBuild_BuilderReuse_2")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	if sm1.Verify(StateB, Event2) {
		t.Error("Transition defined after Build leaked into the built state machine")
	}
	if !sm2.Verify(StateB, Event2) {
		t.Error("Expected second Build to include the new transition")
	}
	if !strings.HasPrefix(sm1.ShowStateMachine(), "StateMachine(id=TestBuild_BuilderReuse_1)") {
		t.Errorf("Second Build changed the first machine's id: %q", sm1.ShowStateMachine())
	}
}

// TestBuild_BuilderReuseConcurrentWithFireEvent tests defining transitions while a built machine is in use.
// It used to write maps that FireEvent was reading and crash with a concurrent map access.
func TestBuild_BuilderReuseConcurrentWithFireEvent(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(nil).
		Perform(&noopAction{})

	sm, err := builder.Build("TestBuild_BuilderReuseConcurrent")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if _, err := sm.FireEvent(StateA, Event1, testPayload{}); err != nil {
				t.Errorf("FireEvent failed: %v", err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			builder.ExternalTransition().
				From(StateA).
				To(StateC).
				On(testEvent(fmt.Sprintf("Event_%d", i))).
				When(nil).
				Perform(&noopAction{})
		}
	}()
	wg.Wait()
}

// TestNilConditionAndActionFunc tests that nil funcs act as no condition and no action instead of panicking
func TestNilConditionAndActionFunc(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		WhenFunc(nil).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateB).
		To(StateC).
		On(Event1).
		When(nil).
		PerformFunc(nil)

	builder.InternalTransition().
		Within(StateC).
		On(Event1).
		WhenFunc(nil).
		PerformFunc(nil)

	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateC, StateD).
		On(Event2).
		WhenFunc(nil).
		PerformFunc(nil)

	sm, err := builder.Build("TestNilConditionAndActionFunc")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	tests := []struct {
		from     testState
		expected testState
	}{
		{StateA, StateB},
		{StateB, StateC},
		{StateC, StateC},
	}
	for _, tt := range tests {
		newState, err := sm.FireEvent(tt.from, Event1, testPayload{})
		if err != nil {
			t.Fatalf("FireEvent from %v failed: %v", tt.from, err)
		}
		if newState != tt.expected {
			t.Errorf("From %v: expected state %v, got %v", tt.from, tt.expected, newState)
		}
	}

	results, err := sm.FireParallelEvent(StateA, Event2, testPayload{})
	if err != nil {
		t.Fatalf("FireParallelEvent failed: %v", err)
	}
	if len(results) != 2 || results[0] != StateC || results[1] != StateD {
		t.Errorf("Expected [%v %v], got %v", StateC, StateD, results)
	}
}

// TestActionError_KeepsCause tests that the error returned by an action stays reachable
func TestActionError_KeepsCause(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	cause := errors.New("db timeout")
	failing := func(from, to testState, event testEvent, payload testPayload) error {
		return cause
	}

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(nil).
		PerformFunc(failing)

	builder.ExternalParallelTransition().
		From(StateB).
		ToAmong(StateC, StateD).
		On(Event1).
		When(nil).
		PerformFunc(failing)

	sm, err := builder.Build("TestActionError_KeepsCause")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	_, fireErr := sm.FireEvent(StateA, Event1, testPayload{})
	_, parallelErr := sm.FireParallelEvent(StateB, Event1, testPayload{})

	tests := []struct {
		name    string
		err     error
		message string
	}{
		{"FireEvent", fireErr, "action execution failed (A -[Event1]-> B): db timeout"},
		{"FireParallelEvent", parallelErr, "action execution failed (B -[Event1]-> C): db timeout"},
	}
	for _, tt := range tests {
		if !errors.Is(tt.err, ErrActionExecutionFailed) {
			t.Errorf("%s: expected ErrActionExecutionFailed, got %v", tt.name, tt.err)
		}
		if !errors.Is(tt.err, cause) {
			t.Errorf("%s: expected error to wrap the action's error, got %v", tt.name, tt.err)
		}
		if tt.err == nil || tt.err.Error() != tt.message {
			t.Errorf("%s: expected message %q, got %q", tt.name, tt.message, tt.err)
		}
	}
}

// TestBuild_DoesNotRegister tests that Build leaves the global registry untouched
func TestBuild_DoesNotRegister(t *testing.T) {
	for i := 0; i < 2; i++ {
		builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

		builder.ExternalTransition().
			From(StateA).
			To(StateB).
			On(Event1).
			Perform(&noopAction{})

		// Building the same ID repeatedly must not conflict
		if _, err := builder.Build("TestBuild_DoesNotRegister"); err != nil {
			t.Fatalf("Build #%d failed: %v", i+1, err)
		}
	}

	_, err := GetStateMachine[testState, testEvent, testPayload]("TestBuild_DoesNotRegister")
	if !errors.Is(err, ErrStateMachineNotFound) {
		t.Errorf("Expected built machine not to be registered, got %v", err)
	}
}

// TestFireParallelEvent_StopsAtFirstFailure tests that a failing action skips the remaining
// transitions and the states already reached are returned with the error
func TestFireParallelEvent_StopsAtFirstFailure(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	var executed []testState
	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateC, StateD).
		On(Event1).
		PerformFunc(func(from, to testState, event testEvent, payload testPayload) error {
			executed = append(executed, to)
			if to == StateC {
				return errors.New("boom")
			}
			return nil
		})

	sm, err := builder.Build("TestFireParallelEvent_StopsAtFirstFailure")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	results, err := sm.FireParallelEvent(StateA, Event1, testPayload{})
	if !errors.Is(err, ErrActionExecutionFailed) {
		t.Fatalf("Expected ErrActionExecutionFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "A -[Event1]-> C") {
		t.Errorf("Expected error to name the failed transition, got %v", err)
	}
	if len(results) != 1 || results[0] != StateB {
		t.Errorf("Expected states already reached [%v], got %v", StateB, results)
	}
	if len(executed) != 2 || executed[0] != StateB || executed[1] != StateC {
		t.Errorf("Expected actions for [%v %v] only, got %v", StateB, StateC, executed)
	}
}

// TestFireEvent_ParallelTransition tests that FireEvent rejects parallel transitions instead of firing only one
func TestFireEvent_ParallelTransition(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	action := &recordAction{}
	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateC).
		On(Event1).
		Perform(action)

	sm, err := builder.Build("TestFireEvent_ParallelTransition")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	_, err = sm.FireEvent(StateA, Event1, testPayload{})
	if !errors.Is(err, ErrParallelTransition) {
		t.Errorf("Expected ErrParallelTransition, got %v", err)
	}
	if action.executed {
		t.Error("Expected no action to be executed")
	}
}

// TestBuild_MixedTransitions tests that parallel and non-parallel transitions cannot share a state and event
func TestBuild_MixedTransitions(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&valueCondition{expectedValue: "toB"}).
		Perform(&noopAction{})

	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateC, StateD).
		On(Event1).
		Perform(&noopAction{})

	_, err := builder.Build("TestBuild_MixedTransitions")
	if !errors.Is(err, ErrMixedTransitions) {
		t.Errorf("Expected ErrMixedTransitions, got %v", err)
	}
}

// TestBuild_DuplicateTransitionIgnoresCondition tests that, as in COLA, a source, event and target
// allow only one transition even if the conditions differ
func TestBuild_DuplicateTransitionIgnoresCondition(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.InternalTransition().
		Within(StateA).
		On(Event1).
		When(&valueCondition{expectedValue: "up"}).
		Perform(&noopAction{})

	builder.InternalTransition().
		Within(StateA).
		On(Event1).
		When(&valueCondition{expectedValue: "down"}).
		Perform(&noopAction{})

	_, err := builder.Build("TestBuild_DuplicateTransitionIgnoresCondition")
	if !errors.Is(err, ErrDuplicateTransition) {
		t.Errorf("Expected ErrDuplicateTransition, got %v", err)
	}
}

// TestBuild_MultipleUnconditionalTransitions tests that only one fallback transition is allowed per state and event
func TestBuild_MultipleUnconditionalTransitions(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateA).
		To(StateC).
		On(Event1).
		WhenFunc(nil).
		Perform(&noopAction{})

	_, err := builder.Build("TestBuild_MultipleUnconditionalTransitions")
	if !errors.Is(err, ErrAmbiguousTransition) {
		t.Errorf("Expected ErrAmbiguousTransition, got %v", err)
	}
}

// TestFireEvent_UnconditionalIsFallback tests that, as in COLA, a satisfied condition wins over
// an unconditional transition defined before it
func TestFireEvent_UnconditionalIsFallback(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateA).
		To(StateC).
		On(Event1).
		When(&valueCondition{expectedValue: "toC"}).
		Perform(&noopAction{})

	sm, err := builder.Build("TestFireEvent_UnconditionalIsFallback")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	newState, err := sm.FireEvent(StateA, Event1, testPayload{Value: "toC"})
	if err != nil || newState != StateC {
		t.Errorf("Expected StateC, got %v (err: %v)", newState, err)
	}

	newState, err = sm.FireEvent(StateA, Event1, testPayload{Value: "other"})
	if err != nil || newState != StateB {
		t.Errorf("Expected fallback StateB, got %v (err: %v)", newState, err)
	}
}

// TestBuild_FallbackAfterConditional tests that an unconditional transition may follow conditional ones as a fallback
func TestBuild_FallbackAfterConditional(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&valueCondition{expectedValue: "toB"}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateA).
		To(StateC).
		On(Event1).
		Perform(&noopAction{})

	sm, err := builder.Build("TestBuild_FallbackAfterConditional")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	newState, err := sm.FireEvent(StateA, Event1, testPayload{Value: "toB"})
	if err != nil || newState != StateB {
		t.Errorf("Expected StateB, got %v (err: %v)", newState, err)
	}

	newState, err = sm.FireEvent(StateA, Event1, testPayload{Value: "other"})
	if err != nil || newState != StateC {
		t.Errorf("Expected StateC, got %v (err: %v)", newState, err)
	}
}

// TestBuild_ParallelDuplicateTarget tests that a parallel transition cannot enter the same target twice
func TestBuild_ParallelDuplicateTarget(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalParallelTransition().
		From(StateA).
		ToAmong(StateB, StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	_, err := builder.Build("TestBuild_ParallelDuplicateTarget")
	if !errors.Is(err, ErrDuplicateTransition) {
		t.Errorf("Expected ErrDuplicateTransition, got %v", err)
	}
}

// TestBuild_IncompleteTransition tests that definitions missing Perform are reported instead of silently dropped
func TestBuild_IncompleteTransition(t *testing.T) {
	type builderType = StateMachineBuilder[testState, testEvent, testPayload]

	tests := []struct {
		name     string
		define   func(b *builderType)
		expected string
	}{
		{"ExternalTransition", func(b *builderType) {
			b.ExternalTransition().From(StateB).To(StateC).On(Event2).When(&alwaysTrueCondition{})
		}, "B -[Event2]-> C"},
		{"ExternalTransitions", func(b *builderType) {
			b.ExternalTransitions().FromAmong(StateB, StateC).To(StateD).On(Event2)
		}, "[B C] -[Event2]-> D"},
		{"InternalTransition", func(b *builderType) {
			b.InternalTransition().Within(StateB).On(Event2)
		}, "B -[Event2]-> B"},
		{"ExternalParallelTransition", func(b *builderType) {
			b.ExternalParallelTransition().From(StateB).ToAmong(StateC, StateD).On(Event2)
		}, "B -[Event2]-> [C D]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

			builder.ExternalTransition().
				From(StateA).
				To(StateB).
				On(Event1).
				Perform(&noopAction{})

			tt.define(builder)

			_, err := builder.Build("TestBuild_IncompleteTransition")
			if !errors.Is(err, ErrIncompleteTransition) {
				t.Fatalf("Expected ErrIncompleteTransition, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Expected error to describe %q, got %v", tt.expected, err)
			}
		})
	}
}

// TestOn_PerformWithoutCondition tests that When can be skipped for transitions without condition
func TestOn_PerformWithoutCondition(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	action := &recordAction{}
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		Perform(action)

	builder.InternalTransition().
		Within(StateB).
		On(Event2).
		PerformFunc(func(from, to testState, event testEvent, payload testPayload) error {
			return nil
		})

	builder.ExternalTransitions().
		FromAmong(StateA, StateB).
		To(StateC).
		On(Event3).
		Perform(&noopAction{})

	builder.ExternalParallelTransition().
		From(StateC).
		ToAmong(StateA, StateD).
		On(Event1).
		Perform(&noopAction{})

	sm, err := builder.Build("TestOn_PerformWithoutCondition")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	tests := []struct {
		from     testState
		event    testEvent
		expected testState
	}{
		{StateA, Event1, StateB},
		{StateB, Event2, StateB},
		{StateB, Event3, StateC},
	}
	for _, tt := range tests {
		newState, err := sm.FireEvent(tt.from, tt.event, testPayload{})
		if err != nil || newState != tt.expected {
			t.Errorf("%v -[%v]->: expected %v, got %v (err: %v)", tt.from, tt.event, tt.expected, newState, err)
		}
	}
	if !action.executed {
		t.Error("Expected action to be executed")
	}

	results, err := sm.FireParallelEvent(StateC, Event1, testPayload{})
	if err != nil || len(results) != 2 || results[0] != StateA || results[1] != StateD {
		t.Errorf("Expected [%v %v], got %v (err: %v)", StateA, StateD, results, err)
	}
}

// TestShowStateMachine_StableOrder tests that output is sorted by name instead of following map order
func TestShowStateMachine_StableOrder(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().From(StateC).To(StateA).On(Event1).Perform(&noopAction{})
	builder.ExternalTransition().From(StateA).To(StateB).On(Event2).Perform(&noopAction{})
	builder.ExternalTransition().From(StateB).To(StateC).On(Event3).Perform(&noopAction{})
	builder.ExternalTransition().From(StateA).To(StateC).On(Event1).Perform(&noopAction{})

	sm, err := builder.Build("TestShowStateMachine_StableOrder")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	expected := "StateMachine(id=TestShowStateMachine_StableOrder):\n" +
		"  A --Event1(EXTERNAL)--> C\n" +
		"  A --Event2(EXTERNAL)--> B\n" +
		"  B --Event3(EXTERNAL)--> C\n" +
		"  C --Event1(EXTERNAL)--> A\n"

	// Map iteration order is random, so repeat to catch any dependency on it
	for i := 0; i < 20; i++ {
		if output := sm.ShowStateMachine(); output != expected {
			t.Fatalf("Unexpected output on run %d:\n%s\nexpected:\n%s", i+1, output, expected)
		}
	}
}

// TestGenerateDiagram_EscapesNames tests that names with spaces, quotes, pipes and line breaks
// produce valid diagram syntax, in a stable order
func TestGenerateDiagram_EscapesNames(t *testing.T) {
	builder := NewStateMachineBuilder[testState, testEvent, testPayload]()

	builder.ExternalTransition().
		From("Pending Payment").
		To(`Say "hi"`).
		On("pay|card").
		Perform(&noopAction{})

	builder.InternalTransition().
		Within(`Say "hi"`).
		On("multi\nline").
		Perform(&noopAction{})

	sm, err := builder.Build("TestGenerateDiagram_EscapesNames")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	tests := []struct {
		format   DiagramFormat
		expected string
	}{
		{PlantUML, "@startuml\n" +
			"title StateMachine: TestGenerateDiagram_EscapesNames\n" +
			"state \"Pending Payment\" as s0\n" +
			"state \"Say 'hi'\" as s1\n" +
			"s0 --> s1 : pay|card\n" +
			"s1 --> s1 : multi\\nline\n" +
			"@enduml\n"},
		{MarkdownFlowchart, "```mermaid\nflowchart TD\n" +
			"    state_0[\"Pending Payment\"]\n" +
			"    state_1[\"Say #quot;hi#quot;\"]\n" +
			"    state_0 -->|\"pay|card\"| state_1\n" +
			"    state_1 -->|\"multi line\"| state_1\n" +
			"```\n"},
		{MarkdownStateDiagram, "```mermaid\nstateDiagram-v2\n" +
			"    state \"Pending Payment\" as s0\n" +
			"    state \"Say #quot;hi#quot;\" as s1\n" +
			"    s0 --> s1 : pay|card\n" +
			"    s1 --> s1 : multi line [internal]\n" +
			"```\n"},
		{MarkdownTable, "# State Machine: TestGenerateDiagram_EscapesNames\n\n" +
			"## States\n\n" +
			"- `Pending Payment`\n" +
			"- `Say \"hi\"`\n\n" +
			"## Transitions\n\n" +
			"| Source State | Event | Target State | Type |\n" +
			"|-------------|-------|--------------|------|\n" +
			"| `Pending Payment` | `pay\\|card` | `Say \"hi\"` | External |\n" +
			"| `Say \"hi\"` | `multi line` | `Say \"hi\"` | Internal |\n"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Format_%d", tt.format), func(t *testing.T) {
			// Map iteration order is random, so repeat to catch any dependency on it
			for i := 0; i < 20; i++ {
				if output := sm.GenerateDiagram(tt.format); output != tt.expected {
					t.Fatalf("Unexpected output on run %d:\n%s\nexpected:\n%s", i+1, output, tt.expected)
				}
			}
		})
	}
}
