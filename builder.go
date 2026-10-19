package fsm

import "fmt"

// Step markers are phantom type parameters recording where a builder is in the method chain
type fromStep interface{}
type toStep interface{}
type onStep interface{}
type whenStep interface{}
type performStep interface{}
type toAmongStep interface{}

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

	// WhenInterface allows calling Perform or PerformFunc directly for a transition without condition
	WhenInterface[S, E, P]
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
	_ ExternalTransitionBuilderInterface[string, string, any]         = (*transitionBuilder[string, string, any, fromStep])(nil)
	_ ExternalTransitionsBuilderInterface[string, string, any]        = (*externalTransitionsBuilder[string, string, any])(nil)
	_ ExternalParallelTransitionBuilderInterface[string, string, any] = (*externalParallelTransitionBuilder[string, string, any])(nil)
	_ InternalTransitionBuilderInterface[string, string, any]         = (*internalTransitionBuilder[string, string, any])(nil)
	_ ParallelFromInterface[string, string, any]                      = (*parallelFromBuilder[string, string, any, toAmongStep])(nil)
	_ ToInterface[string, string, any]                                = (*parallelFromBuilder[string, string, any, onStep])(nil)
	_ OnInterface[string, string, any]                                = (*parallelFromBuilder[string, string, any, whenStep])(nil)
	_ WhenInterface[string, string, any]                              = (*parallelFromBuilder[string, string, any, performStep])(nil)
	_ FromInterface[string, string, any]                              = (*fromBuilder[string, string, any, toStep])(nil)
	_ ToInterface[string, string, any]                                = (*fromBuilder[string, string, any, onStep])(nil)
	_ OnInterface[string, string, any]                                = (*fromBuilder[string, string, any, whenStep])(nil)
	_ WhenInterface[string, string, any]                              = (*fromBuilder[string, string, any, performStep])(nil)
	_ ToInterface[string, string, any]                                = (*onTransitionBuilder[string, string, any, onStep])(nil)
	_ OnInterface[string, string, any]                                = (*onTransitionBuilder[string, string, any, whenStep])(nil)
	_ WhenInterface[string, string, any]                              = (*onTransitionBuilder[string, string, any, performStep])(nil)
)

// StateMachineBuilder builds state machines with a fluent API
type StateMachineBuilder[S comparable, E comparable, P any] struct {
	stateMachine *stateMachineImpl[S, E, P]
	chains       []transitionChain
}

// transitionChain is a transition definition started on the builder
type transitionChain interface {
	// incomplete describes the definition if Perform was never called, or returns "" otherwise
	incomplete() string
}

// track records a transition definition so Build can reject it if it never reaches Perform
func (b *StateMachineBuilder[S, E, P]) track(chain transitionChain) {
	b.chains = append(b.chains, chain)
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
	chain := &transitionBuilder[S, E, P, fromStep]{
		stateMachine:   b.stateMachine,
		transitionType: external,
	}
	b.track(chain)
	return chain
}

// InternalTransition starts defining an internal transition
// Returns:
//
//	A internal transition builder for configuring the internal transition
func (b *StateMachineBuilder[S, E, P]) InternalTransition() InternalTransitionBuilderInterface[S, E, P] {
	return &internalTransitionBuilder[S, E, P]{
		builder: b,
	}
}

// ExternalTransitions starts defining multiple external transitions from different source states to the same target state
// Returns:
//
//	A external transitions builder for configuring the transitions
func (b *StateMachineBuilder[S, E, P]) ExternalTransitions() ExternalTransitionsBuilderInterface[S, E, P] {
	return &externalTransitionsBuilder[S, E, P]{
		builder: b,
	}
}

// ExternalParallelTransition starts defining an external parallel transition
// Returns:
//
//	A external parallel transition builder for configuring the parallel transition
func (b *StateMachineBuilder[S, E, P]) ExternalParallelTransition() ExternalParallelTransitionBuilderInterface[S, E, P] {
	return &externalParallelTransitionBuilder[S, E, P]{
		builder: b,
	}
}

// Build finalizes the state machine with the given ID
// The machine is not registered globally; call RegisterStateMachine to make it
// available through GetStateMachine.
// Parameters:
//
//	machineId: Identifier of the state machine, shown in its string and diagram output
//
// Returns:
//
//	The built state machine and possible error
func (b *StateMachineBuilder[S, E, P]) Build(machineId string) (StateMachine[S, E, P], error) {
	// Validate the state machine configuration
	if err := b.validate(); err != nil {
		return nil, err
	}

	// Build an independent copy so later builder calls can't modify the returned machine
	return b.stateMachine.snapshot(machineId), nil
}

