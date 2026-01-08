package fsm

import "fmt"

// FromStep Step marker interfaces to enforce the correct order of method calls
type FromStep interface{}
type ToStep interface{}
type OnStep interface{}
type WhenStep interface{}
type PerformStep interface{}
type ToAmongStep interface{}
type WithinStep interface{}

// TransitionStarterInterface is the interface for starting a transition definition
type TransitionStarterInterface[S comparable, E comparable, P any] interface {
	// ExternalTransition starts defining an external transition
	ExternalTransition() ExternalTransitionBuilderInterface[S, E, P]

	// InternalTransition starts defining an internal transition
	InternalTransition() InternalTransitionBuilderInterface[S, E, P]

	// ExternalTransitions starts defining multiple external transitions
	ExternalTransitions() ExternalTransitionsBuilderInterface[S, E, P]

	// ExternalParallelTransition starts defining an external parallel transition
	ExternalParallelTransition() ExternalParallelTransitionBuilderInterface[S, E, P]
}

// ExternalTransitionBuilderInterface is the interface for building external transitions
type ExternalTransitionBuilderInterface[S comparable, E comparable, P any] interface {
	// From specifies the source state
	From(state S) FromInterface[S, E, P]
}

// FromInterface is the interface for specifying the source state of a transition
type FromInterface[S comparable, E comparable, P any] interface {
	// To specifies the target state
	To(state S) ToInterface[S, E, P]
}

// ToInterface is the interface for specifying the target state of a transition
type ToInterface[S comparable, E comparable, P any] interface {
	// On specifies the triggering event
	On(event E) OnInterface[S, E, P]
}

// OnInterface is the interface for specifying the triggering event of a transition
type OnInterface[S comparable, E comparable, P any] interface {
	// When specifies the condition for the transition
	When(condition Condition[P]) WhenInterface[S, E, P]

	// WhenFunc specifies a function as the condition for the transition
	WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P]
}

// WhenInterface is the interface for specifying the condition of a transition
type WhenInterface[S comparable, E comparable, P any] interface {
	// Perform specifies the action to execute during the transition
	Perform(action Action[S, E, P])

	// PerformFunc specifies a function as the action to execute during the transition
	PerformFunc(actionFunc func(from, to S, event E, payload P) error)
}

// InternalTransitionBuilderInterface is the interface for building internal transitions
type InternalTransitionBuilderInterface[S comparable, E comparable, P any] interface {
	// Within specifies the state where the internal transition occurs
	Within(state S) ToInterface[S, E, P]
}

// ExternalTransitionsBuilderInterface is the interface for building multiple external transitions
type ExternalTransitionsBuilderInterface[S comparable, E comparable, P any] interface {
	// FromAmong specifies multiple source states
	FromAmong(states ...S) FromInterface[S, E, P]
}

// ExternalParallelTransitionBuilderInterface is the interface for building external parallel transitions
type ExternalParallelTransitionBuilderInterface[S comparable, E comparable, P any] interface {
	// From specifies the source state
	From(state S) ParallelFromInterface[S, E, P]
}

// ParallelFromInterface is the interface for specifying the source state of a parallel transition
type ParallelFromInterface[S comparable, E comparable, P any] interface {
	// ToAmong specifies multiple target states
	ToAmong(states ...S) ToInterface[S, E, P]
}

