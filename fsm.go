package fsm

import (
	"fmt"
	"sort"
	"strings"
)

// StateMachine is a generic state machine interface.
//
// S: State type, must be comparable (e.g., string, int)
// E: Event type, must be comparable
// P: Payload type, can be any type, used to pass data during state transitions
//
// Thread Safety:
//   - StateMachine is safe for concurrent use by multiple goroutines.
//   - Multiple goroutines can call FireEvent simultaneously.
//   - A built StateMachine is immutable, so its methods take no locks.
//   - StateMachineBuilder is not safe for concurrent use. Build returns an
//     independent copy, so reusing the builder never affects built machines.
//
// Payload Usage:
//   - The payload is passed by value to conditions and actions.
//   - However, if the payload contains pointers, slices, or maps,
//     the underlying data is shared and may cause race conditions.
//   - For concurrent scenarios with mutable payload data, consider
//     passing a copy or using synchronization within your actions.
type StateMachine[S comparable, E comparable, P any] interface {
	// FireEvent triggers a state transition based on the current state and event
	// Returns the new state and any error that occurred
	// Returns ErrParallelTransition if the event is defined as a parallel transition
	FireEvent(sourceState S, event E, payload P) (S, error)

	// FireParallelEvent triggers parallel state transitions based on the current state and event
	// Returns a slice of new states and any error that occurred
	// If an action fails, the remaining transitions are skipped and the states already
	// transitioned to are returned together with the error
	FireParallelEvent(sourceState S, event E, payload P) ([]S, error)

	// Verify checks if there is a valid transition for the given state and event
	// Returns true if a transition exists, false otherwise
	Verify(sourceState S, event E) bool

	// ShowStateMachine returns a string representation of the state machine
	ShowStateMachine() string

	// GenerateDiagram returns a diagram of the state machine in the specified formats
	// If formats is nil or empty, defaults to PlantUML
	// If multiple formats are provided, returns all requested formats concatenated
	GenerateDiagram(formats ...DiagramFormat) string
}

// DiagramFormat defines the supported diagram formats
type DiagramFormat int

const (
	// PlantUML format for UML diagrams
	PlantUML DiagramFormat = iota
	// MarkdownTable format for tabular representation
	MarkdownTable
	// MarkdownFlowchart format for flowcharts
	MarkdownFlowchart
	// MarkdownStateDiagram format for Mermaid state diagrams
	MarkdownStateDiagram
)

// State represents a state in the state machine
type State[S comparable, E comparable, P any] struct {
	id               S
	eventTransitions map[E][]*Transition[S, E, P]
}

// NewState creates a new state
func NewState[S comparable, E comparable, P any](id S) *State[S, E, P] {
	return &State[S, E, P]{
		id:               id,
		eventTransitions: make(map[E][]*Transition[S, E, P]),
	}
}

// AddTransition adds a transition to this state
func (s *State[S, E, P]) AddTransition(event E, target *State[S, E, P], transType TransitionType) *Transition[S, E, P] {
	transition := &Transition[S, E, P]{
		Source:    s,
		Target:    target,
		Event:     event,
		TransType: transType,
	}

	if _, ok := s.eventTransitions[event]; !ok {
		s.eventTransitions[event] = make([]*Transition[S, E, P], 0)
	}
	s.eventTransitions[event] = append(s.eventTransitions[event], transition)
	return transition
}

// AddParallelTransitions adds multiple transitions for the same event to different target states
func (s *State[S, E, P]) AddParallelTransitions(event E, targets []*State[S, E, P], transType TransitionType) []*Transition[S, E, P] {
	transitions := make([]*Transition[S, E, P], 0, len(targets))

	for _, target := range targets {
		transition := s.AddTransition(event, target, transType)
		transitions = append(transitions, transition)
	}

	return transitions
}

// GetEventTransitions returns all transitions for a given event
func (s *State[S, E, P]) GetEventTransitions(event E) []*Transition[S, E, P] {
	return s.eventTransitions[event]
}

// GetID returns the state ID
func (s *State[S, E, P]) GetID() S {
	return s.id
}

// TransitionType defines the type of transition
type TransitionType int

const (
	// External transitions change the state from source to target
	External TransitionType = iota
	// Internal transitions don't change the state
	Internal
)

// Condition is an interface for transition conditions
type Condition[P any] interface {
	// IsSatisfied returns true if the condition is met
	IsSatisfied(payload P) bool
}