// validate checks the state machine configuration for errors
func (b *StateMachineBuilder[S, E, P]) validate() error {
	// A definition that never reached Perform would otherwise be silently dropped
	for _, chain := range b.chains {
		if description := chain.incomplete(); description != "" {
			return fmt.Errorf("%w: %s", ErrIncompleteTransition, description)
		}
	}

	// Check if any transitions are defined
	if len(b.stateMachine.stateMap) == 0 {
		return ErrNoTransitionsDefined
	}

	// Walk in sorted order so the first error reported is stable
	for _, stateId := range sortedKeys(b.stateMachine.stateMap) {
		state := b.stateMachine.stateMap[stateId]
		for _, event := range sortedKeys(state.eventTransitions) {
			if err := validateTransitions(stateId, event, state.eventTransitions[event]); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateTransitions checks the transitions defined for one source state and event
func validateTransitions[S comparable, E comparable, P any](source S, event E, transitions []*transitionImpl[S, E, P]) error {
	parallel := transitions[0].parallel
	seen := make(map[S]bool, len(transitions))
	hasFallback := false
	for _, t := range transitions {
		if t.parallel != parallel {
			return fmt.Errorf("%w: %v -[%v]->", ErrMixedTransitions, source, event)
		}

		// As in COLA, a source, event and target allow only one transition, whatever its condition
		if seen[t.target.id] {
			return fmt.Errorf("%w: %v -[%v]-> %v", ErrDuplicateTransition, source, event, t.target.id)
		}
		seen[t.target.id] = true

		// FireEvent falls back to the unconditional transition, so there can be only one
		if !parallel && isUnconditional(t.condition) {
			if hasFallback {
				return fmt.Errorf("%w: %v -[%v]-> %v", ErrAmbiguousTransition, source, event, t.target.id)
			}
			hasFallback = true
		}
	}

	return nil
}

// externalTransitionsBuilder builds external transitions from multiple source states to a single target state
type externalTransitionsBuilder[S comparable, E comparable, P any] struct {
	builder *StateMachineBuilder[S, E, P]
}

// FromAmong specifies multiple source states
// Parameters:
//
//	states: Multiple source states
//
// Returns:
//
//	The from builder for method chaining
func (b *externalTransitionsBuilder[S, E, P]) FromAmong(states ...S) FromInterface[S, E, P] {
	chain := &fromBuilder[S, E, P, toStep]{
		stateMachine:   b.builder.stateMachine,
		sourceIds:      states,
		transitionType: external,
	}
	b.builder.track(chain)
	return chain
}

// externalParallelTransitionBuilder builds external parallel transitions
type externalParallelTransitionBuilder[S comparable, E comparable, P any] struct {
	builder *StateMachineBuilder[S, E, P]
}

// From specifies the source state
// Parameters:
//
//	state: Source state
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *externalParallelTransitionBuilder[S, E, P]) From(state S) ParallelFromInterface[S, E, P] {
	chain := &parallelFromBuilder[S, E, P, toAmongStep]{
		stateMachine:   b.builder.stateMachine,
		sourceId:       state,
		transitionType: external,
	}
	b.builder.track(chain)
	return chain
}

// parallelFromBuilder builds the "from" part of a parallel transition
type parallelFromBuilder[S comparable, E comparable, P any, Next any] struct {
	stateMachine   *stateMachineImpl[S, E, P]
	transitionType transitionType
	sourceId       S
	targetIds      []S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
	done           bool
}

// ToAmong specifies multiple target states
// Parameters:
//
//	states: Multiple target states
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *parallelFromBuilder[S, E, P, Next]) ToAmong(states ...S) ToInterface[S, E, P] {
	b.targetIds = states
	return (*parallelFromBuilder[S, E, P, onStep])(b)
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers these transitions
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *parallelFromBuilder[S, E, P, Next]) On(event E) OnInterface[S, E, P] {
	b.event = event
	return (*parallelFromBuilder[S, E, P, whenStep])(b)
}

// When specifies the condition for all transitions
// Parameters:
//
//	condition: The condition that must be satisfied for the transitions to occur
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *parallelFromBuilder[S, E, P, Next]) When(condition Condition[P]) WhenInterface[S, E, P] {
	b.condition = condition
	return (*parallelFromBuilder[S, E, P, performStep])(b)
}

// WhenFunc specifies a function as the condition for all transitions
// Parameters:
//
//	conditionFunc: The function that must return true for the transitions to occur
//
// Returns:
//
//	The parallel from builder for method chaining
func (b *parallelFromBuilder[S, E, P, Next]) WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return (*parallelFromBuilder[S, E, P, performStep])(b)
}

