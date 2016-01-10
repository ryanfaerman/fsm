FSM
===

[ ![Codeship Status for ryanfaerman/fsm](https://codeship.com/projects/7529e360-b173-0132-b520-32bd639983ea/status?branch=master)](https://codeship.com/projects/69855) [![GoDoc](https://godoc.org/github.com/ryanfaerman/fsm?status.png)](https://godoc.org/github.com/ryanfaerman/fsm)


FSM provides a lightweight finite state machine for Golang. It runs allows any number of transition checks you'd like the it runs them in parallel. It's tested and benchmarked too.

## Install

```
go get github.com/ryanfaerman/fsm
```

## Usage

```go
package main

import (
    "log"
    "fmt"
    "github.com/ryanfaerman/fsm"
)

type Thing struct {
    State fsm.State

    // our machine cache
    machine *fsm.Machine
}

// Add methods to comply with the fsm.Stater interface
func (t *Thing) CurrentState() fsm.State { return t.State }
func (t *Thing) SetState(s fsm.State)    { t.State = s }

// A helpful function that lets us apply arbitrary rulesets to this
// instances state machine without reallocating the machine. While not
// required, it's something I like to have.
func (t *Thing) Apply(r *fsm.Ruleset) *fsm.Machine {
    if t.machine == nil {
        t.machine = &fsm.Machine{Subject: t}
    }

    t.machine.Rules = r
    return t.machine
}

func main() {
    var err error

    some_thing := Thing{State: "pending"} // Our subject
    fmt.Println(some_thing)

    // Establish some rules for our FSM
    rules := fsm.Ruleset{}
    rules.AddTransition(fsm.T{"pending", "started"})
    rules.AddTransition(fsm.T{"started", "finished"})

    err = some_thing.Apply(&rules).Transition("started")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(some_thing)
}

```

*Note:* FSM makes no effort to determine the default state for any ruleset. That's your job.

The `Apply(r *fsm.Ruleset) *fsm.Machine` method is absolutely optional. I like having it though. It solves a pretty common problem I usually have when working with permissions - some users aren't allowed to transition between certain states.

Since the rules are applied to the the subject (through the machine) I can have a simple lookup to determine the ruleset that the subject has to follow for a given user. As a result, I rarely need to use any complicated guards but I can if need be. I leave the lookup and the maintaining of independent rulesets as an exercise of the user.

## Benchmarks
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


