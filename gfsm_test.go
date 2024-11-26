package gfsm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type StartStopSM int

const (
	Start StartStopSM = iota
	Stop
	InProgress
)

type StartState struct {
	t *testing.T
}

func (s *StartState) OnEnter(smCtx StateMachineContext) {
	s.t = smCtx.(*aContext).t
}

func (s *StartState) OnExit(_ StateMachineContext) {
}

func (s *StartState) Execute(_ StateMachineContext, eventCtx EventContext) StartStopSM {
	assert.NotNil(s.t, eventCtx)
	_, ok := eventCtx.(StartData)
	assert.True(s.t, ok)

	return InProgress
}

type StopState struct {
	t *testing.T
}

func (s *StopState) OnEnter(smCtx StateMachineContext) {
	s.t = smCtx.(*aContext).t
}

func (s *StopState) OnExit(_ StateMachineContext) {
}

func (s *StopState) Execute(_ StateMachineContext, _ EventContext) StartStopSM {
	return Start
}

type InProgressState struct {
	t *testing.T
}

func (s *InProgressState) OnEnter(smCtx StateMachineContext) {
	s.t = smCtx.(*aContext).t
}

func (s *InProgressState) OnExit(_ StateMachineContext) {
}

func (s *InProgressState) Execute(_ StateMachineContext, eventCtx EventContext) StartStopSM {
	assert.NotNil(s.t, eventCtx)
	_, ok := eventCtx.(InProgressData)
	assert.True(s.t, ok)

	return InProgress
}

type StartData struct {
	id uuid.UUID
}

type InProgressData struct {
}

type aContext struct {
	t *testing.T
}

func newSmManual(t *testing.T) StateMachineHandler[StartStopSM] {
	return &stateMachine[StartStopSM]{
		currentStateID: Start,
		states: StatesMap[StartStopSM]{
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
		smCtx: &aContext{t: t},
	}
}

func newSmWithBuilder(t *testing.T) StateMachineHandler[StartStopSM] {
	return NewBuilder[StartStopSM]().
		SetDefaultState(Start).
		SetSmContext(&aContext{t: t}).
		RegisterState(Start, &StartState{}, []StartStopSM{Stop, InProgress}).
		RegisterState(Stop, &StopState{}, []StartStopSM{Start}).
		RegisterState(InProgress, &InProgressState{}, []StartStopSM{Stop}).
		Build()
}

func TestStateMachine(t *testing.T) {
	var tests = []struct {
		sm       StateMachineHandler[StartStopSM]
		testName string
	}{
		{newSmManual(t), "manual SM creation"},
		{newSmWithBuilder(t), "builder SM creation"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			sm := test.sm

			sm.Start()

			assert.Equal(t, sm.State(), Start)
			err := sm.ProcessEvent(StartData{id: uuid.New()})
			assert.NoError(t, err)
			assert.Equal(t, sm.State(), InProgress)
			err = sm.ProcessEvent(InProgressData{})
			assert.NoError(t, err)
			assert.Equal(t, sm.State(), InProgress)

			sm.Stop()
		})
	}
}

func TestDoubleStateCreation(t *testing.T) {
	builder := NewBuilder[StartStopSM]().
		RegisterState(Stop, &StopState{}, []StartStopSM{Stop})

	assert.Panics(t, func() {
		builder.RegisterState(Stop, &StopState{}, []StartStopSM{Stop})
	})
}

func TestResetStatMachine(t *testing.T) {
	sm := newSmWithBuilder(t)

	sm.Start()
	err := sm.ProcessEvent(StartData{id: uuid.New()})
	assert.NoError(t, err)
	assert.Equal(t, sm.State(), InProgress)

	sm.Reset()
	assert.Equal(t, sm.State(), Start)

	sm.Stop()
}