// Perform specifies the action to execute during all transitions
// Parameters:
//
//	action: The action to execute when the transitions occur
func (b *parallelFromBuilder[S, E, P, Next]) Perform(action Action[S, E, P]) {
	b.action = action
	b.done = true

	// Get or create source state
	sourceState := b.stateMachine.getOrCreateState(b.sourceId)

	// Create transitions to all target states
	for _, targetId := range b.targetIds {
		targetState := b.stateMachine.getOrCreateState(targetId)
		transition := sourceState.addTransition(b.event, targetState, b.transitionType)
		transition.condition = b.condition
		transition.action = b.action
		transition.parallel = true
	}
}

// PerformFunc specifies a function as the action to execute during all transitions
// Parameters:
//
//	actionFunc: The function to execute when the transitions occur
func (b *parallelFromBuilder[S, E, P, Next]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.Perform(ActionFunc[S, E, P](actionFunc))
}

// incomplete implements transitionChain
func (b *parallelFromBuilder[S, E, P, Next]) incomplete() string {
	if b.done {
		return ""
	}
	return fmt.Sprintf("%v -[%v]-> %v", b.sourceId, b.event, b.targetIds)
}

// transitionBuilder builds individual transitions
type transitionBuilder[S comparable, E comparable, P any, Next any] struct {
	stateMachine   *stateMachineImpl[S, E, P]
	transitionType transitionType
	sourceId       S
	targetId       S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
	done           bool
}

// From specifies the source state
// Parameters:
//
//	state: Source state
//
// Returns:
//
//	The transition builder for method chaining
func (b *transitionBuilder[S, E, P, Next]) From(state S) FromInterface[S, E, P] {
	b.sourceId = state
	return (*transitionBuilder[S, E, P, toStep])(b)
}

// To specifies the target state
// Parameters:
//
//	state: Target state
//
// Returns:
//
//	The transition builder for method chaining
func (b *transitionBuilder[S, E, P, Next]) To(state S) ToInterface[S, E, P] {
	b.targetId = state
	return (*transitionBuilder[S, E, P, onStep])(b)
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers this transition
//
// Returns:
//
//	The transition builder for method chaining
func (b *transitionBuilder[S, E, P, Next]) On(event E) OnInterface[S, E, P] {
	b.event = event
	return (*transitionBuilder[S, E, P, whenStep])(b)
}

// When specifies the condition for the transition
// Parameters:
//
//	condition: The condition that must be satisfied for the transition to occur
//
// Returns:
//
//	The transition builder for method chaining
func (b *transitionBuilder[S, E, P, Next]) When(condition Condition[P]) WhenInterface[S, E, P] {
	b.condition = condition
	return (*transitionBuilder[S, E, P, performStep])(b)
}

// WhenFunc specifies a function as the condition for the transition
// Parameters:
//
//	conditionFunc: The function that must return true for the transition to occur
//
// Returns:
//
//	The transition builder for method chaining
func (b *transitionBuilder[S, E, P, Next]) WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return (*transitionBuilder[S, E, P, performStep])(b)
}

// Perform specifies the action to execute during the transition
// Parameters:
//
//	action: The action to execute when the transition occurs
func (b *transitionBuilder[S, E, P, Next]) Perform(action Action[S, E, P]) {
	b.action = action
	b.done = true

	// Get or create states
	sourceState := b.stateMachine.getOrCreateState(b.sourceId)
	targetState := b.stateMachine.getOrCreateState(b.targetId)

	// Create transition
	transition := sourceState.addTransition(b.event, targetState, b.transitionType)
	transition.condition = b.condition
	transition.action = b.action
}

// PerformFunc specifies a function as the action to execute during the transition
// Parameters:
//
//	actionFunc: The function to execute when the transition occurs
func (b *transitionBuilder[S, E, P, Next]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.Perform(ActionFunc[S, E, P](actionFunc))
}

// incomplete implements transitionChain
func (b *transitionBuilder[S, E, P, Next]) incomplete() string {
	if b.done {
		return ""
	}
	return fmt.Sprintf("%v -[%v]-> %v", b.sourceId, b.event, b.targetId)
}

// fromBuilder builds the "from" part of multiple transitions
type fromBuilder[S comparable, E comparable, P any, Next any] struct {
	stateMachine   *stateMachineImpl[S, E, P]
	transitionType transitionType
	sourceIds      []S
	targetId       S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
	done           bool
}

// To specifies the target state
// Parameters:
//
//	state: Target state
//
// Returns:
//
//	The from builder for method chaining
func (b *fromBuilder[S, E, P, Next]) To(state S) ToInterface[S, E, P] {
	b.targetId = state
	return (*fromBuilder[S, E, P, onStep])(b)
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers these transitions
//
// Returns:
//
//	The from builder for method chaining
func (b *fromBuilder[S, E, P, Next]) On(event E) OnInterface[S, E, P] {
	b.event = event
	return (*fromBuilder[S, E, P, whenStep])(b)
}

// When specifies the condition for all transitions
// Parameters:
//
//	condition: The condition that must be satisfied for the transitions to occur
//
// Returns:
//
//	The from builder for method chaining
func (b *fromBuilder[S, E, P, Next]) When(condition Condition[P]) WhenInterface[S, E, P] {
	b.condition = condition
	return (*fromBuilder[S, E, P, performStep])(b)
}

// WhenFunc specifies a function as the condition for all transitions
// Parameters:
//
//	conditionFunc: The function that must return true for the transitions to occur
//
// Returns:
//
//	The from builder for method chaining
func (b *fromBuilder[S, E, P, Next]) WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return (*fromBuilder[S, E, P, performStep])(b)
}

