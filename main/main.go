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
func (f FlowState) ID() fsm.ID { return f.Name }

// checkEvolve will be a guard for checking if a transition can go through
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
	flow3 := []fsm.State{pendingt, finished}

	// Define our machine and its rules
	machine := fsm.Machine{}
	rules := fsm.Ruleset{}
	// Remember, for transitions only the ID() function matters (but you can do more in guards)
	rules.AddRule(fsm.NewTransition(pendingf, startedt), checkEvolve)
	rules.AddRule(fsm.NewTransition(startedt, finished), checkEvolve)
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

	// Test flow3
	machine.State = flow3[0]
	for _, s := range flow3[1:] {
		if err = machine.Transition(s); err != nil {
			fmt.Println("To be expected", err)
			break
		}
	}
	fmt.Println(machine) // pending
}