// Type assertions to ensure implementations satisfy interfaces
var (
	_ TransitionStarterInterface[string, string, any]                 = (*StateMachineBuilder[string, string, any])(nil)
	_ ExternalTransitionBuilderInterface[string, string, any]         = (*TransitionBuilder[string, string, any, FromStep])(nil)
	_ ExternalTransitionsBuilderInterface[string, string, any]        = (*ExternalTransitionsBuilder[string, string, any])(nil)
	_ ExternalParallelTransitionBuilderInterface[string, string, any] = (*ExternalParallelTransitionBuilder[string, string, any])(nil)
	_ InternalTransitionBuilderInterface[string, string, any]         = (*InternalTransitionBuilder[string, string, any])(nil)
	_ ParallelFromInterface[string, string, any]                      = (*ParallelFromBuilder[string, string, any, ToAmongStep])(nil)
	_ ToInterface[string, string, any]                                = (*ParallelFromBuilder[string, string, any, OnStep])(nil)
	_ OnInterface[string, string, any]                                = (*ParallelFromBuilder[string, string, any, WhenStep])(nil)
	_ WhenInterface[string, string, any]                              = (*ParallelFromBuilder[string, string, any, PerformStep])(nil)
	_ FromInterface[string, string, any]                              = (*FromBuilder[string, string, any, ToStep])(nil)
	_ ToInterface[string, string, any]                                = (*FromBuilder[string, string, any, OnStep])(nil)
	_ OnInterface[string, string, any]                                = (*FromBuilder[string, string, any, WhenStep])(nil)
	_ WhenInterface[string, string, any]                              = (*FromBuilder[string, string, any, PerformStep])(nil)
	_ ToInterface[string, string, any]                                = (*OnTransitionBuilder[string, string, any, OnStep])(nil)
	_ OnInterface[string, string, any]                                = (*OnTransitionBuilder[string, string, any, WhenStep])(nil)
	_ WhenInterface[string, string, any]                              = (*OnTransitionBuilder[string, string, any, PerformStep])(nil)
)

// StateMachineBuilder builds state machines with a fluent API
type StateMachineBuilder[S comparable, E comparable, P any] struct {
	stateMachine *StateMachineImpl[S, E, P]
}

// NewStateMachineBuilder creates a new builder
// Returns:
//
//	A new state machine builder instance
func NewStateMachineBuilder[S comparable, E comparable, P any]() *StateMachineBuilder[S, E, P] {
	return &StateMachineBuilder[S, E, P]{
		stateMachine: newStateMachine[S, E, P](""),
	}
}

// ExternalTransition starts defining an external transition
// Returns:
//
//	A transition builder for configuring the external transition
func (b *StateMachineBuilder[S, E, P]) ExternalTransition() ExternalTransitionBuilderInterface[S, E, P] {
	return &TransitionBuilder[S, E, P, FromStep]{
		stateMachine:   b.stateMachine,
		transitionType: External,
	}
}

// InternalTransition starts defining an internal transition
// Returns:
//
//	A internal transition builder for configuring the internal transition
func (b *StateMachineBuilder[S, E, P]) InternalTransition() InternalTransitionBuilderInterface[S, E, P] {
	return &InternalTransitionBuilder[S, E, P]{
		stateMachine: b.stateMachine,
	}
}

// ExternalTransitions starts defining multiple external transitions from different source states to the same target state
// Returns:
//
//	A external transitions builder for configuring the transitions
func (b *StateMachineBuilder[S, E, P]) ExternalTransitions() ExternalTransitionsBuilderInterface[S, E, P] {
	return &ExternalTransitionsBuilder[S, E, P]{
		stateMachine: b.stateMachine,
	}
}

// ExternalParallelTransition starts defining an external parallel transition
// Returns:
//
//	A external parallel transition builder for configuring the parallel transition
func (b *StateMachineBuilder[S, E, P]) ExternalParallelTransition() ExternalParallelTransitionBuilderInterface[S, E, P] {
	return &ExternalParallelTransitionBuilder[S, E, P]{
		stateMachine: b.stateMachine,
	}
}

// Build finalizes the state machine with the given ID
// Parameters:
//
//	machineId: Unique identifier for the state machine
//
// Returns:
//
//	The built state machine and possible error
func (b *StateMachineBuilder[S, E, P]) Build(machineId string) (StateMachine[S, E, P], error) {
	b.stateMachine.id = machineId

	// Validate the state machine configuration
	if err := b.validate(); err != nil {
		return nil, err
	}

	b.stateMachine.SetReady(true)

	// Register the state machine in a factory
	err := RegisterStateMachine[S, E, P](machineId, b.stateMachine)
	if err != nil {
		return nil, err
	}
	return b.stateMachine, nil
}

