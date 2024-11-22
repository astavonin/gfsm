package gfsm

// EventContext is an abstraction that represent any data that user passes to the current state for execution
// in StateMachineHandler.ProcessEvent(...) call. The data will be forwarded as StateAction.Execute(...) argument.
type EventContext interface{}

// StateMachineContext is an abstraction that represent any data that user passes to the state machine. The data will
// be forwarded as StateAction.OnEnter amd OnExit arguments.
type StateMachineContext interface{}

// StateAction is the interface which each state must implement.
type StateAction[StateIdentifier comparable] interface {
	// OnEnter will be called once on the state entering.
	OnEnter(smCtx StateMachineContext)
	// OnExit will be called once on the state exiting.
	OnExit(smCtx StateMachineContext)
	// Execute is the call that state machine routes to the current state from StateMachineHandler.ProcessEvent(...)
	Execute(smCtx StateMachineContext, eventCtx EventContext) StateIdentifier
}
