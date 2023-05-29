package gfsm

import "fmt"

type EventContext interface{}
type StateMachineContext interface{}

type StateAction[StateID comparable] interface {
	OnEnter(smCtx StateMachineContext)
	OnExit(smCtx StateMachineContext)
	Execute(smCtx StateMachineContext, eventCtx EventContext) StateID
}

type Transitions[StateID comparable] []StateID

type State[StateID comparable] struct {
	action      StateAction[StateID]
	transitions Transitions[StateID]
}

type States[StateID comparable] map[StateID]State[StateID]

type StateMachine[StateID comparable] struct {
	state  StateID
	states States[StateID]
	smCtx  StateMachineContext
}

func (s *StateMachine[StateID]) Start() {
	state := s.states[s.state]
	state.action.OnEnter(s.smCtx)
}

func (s *StateMachine[StateID]) Stop() {
	state := s.states[s.state]
	state.action.OnExit(s.smCtx)
}

func (s *StateMachine[StateID]) ProcessEvent(eventCtx EventContext) {
	state := s.states[s.state]
	newStateID := state.action.Execute(s.smCtx, eventCtx)
	found := false
	for _, transition := range state.transitions {
		if newStateID == transition {
			found = true
			break
		}
	}
	if !found {
		fmt.Println("invalid transaction")
		return
	}
	if newStateID == s.state {
		fmt.Println("same state, no transition")
		return
	}
	s.state = newStateID
	state.action.OnExit(s.smCtx)
	newState := s.states[newStateID]
	newState.action.OnEnter(s.smCtx)
}
