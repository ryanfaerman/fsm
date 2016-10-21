package fsm_test

import (
	"testing"
	"time"

	"github.com/nbio/st"
	"github.com/processout/fsm"
)

var (
	statePending  = fsm.NewState(fsm.String("pending"))
	stateStarted  = fsm.NewState(fsm.String("started"))
	stateFinished = fsm.NewState(fsm.String("finished"))
)

// Thing is a minimal struct that is an fsm.Stater
type Thing struct {
	State fsm.State
}

func (t *Thing) CurrentState() fsm.State { return t.State }
func (t *Thing) SetState(s fsm.State)    { t.State = s }

func TestRulesetTransitions(t *testing.T) {
	rules := fsm.CreateRuleset(
		fsm.T{statePending, stateStarted},
		fsm.T{stateStarted, stateFinished},
	)

	examples := []struct {
		subject fsm.Stater
		goal    fsm.State
		outcome bool
	}{
		// A Stater is responsible for setting its default state
		{&Thing{}, stateStarted, false},
		{&Thing{}, statePending, false},
		{&Thing{}, stateFinished, false},

		{&Thing{State: statePending}, stateStarted, true},
		{&Thing{State: statePending}, statePending, false},
		{&Thing{State: statePending}, stateFinished, false},

		{&Thing{State: stateStarted}, stateStarted, false},
		{&Thing{State: stateStarted}, statePending, false},
		{&Thing{State: stateStarted}, stateFinished, true},
	}

	for i, ex := range examples {
		st.Expect(t, rules.Permitted(ex.subject, ex.goal), ex.outcome, i)
	}
}

func TestRulesetParallelGuarding(t *testing.T) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{statePending, stateStarted})
	rules.AddTransition(fsm.T{stateStarted, stateFinished})

	// Add two failing rules, the slow should be caught first
	rules.AddRule(fsm.T{stateStarted, stateFinished}, func(subject fsm.Stater, goal fsm.State) bool {
		time.Sleep(1 * time.Second)
		t.Error("Slow rule should have been short-circuited")
		return false
	})

	rules.AddRule(fsm.T{stateStarted, stateFinished}, func(subject fsm.Stater, goal fsm.State) bool {
		return false
	})

	st.Expect(t, rules.Permitted(&Thing{State: stateStarted}, stateFinished), false)
}

func TestMachineTransition(t *testing.T) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{statePending, stateStarted})
	rules.AddTransition(fsm.T{stateStarted, stateFinished})

	some_thing := Thing{State: statePending}
	the_machine := fsm.New(fsm.WithRules(rules), fsm.WithSubject(&some_thing))

	var err error

	// should not be able to transition to the current state
	err = the_machine.Transition(statePending)
	st.Expect(t, err, fsm.ErrInvalidTransition)
	st.Expect(t, some_thing.State, fsm.State(statePending))

	// should not be able to skip states
	err = the_machine.Transition(stateFinished)
	st.Expect(t, err, fsm.ErrInvalidTransition)
	st.Expect(t, some_thing.State, fsm.State(statePending))

	// should be able to transition to the next valid state
	err = the_machine.Transition(stateStarted)
	st.Expect(t, err, nil)
	st.Expect(t, some_thing.State, fsm.State(stateStarted))
}

func BenchmarkRulesetParallelGuarding(b *testing.B) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{statePending, stateStarted})
	rules.AddTransition(fsm.T{stateStarted, stateFinished})

	// Add two failing rules, one very slow and the other terribly fast
	rules.AddRule(fsm.T{stateStarted, stateFinished}, func(subject fsm.Stater, goal fsm.State) bool {
		time.Sleep(1 * time.Second)
		return false
	})

	rules.AddRule(fsm.T{stateStarted, stateFinished}, func(subject fsm.Stater, goal fsm.State) bool {
		return false
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(&Thing{State: stateStarted}, stateFinished)
	}
}

func BenchmarkRulesetTransitionPermitted(b *testing.B) {
	// Permitted a transaction requires the transition to be valid and all of its
	// guards to pass. Since we have to run every guard and there won't be any
	// short-circuiting, this should actually be a little bit slower as a result,
	// depending on the number of guards that must pass.
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{statePending, stateStarted})
	rules.AddTransition(fsm.T{stateStarted, stateFinished})

	some_thing := &Thing{State: stateStarted}

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
	rules.AddTransition(fsm.T{statePending, stateStarted})
	rules.AddTransition(fsm.T{stateStarted, stateFinished})

	some_thing := &Thing{State: statePending}

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
	rules.AddTransition(fsm.T{statePending, stateStarted})

	rules.AddRule(fsm.T{stateStarted, stateFinished}, func(subject fsm.Stater, goal fsm.State) bool {
		return false
	})

	some_thing := &Thing{State: stateStarted}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(some_thing, stateFinished)
	}
}
