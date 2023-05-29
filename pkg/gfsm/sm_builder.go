package gfsm

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
	s.sm.state = stateID
	s.hasDefaultState = true

	return s
}

func (s *stateMachineBuilder[StateIdentifier]) SetSmContext(ctx StateMachineContext) StateMachineBuilder[StateIdentifier] {
	s.sm.smCtx = ctx

	return s
}