// Perform specifies the action to execute during all transitions
// Parameters:
//
//	action: The action to execute when the transitions occur
func (b *fromBuilder[S, E, P, Next]) Perform(action Action[S, E, P]) {
	b.action = action
	b.done = true

	// Get or create target state
	targetState := b.stateMachine.getOrCreateState(b.targetId)

	// Create transitions from all source states
	for _, sourceId := range b.sourceIds {
		sourceState := b.stateMachine.getOrCreateState(sourceId)
		transition := sourceState.addTransition(b.event, targetState, b.transitionType)
		transition.condition = b.condition
		transition.action = b.action
	}
}

// PerformFunc specifies a function as the action to execute during all transitions
// Parameters:
//
//	actionFunc: The function to execute when the transitions occur
func (b *fromBuilder[S, E, P, Next]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.Perform(ActionFunc[S, E, P](actionFunc))
}

// incomplete implements transitionChain
func (b *fromBuilder[S, E, P, Next]) incomplete() string {
	if b.done {
		return ""
	}
	return fmt.Sprintf("%v -[%v]-> %v", b.sourceIds, b.event, b.targetId)
}

// internalTransitionBuilder builds internal transitions
type internalTransitionBuilder[S comparable, E comparable, P any] struct {
	builder *StateMachineBuilder[S, E, P]
}

// Within specifies the state where the internal transition occurs
// Parameters:
//
//	state: The state where the internal transition occurs
//
// Returns:
//
//	The internal transition builder for method chaining
func (b *internalTransitionBuilder[S, E, P]) Within(state S) ToInterface[S, E, P] {
	chain := &onTransitionBuilder[S, E, P, onStep]{
		stateMachine:   b.builder.stateMachine,
		stateId:        state,
		transitionType: internal,
	}
	b.builder.track(chain)
	return chain
}

// onTransitionBuilder builds the "on" part of an internal transition
type onTransitionBuilder[S comparable, E comparable, P any, Next any] struct {
	stateMachine   *stateMachineImpl[S, E, P]
	stateId        S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
	transitionType transitionType
	done           bool
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers this transition
//
// Returns:
//
//	The on transition builder for method chaining
func (b *onTransitionBuilder[S, E, P, Next]) On(event E) OnInterface[S, E, P] {
	b.event = event
	return (*onTransitionBuilder[S, E, P, whenStep])(b)
}

// When specifies the condition for the transition
// Parameters:
//
//	condition: The condition that must be satisfied for the transition to occur
//
// Returns:
//
//	The on transition builder for method chaining
func (b *onTransitionBuilder[S, E, P, Next]) When(condition Condition[P]) WhenInterface[S, E, P] {
	b.condition = condition
	return (*onTransitionBuilder[S, E, P, performStep])(b)
}

// WhenFunc specifies a function as the condition for the transition
// Parameters:
//
//	conditionFunc: The function that must return true for the transition to occur
//
// Returns:
//
//	The on transition builder for method chaining
func (b *onTransitionBuilder[S, E, P, Next]) WhenFunc(conditionFunc func(payload P) bool) WhenInterface[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return (*onTransitionBuilder[S, E, P, performStep])(b)
}

// Perform specifies the action to execute during the transition
// Parameters:
//
//	action: The action to execute when the transition occurs
func (b *onTransitionBuilder[S, E, P, Next]) Perform(action Action[S, E, P]) {
	b.action = action
	b.done = true

	// Get or create state
	state := b.stateMachine.getOrCreateState(b.stateId)

	// Create internal transition
	transition := state.addTransition(b.event, state, b.transitionType)
	transition.condition = b.condition
	transition.action = b.action
}

// PerformFunc specifies a function as the action to execute during the transition
// Parameters:
//
//	actionFunc: The function to execute when the transition occurs
func (b *onTransitionBuilder[S, E, P, Next]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.Perform(ActionFunc[S, E, P](actionFunc))
}

// incomplete implements transitionChain
func (b *onTransitionBuilder[S, E, P, Next]) incomplete() string {
	if b.done {
		return ""
	}
	return fmt.Sprintf("%v -[%v]-> %v", b.stateId, b.event, b.stateId)
}
