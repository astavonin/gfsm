package gfsm

import "fmt"

var (
	ErrNoValidTransition = fmt.Errorf("no valid transition")
)

type Transitions[StateIdentifier comparable] map[StateIdentifier]struct{}

type state[StateIdentifier comparable] struct {
	action      StateAction[StateIdentifier]
	transitions Transitions[StateIdentifier]
}

type States[StateIdentifier comparable] map[StateIdentifier]state[StateIdentifier]

type StateMachineHandler[StateIdentifier comparable] interface {
	Start()
	Stop()

	State() StateIdentifier

	ProcessEvent(eventCtx EventContext) error
}

type StateMachineBuilder[StateIdentifier comparable] interface {
	RegisterState(stateID StateIdentifier, action StateAction[StateIdentifier], transitions []StateIdentifier) StateMachineBuilder[StateIdentifier]
	SetDefaultState(stateID StateIdentifier) StateMachineBuilder[StateIdentifier]
	SetSmContext(ctx StateMachineContext) StateMachineBuilder[StateIdentifier]

	Build() StateMachineHandler[StateIdentifier]
}

func newBuilder[StateIdentifier comparable]() StateMachineBuilder[StateIdentifier] {
	return &stateMachineBuilder[StateIdentifier]{
		hasState: false,
		sm: &stateMachine[StateIdentifier]{
			states: States[StateIdentifier]{},
		},
	}
}

type stateMachine[StateIdentifier comparable] struct {
	state  StateIdentifier
	states States[StateIdentifier]
	smCtx  StateMachineContext
}

func (s *stateMachine[StateIdentifier]) Start() {
	state := s.states[s.state]
	state.action.OnEnter(s.smCtx)
}

func (s *stateMachine[StateIdentifier]) Stop() {
	state := s.states[s.state]
	state.action.OnExit(s.smCtx)
}

func (s *stateMachine[StateIdentifier]) State() StateIdentifier {
	return s.state
}

func (s *stateMachine[StateIdentifier]) ProcessEvent(eventCtx EventContext) error {
	state := s.states[s.state]
	newStateIdentifier := state.action.Execute(s.smCtx, eventCtx)
	// do not need to change state
	if newStateIdentifier == s.state {
		return nil
	}

	// need to change state
	_, found := state.transitions[newStateIdentifier]
	if !found {
		return fmt.Errorf("cannot switch from %v to %v: %w", s.state, newStateIdentifier, ErrNoValidTransition)
	}
	s.state = newStateIdentifier
	state.action.OnExit(s.smCtx)
	newState := s.states[newStateIdentifier]
	newState.action.OnEnter(s.smCtx)

	return nil
}
