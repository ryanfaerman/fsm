package fsm_test

import (
	"errors"
	"testing"
	"time"

	"github.com/nbio/st"
	"github.com/processout/fsm"
)

var (
	stateNone     = fsm.NewState(fsm.String(""))
	statePending  = fsm.NewState(fsm.String("pending"))
	stateStarted  = fsm.NewState(fsm.String("started"))
	stateFinished = fsm.NewState(fsm.String("finished"))
	testError     = errors.New("test error")
)

func TestRulesetTransitions(t *testing.T) {
	rules := fsm.CreateRuleset(
		fsm.T{"pending", "started"},
		fsm.T{"started", "finished"},
	)

	examples := []struct {
		start   fsm.State
		goal    fsm.State
		outcome bool
	}{
		// A Stater is responsible for setting its default state
		{stateNone, stateStarted, false},
		{stateNone, statePending, false},
		{stateNone, stateFinished, false},

		{statePending, stateStarted, true},
		{statePending, statePending, false},
		{statePending, stateFinished, false},

		{stateStarted, stateStarted, false},
		{stateStarted, statePending, false},
		{stateStarted, stateFinished, true},
	}

	for i, ex := range examples {
		out := rules.Permitted(ex.start, ex.goal) == nil
		st.Expect(t, out, ex.outcome, i)
	}
}

func TestRulesetParallelGuarding(t *testing.T) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	// Add two failing rules, the slow should be caught first
	rules.AddRule(fsm.T{"started", "finished"}, func(start fsm.State, goal fsm.State) error {
		time.Sleep(1 * time.Second)
		t.Error("Slow rule should have been short-circuited")
		return testError
	})

	rules.AddRule(fsm.T{"started", "finished"}, func(start fsm.State, goal fsm.State) error {
		return testError
	})

	st.Expect(t, rules.Permitted(stateStarted, stateFinished).Error(),
		"Guard failed from started to finished: "+testError.Error())
}

func TestMachineTransition(t *testing.T) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	the_machine := fsm.Machine{
		State: statePending,
		Rules: &rules,
	}

	var err error

	// should not be able to transition to the current state
	err = the_machine.Transition(statePending)
	st.Expect(t, err, errors.New("No rules found for pending to pending"))
	st.Expect(t, the_machine.State, statePending)

	// should not be able to skip states
	err = the_machine.Transition(stateFinished)
	st.Expect(t, err, errors.New("No rules found for pending to finished"))
	st.Expect(t, the_machine.State, statePending)

	// should be able to transition to the next valid state
	err = the_machine.Transition(stateStarted)
	st.Expect(t, err, nil)
	st.Expect(t, the_machine.State, stateStarted)
}

func BenchmarkRulesetParallelGuarding(b *testing.B) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	// Add two failing rules, one very slow and the other terribly fast
	rules.AddRule(fsm.T{"started", "finished"}, func(start fsm.State, goal fsm.State) error {
		time.Sleep(1 * time.Second)
		return testError
	})

	rules.AddRule(fsm.T{"started", "finished"}, func(start fsm.State, goal fsm.State) error {
		return testError
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(stateStarted, stateFinished)
	}
}

func BenchmarkRulesetTransitionPermitted(b *testing.B) {
	// Permitted a transaction requires the transition to be valid and all of its
	// guards to pass. Since we have to run every guard and there won't be any
	// short-circuiting, this should actually be a little bit slower as a result,
	// depending on the number of guards that must pass.
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	some_thing := stateStarted

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(some_thing, stateFinished)
	}

}

func BenchmarkRulesetTransitionInvalid(b *testing.B) {
	// This should be incredibly fast, since fsm.T{statePending, stateFinished}
	// doesn't exist in the Ruleset. We expect some small overhead from creating
	// the transition to check the internal map, but otherwise, we should be
	// bumping up against the speed of a map lookup itself.

	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	some_thing := statePending

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(some_thing, stateFinished)
	}
}

func BenchmarkRulesetRuleForbids(b *testing.B) {
	// Here, we explicity create a transition that is forbidden. This simulates an
	// otherwise valid transition that would be denied based on a user role or the like.
	// It should be slower than a standard invalid transition, since we have to
	// actually execute a function to perform the check. The first guard to
	// fail (returning false) will short circuit the execution, getting some some speed.

	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})

	rules.AddRule(fsm.T{"started", "finished"}, func(start fsm.State, goal fsm.State) error {
		return testError
	})

	some_thing := stateStarted

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(some_thing, stateFinished)
	}
}