// validate checks the state machine configuration for errors
func (b *StateMachineBuilder[S, E, P]) validate() error {
	// Check if any transitions are defined
	if len(b.stateMachine.stateMap) == 0 {
		return ErrNoTransitionsDefined
	}

	// Check for duplicate transitions (same source, event, and target)
	for _, state := range b.stateMachine.stateMap {
		for event, transitions := range state.eventTransitions {
			seen := make(map[string]bool)
			for _, t := range transitions {
				// Create a unique key for source-event-target combination
				key := fmt.Sprintf("%v->%v->%v", t.Source.id, event, t.Target.id)
				if seen[key] {
					return fmt.Errorf("%w: %v -[%v]-> %v", ErrDuplicateTransition, t.Source.id, event, t.Target.id)
				}
				seen[key] = true
			}
		}
	}

	return nil
}

// ExternalTransitionsBuilder builds external transitions from multiple source states to a single target state
type ExternalTransitionsBuilder[S comparable, E comparable, P any] struct {
	stateMachine *StateMachineImpl[S, E, P]
}

// FromAmong specifies multiple source states
// Parameters:
//
//	states: Multiple source states
//
// Returns:
//
//	The from builder for method chaining
func (b *ExternalTransitionsBuilder[S, E, P]) FromAmong(states ...S) FromInterface[S, E, P] {
	return &FromBuilder[S, E, P, ToStep]{
		stateMachine:   b.stateMachine,
		sourceIds:      states,
		transitionType: External,
	}
}

// ExternalParallelTransitionBuilder builds external parallel transitions
type ExternalParallelTransitionBuilder[S comparable, E comparable, P any] struct {
	stateMachine *StateMachineImpl[S, E, P]
}

// From specifies the source state
// Parameters:
//
//	state: Source state
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *ExternalParallelTransitionBuilder[S, E, P]) From(state S) ParallelFromInterface[S, E, P] {
	return &ParallelFromBuilder[S, E, P, ToAmongStep]{
		stateMachine:   b.stateMachine,
		sourceId:       state,
		transitionType: External,
	}
}

// ParallelFromBuilder builds the "from" part of a parallel transition
type ParallelFromBuilder[S comparable, E comparable, P any, Next any] struct {
	stateMachine   *StateMachineImpl[S, E, P]
	transitionType TransitionType
	sourceId       S
	targetIds      []S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
}

// ToAmong specifies multiple target states
// Parameters:
//
//	states: Multiple target states
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *ParallelFromBuilder[S, E, P, Next]) ToAmong(states ...S) ToInterface[S, E, P] {
	b.targetIds = states
	return (*ParallelFromBuilder[S, E, P, OnStep])(b)
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers these transitions
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *ParallelFromBuilder[S, E, P, Next]) On(event E) OnInterface[S, E, P] {
	b.event = event
	return (*ParallelFromBuilder[S, E, P, WhenStep])(b)
}

// When specifies the condition for all transitions
// Parameters:
//
//	condition: The condition that must be satisfied for the transitions to occur
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *ParallelFromBuilder[S, E, P, Next]) When(condition Condition[P]) WhenInterface[S, E, P] {
	b.condition = condition
	return (*ParallelFromBuilder[S, E, P, PerformStep])(b)
}

// WhenFunc specifies a function as the condition for all transitions
// Parameters:
//
//	conditionFunc: The function that must return true for the transitions to occur
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *ParallelFromBuilder[S, E, P, Next]) WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return (*ParallelFromBuilder[S, E, P, PerformStep])(b)
}

// Perform specifies the action to execute during all transitions
// Parameters:
//
//	action: The action to execute when the transitions occur
func (b *ParallelFromBuilder[S, E, P, Next]) Perform(action Action[S, E, P]) {
	b.action = action

	// Get or create source state
	sourceState := b.stateMachine.GetState(b.sourceId)

	// Create transitions to all target states
	for _, targetId := range b.targetIds {
		targetState := b.stateMachine.GetState(targetId)
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)
		transition.Condition = b.condition
		transition.Action = b.action
	}
}

// PerformFunc specifies a function as the action to execute during all transitions
// Parameters:
//
//	actionFunc: The function to execute when the transitions occur
func (b *ParallelFromBuilder[S, E, P, Next]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.action = ActionFunc[S, E, P](actionFunc)

	// Get or create source state
	sourceState := b.stateMachine.GetState(b.sourceId)

	// Create transitions to all target states
	for _, targetId := range b.targetIds {
		targetState := b.stateMachine.GetState(targetId)
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)
		transition.Condition = b.condition
		transition.Action = b.action
	}
}

