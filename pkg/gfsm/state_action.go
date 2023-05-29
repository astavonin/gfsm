package gfsm

type EventContext interface{}
type StateMachineContext interface{}

type StateAction[StateID comparable] interface {
	OnEnter(smCtx StateMachineContext)
	OnExit(smCtx StateMachineContext)
	Execute(smCtx StateMachineContext, eventCtx EventContext) StateID
}
