package fsm

import (
	"errors"
	"fmt"
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

// testCounter for unique state machine IDs
var testCounter int
var testCounterMutex sync.Mutex

func getUniqueTestID(prefix string) string {
	testCounterMutex.Lock()
	defer testCounterMutex.Unlock()
	testCounter++
	return fmt.Sprintf("%s_%d", prefix, testCounter)
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

	// Build the state machine with unique ID
	machineID := getUniqueTestID("TestStateMachine")
	sm, err := builder.Build(machineID)
	if err != nil {
		tb.Fatalf("Failed to build state machine: %v", err)
	}

	// Cleanup after test
	tb.Cleanup(func() {
		RemoveStateMachine(machineID)
	})

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

		machineId := fmt.Sprintf("TestMachine-%d", i)
		_, err := builder.Build(machineId)
		if err != nil {
			b.Fatalf("Failed to build state machine: %v", err)
		}

		// Clean up to avoid memory leaks
		RemoveStateMachine(machineId)
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
	defer RemoveStateMachine("TestFireEvent_Success")

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
	defer RemoveStateMachine("TestFireEvent_StateNotFound")

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
	defer RemoveStateMachine("TestFireEvent_TransitionNotFound")

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
	defer RemoveStateMachine("TestFireEvent_ConditionNotMet")

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
	defer RemoveStateMachine("TestFireEvent_ActionError")

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
	defer RemoveStateMachine("TestFireEvent_NilCondition")

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
	// Don't call SetReady(true)

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
	defer RemoveStateMachine("TestFireParallelEvent_Success")

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
	defer RemoveStateMachine("TestFireParallelEvent_PartialCondition")

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
	defer RemoveStateMachine("TestFireParallelEvent_ConditionNotMet")

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
	defer RemoveStateMachine("TestInternalTransition")

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
	defer RemoveStateMachine("TestBuild_SameSourceEventDifferentTarget")

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

	_, err := builder.Build("TestGetStateMachine_Success")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
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

	_, err := builder.Build("TestGetStateMachine_TypeMismatch")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}
	defer RemoveStateMachine("TestGetStateMachine_TypeMismatch")

	// Try to retrieve with different type parameters
	_, err = GetStateMachine[string, string, string]("TestGetStateMachine_TypeMismatch")
	if !errors.Is(err, ErrStateMachineNotFound) {
		t.Errorf("Expected ErrStateMachineNotFound for type mismatch, got %v", err)
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

		_, err := builder.Build(id)
		if err != nil {
			t.Fatalf("Failed to build state machine %s: %v", id, err)
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

	_, err := builder.Build("TestRegister_Duplicate")
	if err != nil {
		t.Fatalf("Failed to build first state machine: %v", err)
	}
	defer RemoveStateMachine("TestRegister_Duplicate")

	// Try to build another with the same ID
	builder2 := NewStateMachineBuilder[testState, testEvent, testPayload]()
	builder2.ExternalTransition().
		From(StateA).
		To(StateC).
		On(Event2).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	_, err = builder2.Build("TestRegister_Duplicate")
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
	defer RemoveStateMachine("TestExternalTransitions_WithFromBuilder")

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
	defer RemoveStateMachine("TestExternalTransitions_WithWhenFunc")

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
	defer RemoveStateMachine("TestExternalParallelTransition_WithWhenAndPerform")

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
	defer RemoveStateMachine("TestExternalParallelTransition_WithWhenFunc")

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
	defer RemoveStateMachine("TestExternalTransition_WithPerformFunc")

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
	defer RemoveStateMachine("TestInternalTransition_WithWhenFunc")

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
	defer RemoveStateMachine("TestGenerateDiagram_InternalTransitionInTable")

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
	defer RemoveStateMachine("TestGenerateDiagram_InternalTransitionInStateDiagram")

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
	defer RemoveStateMachine("TestFireParallelEvent_StateNotFound")

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
	defer RemoveStateMachine("TestFireParallelEvent_TransitionNotFound")

	_, err = sm.FireParallelEvent(StateA, Event2, testPayload{})
	if !errors.Is(err, ErrTransitionNotFound) {
		t.Errorf("Expected ErrTransitionNotFound, got %v", err)
	}
}

// TestFireParallelEvent_NotReady tests parallel event when machine is not ready
func TestFireParallelEvent_NotReady(t *testing.T) {
	sm := newStateMachine[testState, testEvent, testPayload]("NotReadyParallel")
	// Don't call SetReady(true)

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
	defer RemoveStateMachine("TestFireParallelEvent_ActionError")

	_, err = sm.FireParallelEvent(StateA, Event1, testPayload{})
	if !errors.Is(err, ErrActionExecutionFailed) {
		t.Errorf("Expected ErrActionExecutionFailed, got %v", err)
	}
}

// TestVerify_NotReady tests Verify when machine is not ready
func TestVerify_NotReady(t *testing.T) {
	sm := newStateMachine[testState, testEvent, testPayload]("NotReadyVerify")
	// Don't call SetReady(true)

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
	defer RemoveStateMachine("TestShowStateMachine_WithInternalTransition")

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
