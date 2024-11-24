package gfsm

import "fmt"

// StateMachineBuilder interface provides access to a builder that simplifies state machine creation. Builder usage is optional,
// and state machine object can be created manually if needed.
// Refer to newSmManual test for manual state machine creation or newSmWithBuilder as the alternative approach with builder.
type StateMachineBuilder[StateIdentifier comparable] interface {
	// RegisterState call register one more state referenced by stateID with list of all valid transactions listed in transitions
	// and handler (action) into the state machine.
	RegisterState(stateID StateIdentifier, action StateAction[StateIdentifier], transitions []StateIdentifier) StateMachineBuilder[StateIdentifier]
	// SetDefaultState tells which state is the default for the state machine. Each state machine must have a default state.
	// On StateMachineHandler.Start() call, state machine will switch to the defined default state.
	SetDefaultState(stateID StateIdentifier) StateMachineBuilder[StateIdentifier]
	// SetSmContext is an optional call that allow to pass any context that is unique and persistent (but mutable) for each state machine.
	SetSmContext(ctx StateMachineContext) StateMachineBuilder[StateIdentifier]

	// Build is the final call that aggregates all the data from previous calls and creates new state machine.
	Build() StateMachineHandler[StateIdentifier]
}

// NewBuilder function generates StateMachineBuilder which simplifies state machine creation process.
func NewBuilder[StateIdentifier comparable]() StateMachineBuilder[StateIdentifier] {
	return &stateMachineBuilder[StateIdentifier]{
		hasState: false,
		sm: &stateMachine[StateIdentifier]{
			states: StatesMap[StateIdentifier]{},
		},
	}
}

// stateMachineBuilder[StateIdentifier comparable] is an implementation for
// the StateMachineBuilder[StateIdentifier comparable] interface
type stateMachineBuilder[StateIdentifier comparable] struct {
	hasState        bool
	hasDefaultState bool

	sm *stateMachine[StateIdentifier]
}

func (s *stateMachineBuilder[StateIdentifier]) Build() StateMachineHandler[StateIdentifier] {
	if !s.hasState || !s.hasDefaultState {
		panic("state machine is not properly initialised yet")
	}
	return s.sm
}

func (s *stateMachineBuilder[StateIdentifier]) RegisterState(
	stateID StateIdentifier,
	action StateAction[StateIdentifier],
	transitions []StateIdentifier) StateMachineBuilder[StateIdentifier] {

	_, ok := s.sm.states[stateID]
	if ok {
		panic(fmt.Sprintf("state %v is already registered", stateID))
	}

	s.sm.states[stateID] = state[StateIdentifier]{
		action:      action,
		transitions: makeTransitions(transitions),
	}
	s.hasState = true

	return s
}

func makeTransitions[StateIdentifier comparable](transitions []StateIdentifier) Transitions[StateIdentifier] {
	trs := Transitions[StateIdentifier]{}
	for _, transition := range transitions {
		trs[transition] = struct{}{}
	}
	return trs
}

func (s *stateMachineBuilder[StateIdentifier]) SetDefaultState(stateID StateIdentifier) StateMachineBuilder[StateIdentifier] {
	s.sm.currentStateID = stateID
	s.sm.defaultStateID = stateID
	s.hasDefaultState = true

	return s
}

func (s *stateMachineBuilder[StateIdentifier]) SetSmContext(ctx StateMachineContext) StateMachineBuilder[StateIdentifier] {
	s.sm.smCtx = ctx

	return s
}
