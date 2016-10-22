package main

import (
	"errors"
	"fmt"

	"github.com/processout/fsm"
)

// FlowState represents a state within a flow that should follow a
// finite state machine, according to your rules. Has to implement IDer
type FlowState struct {
	Name      string
	CanEvolve bool
}

// ID will return the name as to ID the flow for the transitions
func (f FlowState) ID() string { return f.Name }

// Machine describes the flow in one go, one could also use "fsm.T",
// but there is no caching
type Machine struct {
	State fsm.State

	// our machine cache
	machine *fsm.Machine
}

func checkEvolve(start fsm.State, goal fsm.State) error {
	if start.I().(FlowState).CanEvolve {
		return nil
	}
	return errors.New("Can't evolve")
}

// Say you have two flows, flow1 and flow2, you want to see
// if you can add or if the flow respects the fsm.
func main() {
	var err error

	// Say the flow can only transition if CanEvolve==true
	pendingt := fsm.NewState(FlowState{Name: "pending", CanEvolve: true})
	pendingf := fsm.NewState(FlowState{Name: "pending", CanEvolve: false})
	startedt := fsm.NewState(FlowState{Name: "started", CanEvolve: true})
	finished := fsm.NewState(FlowState{Name: "finished", CanEvolve: false})

	flow1 := []fsm.State{pendingt, startedt, finished}
	flow2 := []fsm.State{pendingf, startedt, finished}

	// Define our machine and its rules
	machine := fsm.Machine{}
	rules := fsm.Ruleset{}
	rules.AddRule(fsm.T{"pending", "started"}, checkEvolve)
	rules.AddRule(fsm.T{"started", "finished"}, checkEvolve)
	machine.Rules = &rules

	// Test flow1
	machine.State = flow1[0]
	for _, s := range flow1[1:] {
		if err = machine.Transition(s); err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Println(machine) // finished

	// Test flow2
	machine.State = flow2[0]
	for _, s := range flow2[1:] {
		if err = machine.Transition(s); err != nil {
			fmt.Println("To be expected:", err)
			break
		}
	}
	fmt.Println(machine) // pending
}
