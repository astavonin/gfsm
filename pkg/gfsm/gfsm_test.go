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

func TestStateMachineManual(t *testing.T) {
	sm := stateMachine[StartStopSM]{
		state: Start,
		states: States[StartStopSM]{
			Start: state[StartStopSM]{
				action: &StartState{},
				transitions: Transitions[StartStopSM]{
					Stop:       struct{}{},
					InProgress: struct{}{},
				},
			},
			Stop: state[StartStopSM]{
				action: &StopState{},
				transitions: Transitions[StartStopSM]{
					Start: struct{}{},
				},
			},
			InProgress: state[StartStopSM]{
				action: &InProgressState{},
				transitions: Transitions[StartStopSM]{
					Stop: struct{}{},
				},
			},
		},
		smCtx: nil,
	}
	sm.Start()
	defer sm.Stop()
	sm.ProcessEvent(StartData{id: uuid.New()})
}

func TestStateMachineBuilder(t *testing.T) {
	sm := newBuilder[StartStopSM]().
		SetDefaultState(Start).
		RegisterState(Start, &StartState{}, []StartStopSM{Stop, InProgress}).
		RegisterState(Stop, &StopState{}, []StartStopSM{Stop}).
		RegisterState(InProgress, &InProgressState{}, []StartStopSM{Stop}).
		Build()

	sm.Start()
	defer sm.Stop()
	sm.ProcessEvent(StartData{id: uuid.New()})
}

func BenchmarkSliceAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = sliceTest()
	}
}

func BenchmarkMapAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mapTest()
	}
}

func sliceTest() bool {
	transitions := map[StartStopSM]struct{}{
		Stop:       {},
		Start:      {},
		InProgress: {},
	}
	_, found := transitions[Stop]
	return found
}

func mapTest() bool {
	transitions := []StartStopSM{
		Start, Stop, InProgress,
	}
	found := false
	for _, transition := range transitions {
		if Stop == transition {
			found = true
			break
		}
	}
	return found
}