// TransitionBuilder builds individual transitions
type TransitionBuilder[S comparable, E comparable, P any, Next any] struct {
	stateMachine   *StateMachineImpl[S, E, P]
	transitionType TransitionType
	sourceId       S
	targetId       S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
}

// From specifies the source state
// Parameters:
//
//	state: Source state
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, P, Next]) From(state S) FromInterface[S, E, P] {
	b.sourceId = state
	return (*TransitionBuilder[S, E, P, ToStep])(b)
}

// To specifies the target state
// Parameters:
//
//	state: Target state
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, P, Next]) To(state S) ToInterface[S, E, P] {
	b.targetId = state
	return (*TransitionBuilder[S, E, P, OnStep])(b)
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers this transition
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, P, Next]) On(event E) OnInterface[S, E, P] {
	b.event = event
	return (*TransitionBuilder[S, E, P, WhenStep])(b)
}

// When specifies the condition for the transition
// Parameters:
//
//	condition: The condition that must be satisfied for the transition to occur
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, P, Next]) When(condition Condition[P]) WhenInterface[S, E, P] {
	b.condition = condition
	return (*TransitionBuilder[S, E, P, PerformStep])(b)
}

// WhenFunc specifies a function as the condition for the transition
// Parameters:
//
//	conditionFunc: The function that must return true for the transition to occur
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, P, Next]) WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return (*TransitionBuilder[S, E, P, PerformStep])(b)
}

// Perform specifies the action to execute during the transition
// Parameters:
//
//	action: The action to execute when the transition occurs
func (b *TransitionBuilder[S, E, P, Next]) Perform(action Action[S, E, P]) {
	b.action = action

	// Get or create states
	sourceState := b.stateMachine.GetState(b.sourceId)
	targetState := b.stateMachine.GetState(b.targetId)

	// Create transition
	transition := sourceState.AddTransition(b.event, targetState, b.transitionType)
	transition.Condition = b.condition
	transition.Action = b.action
}

// PerformFunc specifies a function as the action to execute during the transition
// Parameters:
//
//	actionFunc: The function to execute when the transition occurs
func (b *TransitionBuilder[S, E, P, Next]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.action = ActionFunc[S, E, P](actionFunc)

	// Get or create states
	sourceState := b.stateMachine.GetState(b.sourceId)
	targetState := b.stateMachine.GetState(b.targetId)

	// Create transition
	transition := sourceState.AddTransition(b.event, targetState, b.transitionType)
	transition.Condition = b.condition
	transition.Action = b.action
}

// FromBuilder builds the "from" part of multiple transitions
type FromBuilder[S comparable, E comparable, P any, Next any] struct {
	stateMachine   *StateMachineImpl[S, E, P]
	transitionType TransitionType
	sourceIds      []S
	targetId       S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
}

// To specifies the target state
// Parameters:
//
//	state: Target state
//
// Returns:
//
//	The from builder for method chaining
func (b *FromBuilder[S, E, P, Next]) To(state S) ToInterface[S, E, P] {
	b.targetId = state
	return (*FromBuilder[S, E, P, OnStep])(b)
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers these transitions
//
// Returns:
//
//	The from builder for method chaining
func (b *FromBuilder[S, E, P, Next]) On(event E) OnInterface[S, E, P] {
	b.event = event
	return (*FromBuilder[S, E, P, WhenStep])(b)
}

// When specifies the condition for all transitions
// Parameters:
//
//	condition: The condition that must be satisfied for the transitions to occur
//
// Returns:
//
//	The from builder for method chaining
func (b *FromBuilder[S, E, P, Next]) When(condition Condition[P]) WhenInterface[S, E, P] {
	b.condition = condition
	return (*FromBuilder[S, E, P, PerformStep])(b)
}

// WhenFunc specifies a function as the condition for all transitions
// Parameters:
//
//	conditionFunc: The function that must return true for the transitions to occur
//
// Returns:
//
//	The from builder for method chaining
func (b *FromBuilder[S, E, P, Next]) WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return (*FromBuilder[S, E, P, PerformStep])(b)
}

