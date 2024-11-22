// Package gfsm implements basic state machine functionality.
package gfsm

import "fmt"

var (
	ErrNoValidTransition = fmt.Errorf("no valid transition")
)

// Transitions represents all available transitions from the state.
type Transitions[StateIdentifier comparable] map[StateIdentifier]struct{}

// The state is a struct that is defined by the StateActions it can take, the Transitions it can make
type state[StateIdentifier comparable] struct {
	action      StateAction[StateIdentifier]
	transitions Transitions[StateIdentifier]
}

// StatesMap represent full state machine transactions and allows to verify path from any state to another.
// It is a map of StateIdentifiers to state
type StatesMap[StateIdentifier comparable] map[StateIdentifier]state[StateIdentifier]

// StateMachineHandler is the main state machine interface. All manipulation with the state machine object shall
// be performed using this interface.
type StateMachineHandler[StateIdentifier comparable] interface {
	// Start is the first function that user MUST call before any further interactions with the state machine.
	// On Start call, state machine will switch to the defined default state, which must be specified during state
	// machine creation using StateMachineBuilder.SetDefaultState(...) call
	Start()
	// Stop call shutdowns the state machine. Any further State or ProcessEvent are not permitted on stopped
	// state machine.
	Stop()

	// State returns current state machine state
	State() StateIdentifier

	// ProcessEvent pass data to the sate machine for processing. The data will be forwarded to StateAction.Execute
	// method of the current state. If the event processing will lead to unexpected transaction, ProcessEvent call
	// will return ErrNoValidTransition error
	ProcessEvent(eventCtx EventContext) error

	// Reset will return the statemachine to its default state
	Reset()
}

type stateMachine[StateIdentifier comparable] struct {
	currentStateID StateIdentifier
	defaultStateID StateIdentifier
	states         StatesMap[StateIdentifier]
	smCtx          StateMachineContext
}

func (s *stateMachine[StateIdentifier]) Start() {
	state := s.states[s.currentStateID]
	state.action.OnEnter(s.smCtx)
}

func (s *stateMachine[StateIdentifier]) Stop() {
	state := s.states[s.currentStateID]
	state.action.OnExit(s.smCtx)
}

func (s *stateMachine[StateIdentifier]) State() StateIdentifier {
	return s.currentStateID
}

func (s *stateMachine[StateIdentifier]) ProcessEvent(eventCtx EventContext) error {
	currentState := s.states[s.currentStateID]
	nextStateID := currentState.action.Execute(s.smCtx, eventCtx)
	// do not need to change state
	if nextStateID == s.currentStateID {
		return nil
	}

	return s.ChangeState(nextStateID)
}

func (s *stateMachine[StateIdentifier]) ChangeState(nextStateID StateIdentifier) error {
	currentState := s.states[s.currentStateID]
	_, canSwitch := currentState.transitions[nextStateID]
	if !canSwitch {
		return fmt.Errorf("cannot switch from %v to %v: %w", s.currentStateID, nextStateID, ErrNoValidTransition)
	}
	s.currentStateID = nextStateID
	currentState.action.OnExit(s.smCtx)
	nextState := s.states[nextStateID]
	nextState.action.OnEnter(s.smCtx)

	return nil
}

func (s *stateMachine[StateIdentifier]) Reset() {
	currentState := s.states[s.currentStateID]
	currentState.action.OnExit(s.smCtx)

	defaultState := s.states[s.defaultStateID]
	defaultState.action.OnEnter(s.smCtx)

	s.currentStateID = s.defaultStateID
}
