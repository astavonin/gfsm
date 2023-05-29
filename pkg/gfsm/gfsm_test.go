package gfsm

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type StartStopSM int

const (
	Start StartStopSM = iota
	Stop
	InProgress
)

type StartState struct {
}

func (s StartState) OnEnter(smCtx StateMachineContext) {
	fmt.Println("StartState.OnEnter")
}

func (s StartState) OnExit(smCtx StateMachineContext) {
	fmt.Println("StartState.OnExit")
}

func (s StartState) Execute(smCtx StateMachineContext, eventCtx EventContext) StartStopSM {
	fmt.Println("StartState.Execute")
	return InProgress
}

type StopState struct {
}

func (s StopState) OnEnter(smCtx StateMachineContext) {
	fmt.Println("StopState.OnEnter")
}

func (s StopState) OnExit(smCtx StateMachineContext) {
	fmt.Println("StopState.OnExit")
}

func (s StopState) Execute(smCtx StateMachineContext, eventCtx EventContext) StartStopSM {
	fmt.Println("StopState.Execute")
	return Start
}

type InProgressState struct {
}

func (i InProgressState) OnEnter(smCtx StateMachineContext) {
	fmt.Println("InProgressState.OnEnter")
}

func (i InProgressState) OnExit(smCtx StateMachineContext) {
	fmt.Println("InProgressState.OnExit")
}

func (i InProgressState) Execute(smCtx StateMachineContext, eventCtx EventContext) StartStopSM {
	fmt.Println("InProgressState.Execute")
	return InProgress
}

type StartData struct {
	id uuid.UUID
}

func TestLightSwitchStateMachine(t *testing.T) {
	sm := StateMachine[StartStopSM]{
		state: Start,
		states: States[StartStopSM]{
			Start: State[StartStopSM]{
				action:      &StartState{},
				transitions: []StartStopSM{Stop, InProgress},
			},
			Stop: State[StartStopSM]{
				action:      &StopState{},
				transitions: []StartStopSM{Start},
			},
			InProgress: State[StartStopSM]{
				action:      &InProgressState{},
				transitions: []StartStopSM{Stop},
			},
		},
		smCtx: nil,
	}
	sm.Start()
	defer sm.Stop()
	sm.ProcessEvent(StartData{id: uuid.New()})
}