// Perform specifies the action to execute during all transitions
// Parameters:
//
//	action: The action to execute when the transitions occur
func (b *FromBuilder[S, E, P, Next]) Perform(action Action[S, E, P]) {
	b.action = action

	// Get or create target state
	targetState := b.stateMachine.GetState(b.targetId)

	// Create transitions from all source states
	for _, sourceId := range b.sourceIds {
		sourceState := b.stateMachine.GetState(sourceId)
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)
		transition.Condition = b.condition
		transition.Action = b.action
	}
}

// PerformFunc specifies a function as the action to execute during all transitions
// Parameters:
//
//	actionFunc: The function to execute when the transitions occur
func (b *FromBuilder[S, E, P, Next]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.action = ActionFunc[S, E, P](actionFunc)

	// Get or create target state
	targetState := b.stateMachine.GetState(b.targetId)

	// Create transitions from all source states
	for _, sourceId := range b.sourceIds {
		sourceState := b.stateMachine.GetState(sourceId)
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)
		transition.Condition = b.condition
		transition.Action = b.action
	}
}

// InternalTransitionBuilder builds internal transitions
type InternalTransitionBuilder[S comparable, E comparable, P any] struct {
	stateMachine *StateMachineImpl[S, E, P]
}

// Within specifies the state where the internal transition occurs
// Parameters:
//
//	state: The state where the internal transition occurs
//
// Returns:
//
//	The internal transition builder for method chaining
func (b *InternalTransitionBuilder[S, E, P]) Within(state S) ToInterface[S, E, P] {
	return &OnTransitionBuilder[S, E, P, OnStep]{
		stateMachine:   b.stateMachine,
		stateId:        state,
		transitionType: Internal,
	}
}

// OnTransitionBuilder builds the "on" part of an internal transition
type OnTransitionBuilder[S comparable, E comparable, P any, Next any] struct {
	stateMachine   *StateMachineImpl[S, E, P]
	stateId        S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
	transitionType TransitionType
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers this transition
//
// Returns:
//
//	The on transition builder for method chaining
func (b *OnTransitionBuilder[S, E, P, Next]) On(event E) OnInterface[S, E, P] {
	b.event = event
	return (*OnTransitionBuilder[S, E, P, WhenStep])(b)
}

// When specifies the condition for the transition
// Parameters:
//
//	condition: The condition that must be satisfied for the transition to occur
//
// Returns:
//
//	The on transition builder for method chaining
func (b *OnTransitionBuilder[S, E, P, Next]) When(condition Condition[P]) WhenInterface[S, E, P] {
	b.condition = condition
	return (*OnTransitionBuilder[S, E, P, PerformStep])(b)
}

// WhenFunc specifies a function as the condition for the transition
// Parameters:
//
//	conditionFunc: The function that must return true for the transition to occur
//
// Returns:
//
//	The on transition builder for method chaining
func (b *OnTransitionBuilder[S, E, P, Next]) WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return (*OnTransitionBuilder[S, E, P, PerformStep])(b)
}

// Perform specifies the action to execute during the transition
// Parameters:
//
//	action: The action to execute when the transition occurs
func (b *OnTransitionBuilder[S, E, P, Next]) Perform(action Action[S, E, P]) {
	b.action = action

	// Get or create state
	state := b.stateMachine.GetState(b.stateId)

	// Create internal transition
	transition := state.AddTransition(b.event, state, b.transitionType)
	transition.Condition = b.condition
	transition.Action = b.action
}

// PerformFunc specifies a function as the action to execute during the transition
// Parameters:
//
//	actionFunc: The function to execute when the transition occurs
func (b *OnTransitionBuilder[S, E, P, Next]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.action = ActionFunc[S, E, P](actionFunc)

	// Get or create state
	state := b.stateMachine.GetState(b.stateId)

	// Create internal transition
	transition := state.AddTransition(b.event, state, b.transitionType)
	transition.Condition = b.condition
	transition.Action = b.action
}