// ConditionFunc is a function type that implements Condition interface
type ConditionFunc[P any] func(payload P) bool

// IsSatisfied implements Condition interface.
// A nil ConditionFunc is always satisfied, like a transition without condition.
func (f ConditionFunc[P]) IsSatisfied(payload P) bool {
	if f == nil {
		return true
	}
	return f(payload)
}

// isUnconditional reports whether c always passes: no condition or a nil ConditionFunc
func isUnconditional[P any](c Condition[P]) bool {
	if c == nil {
		return true
	}
	f, ok := c.(ConditionFunc[P])
	return ok && f == nil
}

// Action is an interface for transition actions
type Action[S comparable, E comparable, P any] interface {
	// Execute runs the action during a state transition
	Execute(from, to S, event E, payload P) error
}

// ActionFunc is a function type that implements Action interface
type ActionFunc[S comparable, E comparable, P any] func(from, to S, event E, payload P) error

// Execute implements Action interface.
// A nil ActionFunc does nothing.
func (f ActionFunc[S, E, P]) Execute(from, to S, event E, payload P) error {
	if f == nil {
		return nil
	}
	return f(from, to, event, payload)
}

// Transition represents a state transition
type Transition[S comparable, E comparable, P any] struct {
	Source    *State[S, E, P]
	Target    *State[S, E, P]
	Event     E
	Condition Condition[P]
	Action    Action[S, E, P]
	TransType TransitionType

	// parallel marks transitions defined with ExternalParallelTransition
	parallel bool
}

// Transit executes the transition
func (t *Transition[S, E, P]) Transit(payload P, checkCondition bool) (*State[S, E, P], error) {
	// Verify internal transition
	if t.TransType == Internal && t.Source != t.Target {
		return nil, ErrInternalTransition
	}

	// Check condition if required
	if checkCondition && t.Condition != nil && !t.Condition.IsSatisfied(payload) {
		return t.Source, nil // Stay at source state if condition is not satisfied
	}

	// Execute action
	if t.Action != nil {
		if err := t.Action.Execute(t.Source.GetID(), t.Target.GetID(), t.Event, payload); err != nil {
			return nil, &actionError{
				transition: fmt.Sprintf("%v -[%v]-> %v", t.Source.id, t.Event, t.Target.id),
				cause:      err,
			}
		}
	}

	return t.Target, nil
}

// StateMachineImpl implements the StateMachine interface.
// Once built it is never modified, so reads need no locking.
type StateMachineImpl[S comparable, E comparable, P any] struct {
	id       string
	stateMap map[S]*State[S, E, P]
	ready    bool
}

// newStateMachine creates a new state machine (package private)
func newStateMachine[S comparable, E comparable, P any](id string) *StateMachineImpl[S, E, P] {
	return &StateMachineImpl[S, E, P]{
		id:       id,
		stateMap: make(map[S]*State[S, E, P]),
		ready:    false,
	}
}

// snapshot returns a ready deep copy of the state graph with the given id.
// The copy shares no maps or slices with sm, so it stays immutable even if
// the builder that owns sm keeps adding transitions.
func (sm *StateMachineImpl[S, E, P]) snapshot(id string) *StateMachineImpl[S, E, P] {
	clone := newStateMachine[S, E, P](id)
	for stateId := range sm.stateMap {
		clone.getOrCreateState(stateId)
	}

	for stateId, state := range sm.stateMap {
		source := clone.stateMap[stateId]
		for event, transitions := range state.eventTransitions {
			// Keep the order within an event: FireEvent tries transitions in definition order
			for _, t := range transitions {
				copied := source.AddTransition(event, clone.stateMap[t.Target.id], t.TransType)
				copied.Condition = t.Condition
				copied.Action = t.Action
				copied.parallel = t.parallel
			}
		}
	}

	clone.ready = true
	return clone
}

// forEachTransition calls fn for every transition, with states and events sorted by name
// so output is stable even though Go randomizes map iteration order
func (sm *StateMachineImpl[S, E, P]) forEachTransition(fn func(transition *Transition[S, E, P])) {
	for _, stateId := range sortedKeys(sm.stateMap) {
		state := sm.stateMap[stateId]
		for _, event := range sortedKeys(state.eventTransitions) {
			for _, transition := range state.eventTransitions[event] {
				fn(transition)
			}
		}
	}
}

