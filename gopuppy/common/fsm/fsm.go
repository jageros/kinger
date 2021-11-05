package fsm

import "kinger/gopuppy/common/glog"

type State interface {
	String() string
	OnEnter() error
	OnLeave()
}

type Event string

type FSM struct {
	state State
	stateTransfers map[string]map[Event]State
	stateChanging bool
}

func NewFSM(state State) *FSM {
	return &FSM{state: state}
}

func (f *FSM) GetState() State {
	return f.state
}

func (f *FSM) StateChanging() bool {
	return f.stateChanging
}

func (f *FSM) SetState(state State) {
	f.stateChanging = true
	f.state.OnLeave()
	oldState := f.state
	f.state = state
	if err := state.OnEnter(); err != nil {
		f.state = oldState
		glog.Errorf("FSM SetState error, oldState=%s newState=%s error=%s", f.state, state, err)
	}
	f.stateChanging = false
}

func (f *FSM) AddStateTransfer(state State, event Event, newState State) {
	if f.stateTransfers == nil {
		f.stateTransfers = make(map[string]map[Event]State)
	}

	eventToNewState, ok := f.stateTransfers[state.String()]
	if !ok {
		eventToNewState = make(map[Event]State)
		f.stateTransfers[state.String()] = eventToNewState
	}

	eventToNewState[event] = newState
}

func (f *FSM) EmitEvent(event Event) {
	if f.stateTransfers == nil {
		return
	}

	eventToNewState, ok := f.stateTransfers[f.state.String()]
	if !ok {
		return
	}

	if newState, ok := eventToNewState[event]; ok {
		f.SetState(newState)
	}
}
