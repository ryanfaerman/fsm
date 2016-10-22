FSM
===

[![GoDoc](https://godoc.org/github.com/processout/fsm?status.png)](https://godoc.org/github.com/processout/fsm)

FSM is a fork from [github.com/ryanfaerman/fsm](https://github.com/ryanfaerman/fsm), and is used internally at ProcessOut for quite a few things.

### From the author:
> FSM provides a lightweight finite state machine for Golang. It runs allows any number of transition checks you'd like the it runs them in parallel. It's tested and benchmarked too.

### Why we forked:
We really liked ryanfaerman's go package for a finite state machine, 
but we needed it to be more capable in terms of `guards`,
 as well as being more transparent as to what is happening
(switch from `bool` to `error`).

Basically, you can define more complex `rules` to guarantee that the flow you are
using is sane according to the finite state machine. Designed to both test a machine flow,
and stay on a defined machine flow.


## Install

```
go get github.com/processout/fsm
```


## Usage

```go
// package main is similar to github.com/ryanfaerman/fsm's package, but is adjusted
// for our modifications
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
```

*Note:* FSM makes no effort to determine the default state for any ruleset. That's your job.
You have to set `machine.State` at the start of your flow.

## Benchmarks (from ryanfaerman)
Golang makes it easy enough to benchmark things... why not do a few general benchmarks?

```shell
$ go test -bench=.
PASS
BenchmarkRulesetParallelGuarding      100000   13163 ns/op
BenchmarkRulesetTransitionPermitted  1000000    1805 ns/op
BenchmarkRulesetTransitionDenied    10000000     284 ns/op
BenchmarkRulesetRuleForbids          1000000    1717 ns/op
ok    github.com/ryanfaerman/fsm 14.864s
```

I think that's pretty good. So why do I go through the trouble of running the guards in parallel? Consider the benchmarks below:

```shell
$ go test -bench=.
PASS
BenchmarkRulesetParallelGuarding    100000       13261 ns/op
ok  github.com/ryanfaerman/fsm 1.525s


$ go test -bench=.
PASS
BenchmarkRulesetSerialGuarding         1  1003140956 ns/op
ok  github.com/ryanfaerman/fsm 1.007s
```

For the Parallel vs Serial benchmarks I had a guard that slept for 1 second. While I don't imagine most guards will take that long, the point remains true. Some guards will be comparatively slow -- they'll be accessing the database or consulting with some other outside service -- and why not get an answer back as soon as possible?