// sortedKeys returns the keys of m sorted by their string form, since keys are only comparable
func sortedKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j])
	})
	return keys
}

// FireEvent triggers a state transition based on the current state and event
func (sm *StateMachineImpl[S, E, P]) FireEvent(sourceStateId S, event E, payload P) (S, error) {
	if !sm.ready {
		var zeroState S
		return zeroState, ErrStateMachineNotReady
	}

	// Get source state
	sourceState, ok := sm.stateMap[sourceStateId]
	if !ok {
		var zeroState S
		return zeroState, ErrStateNotFound
	}

	// Get transitions for the event
	transitions := sourceState.GetEventTransitions(event)
	if transitions == nil || len(transitions) == 0 {
		var zeroState S
		return zeroState, ErrTransitionNotFound
	}

	// Parallel transitions fire together; picking one would silently skip the others
	if transitions[0].parallel {
		var zeroState S
		return zeroState, ErrParallelTransition
	}

	// Find the first transition with satisfied condition
	for _, transition := range transitions {
		if transition.Condition == nil || transition.Condition.IsSatisfied(payload) {
			// Skip the condition check in Transit: a second evaluation could disagree
			// and silently return the source state without running the action
			targetState, err := transition.Transit(payload, false)
			if err != nil {
				var zeroState S
				return zeroState, err
			}
			return targetState.GetID(), nil
		}
	}

	var zeroState S
	return zeroState, ErrConditionNotMet
}

// FireParallelEvent triggers parallel state transitions based on the current state and event
func (sm *StateMachineImpl[S, E, P]) FireParallelEvent(sourceStateId S, event E, payload P) ([]S, error) {
	if !sm.ready {
		return nil, ErrStateMachineNotReady
	}

	// Get source state
	sourceState, ok := sm.stateMap[sourceStateId]
	if !ok {
		return nil, ErrStateNotFound
	}

	// Get transitions for the event
	transitions := sourceState.GetEventTransitions(event)
	if transitions == nil || len(transitions) == 0 {
		return nil, ErrTransitionNotFound
	}

	// Execute all transitions with satisfied conditions
	var results []S
	var validTransitions []*Transition[S, E, P]

	// First, find all valid transitions
	for _, transition := range transitions {
		if transition.Condition == nil || transition.Condition.IsSatisfied(payload) {
			validTransitions = append(validTransitions, transition)
		}
	}

	if len(validTransitions) == 0 {
		return nil, ErrConditionNotMet
	}

	// Then execute all valid transitions, stopping at the first failure.
	// Actions that already ran can't be undone, so report the states reached so far.
	for _, transition := range validTransitions {
		targetState, err := transition.Transit(payload, false) // Skip condition check as we've already verified it
		if err != nil {
			return results, err
		}
		results = append(results, targetState.GetID())
	}

	return results, nil
}

// Verify checks if there is a valid transition for the given state and event
func (sm *StateMachineImpl[S, E, P]) Verify(sourceStateId S, event E) bool {
	if !sm.ready {
		return false
	}

	// Get source state
	sourceState, ok := sm.stateMap[sourceStateId]
	if !ok {
		return false
	}

	// Get transitions for the event
	transitions := sourceState.GetEventTransitions(event)

	// Return true if there is at least one transition for this event
	return transitions != nil && len(transitions) > 0
}

// getOrCreateState returns a state by ID, creating it if it doesn't exist.
// Only the builder calls it, before the machine is built.
func (sm *StateMachineImpl[S, E, P]) getOrCreateState(stateId S) *State[S, E, P] {
	if state, ok := sm.stateMap[stateId]; ok {
		return state
	}

	state := NewState[S, E, P](stateId)
	sm.stateMap[stateId] = state
	return state
}

// ShowStateMachine returns a string representation of the state machine
func (sm *StateMachineImpl[S, E, P]) ShowStateMachine() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "StateMachine(id=%s):\n", sm.id)

	sm.forEachTransition(func(transition *Transition[S, E, P]) {
		transType := "EXTERNAL"
		if transition.TransType == Internal {
			transType = "INTERNAL"
		}
		fmt.Fprintf(&sb, "  %v --%v(%s)--> %v\n",
			transition.Source.GetID(), transition.Event, transType, transition.Target.GetID())
	})

	return sb.String()
}
