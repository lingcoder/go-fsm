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
