package main

import (
	"fmt"
	gfsm2 "github.com/astavonin/gfsm"
)

//go:generate stringer -type=State
type State int

const (
	Init State = iota
	Wait
	Abort
	Commit
)

type coordinatorContext struct {
	commitID string
	partCnt  int
}

// ========= Init state handler =========
type initState struct {
}

type commitRequest struct {
	commitID string
}

func (s *initState) OnEnter(_ gfsm2.StateMachineContext) {
}

func (s *initState) OnExit(_ gfsm2.StateMachineContext) {
}

func (s *initState) Execute(smCtx gfsm2.StateMachineContext, eventCtx gfsm2.EventContext) State {
	cCtx := smCtx.(*coordinatorContext)
	req, ok := eventCtx.(commitRequest)
	if !ok {
		fmt.Printf("invalid request\n")
		// nothing to process, keeping state
		return Init
	}

	fmt.Printf("got request %s\n", req.commitID)
	// forwarding request to participants and switching the state
	//for i := 0; i < cCtx.votesCnt; i++ {
	//	...
	//}
	// and saving commit ID as the state machine context
	cCtx.commitID = req.commitID

	return Wait
}

// ========= Wait state handler =========

type waitState struct {
	votesCnt int
}

type commitVote struct {
	commit bool
}

func (s *waitState) OnEnter(_ gfsm2.StateMachineContext) {
	s.votesCnt = 0
}

func (s *waitState) OnExit(_ gfsm2.StateMachineContext) {
}

func (s *waitState) Execute(smCtx gfsm2.StateMachineContext, eventCtx gfsm2.EventContext) State {
	cCtx := smCtx.(*coordinatorContext)
	vote, ok := eventCtx.(commitVote)
	if !ok || !vote.commit {
		fmt.Printf("invalid vote or vote for commit %s was rejected\n", cCtx.commitID)
		// rejecting commit
		return Abort
	}

	fmt.Printf("one more commit confirmation for %s!\n", cCtx.commitID)
	s.votesCnt++
	if s.votesCnt == cCtx.partCnt {
		// all votes were positive, committing
		fmt.Printf("all participants confirmed commit %s!\n", cCtx.commitID)
		return Commit
	}

	return Wait
}

// ========= Commit/Abort state handler =========

type responseState struct {
	votesCnt int
	keepResp State
}

func (s *responseState) OnEnter(smCtx gfsm2.StateMachineContext) {
	cCtx := smCtx.(*coordinatorContext)
	s.votesCnt = cCtx.partCnt
	fmt.Printf("commiting %s\n", cCtx.commitID)
	//for i := 0; i < cCtx.votesCnt; i++ {
	//	sending commit/abort message to each participant
	//}
}

func (s *responseState) OnExit(_ gfsm2.StateMachineContext) {
}

func (s *responseState) Execute(_ gfsm2.StateMachineContext, eventCtx gfsm2.EventContext) State {
	resp, ok := eventCtx.(commitVote)
	if !ok {
		fmt.Printf("invalid responce\n")
		// nothing to process, keeping state
		return s.keepResp
	}
	if !resp.commit {
		// this is abnormal situation, as during the Commit/Abort phase participant can only confirm the commit.
		// we will need to resend commit message to the participant.
		// go resendCommit()
		return s.keepResp
	}

	s.votesCnt--
	if s.votesCnt != 0 {
		return s.keepResp
	}
	return Init
}

func main() {
	sm := gfsm2.NewBuilder[State]().
		SetDefaultState(Init).
		SetSmContext(&coordinatorContext{partCnt: 3}).
		RegisterState(Init, &initState{}, []State{Wait}).
		RegisterState(Wait, &waitState{}, []State{Abort, Commit}).
		RegisterState(Abort, &responseState{
			keepResp: Abort,
		}, []State{Init}).
		RegisterState(Commit, &responseState{
			keepResp: Commit,
		}, []State{Init}).
		Build()

	sm.Start()
	defer sm.Stop()

	fmt.Printf("SM state (pre commit request): %s\n", sm.State())
	err := sm.ProcessEvent(commitRequest{"commit_1"})
	if err != nil {
		fmt.Printf("SM state: %s\n", sm.State())
		return
	}
	fmt.Printf("SM state (postcommit request): %s\n", sm.State())

	for i := 0; i < 3; i++ {
		err := sm.ProcessEvent(commitVote{commit: true})
		if err != nil {
			fmt.Printf("unable to vote: %v\n", err)
			return
		}
		fmt.Printf("SM state (voting): %s\n", sm.State())
	}
	fmt.Printf("SM state (pre confirm): %s\n", sm.State())
	for i := 0; i < 3; i++ {
		err := sm.ProcessEvent(commitVote{commit: true})
		if err != nil {
			fmt.Printf("unable complete: %v\n", err)
			return
		}
		fmt.Printf("SM state (confirming): %s\n", sm.State())
	}
	fmt.Printf("SM state (post confirm): %s\n", sm.State())
}
