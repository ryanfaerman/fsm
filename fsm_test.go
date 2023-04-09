package fsm_test

import (
	"errors"
	"testing"
	"time"

	"github.com/nbio/st"
	"github.com/nozim/fsm"
)

// Thing is a minimal struct that is an fsm.Stater
type Thing struct {
	State fsm.State
}

func (t *Thing) CurrentState() fsm.State { return t.State }
func (t *Thing) SetState(s fsm.State)    { t.State = s }

func TestSomething(t *testing.T) {
st.Expect(t,2+2,4)	
}

func TestAnother(t *testing.T) {
   st.Expect(t, 9,3*3)
}

func TestAndAnotherAnother(t *testing.T) {
   st.Expect(t, 9,3*3)
}

func TestRulesetTransitions(t *testing.T) {
	rules := fsm.CreateRuleset(
		fsm.T{"pending", "started"},
		fsm.T{"started", "finished"},
	)

	examples := []struct {
		subject fsm.Stater
		goal    fsm.State
		outcome error
	}{
		// A Stater is responsible for setting its default state
		{&Thing{}, "started", fsm.InvalidTransition},
		{&Thing{}, "pending", fsm.InvalidTransition},
		{&Thing{}, "finished", fsm.InvalidTransition},

		{&Thing{State: "pending"}, "started", nil},
		{&Thing{State: "pending"}, "pending", fsm.InvalidTransition},
		{&Thing{State: "pending"}, "finished", fsm.InvalidTransition},

		{&Thing{State: "started"}, "started", fsm.InvalidTransition},
		{&Thing{State: "started"}, "pending", fsm.InvalidTransition},
		{&Thing{State: "started"}, "finished", nil},
	}

	for i, ex := range examples {
		st.Expect(t, rules.Permitted(ex.subject, ex.goal), ex.outcome, i)
	}
}

func TestRulesetParallelGuarding(t *testing.T) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	// Add two failing rules, the slow should be caught first
	rules.AddRule(fsm.T{"started", "finished"}, func(subject fsm.Stater, goal fsm.State) error {
		time.Sleep(1 * time.Second)
		t.Error("Slow rule should have been short-circuited")
		return errors.New("Slow guard")
	})

	rules.AddRule(fsm.T{"started", "finished"}, func(subject fsm.Stater, goal fsm.State) error {
		return errors.New("Always reject guard")
	})

	st.Expect(t, rules.Permitted(&Thing{State: "started"}, "finished"), errors.New("Always reject guard"))
}

func TestMachineTransition(t *testing.T) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	some_thing := Thing{State: "pending"}
	the_machine := fsm.New(fsm.WithRules(rules), fsm.WithSubject(&some_thing))

	var err error

	// should not be able to transition to the current state
	err = the_machine.Transition("pending")
	st.Expect(t, err, fsm.InvalidTransition)
	st.Expect(t, some_thing.State, fsm.State("pending"))

	// should not be able to skip states
	err = the_machine.Transition("finished")
	st.Expect(t, err, fsm.InvalidTransition)
	st.Expect(t, some_thing.State, fsm.State("pending"))

	// should be able to transition to the next valid state
	err = the_machine.Transition("started")
	st.Expect(t, err, nil)
	st.Expect(t, some_thing.State, fsm.State("started"))
}

func BenchmarkRulesetParallelGuarding(b *testing.B) {
	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	// Add two failing rules, one very slow and the other terribly fast
	rules.AddRule(fsm.T{"started", "finished"}, func(subject fsm.Stater, goal fsm.State) error {
		time.Sleep(1 * time.Second)
		return errors.New("Slow guard")
	})

	rules.AddRule(fsm.T{"started", "finished"}, func(subject fsm.Stater, goal fsm.State) error {
		return errors.New("Failing guard")
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(&Thing{State: "started"}, "finished")
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

	some_thing := &Thing{State: "started"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(some_thing, "finished")
	}

}

func BenchmarkRulesetTransitionInvalid(b *testing.B) {
	// This should be incredibly fast, since fsm.T{"pending", "finished"}
	// doesn't exist in the Ruleset. We expect some small overhead from creating
	// the transition to check the internal map, but otherwise, we should be
	// bumping up against the speed of a map lookup itself.

	rules := fsm.Ruleset{}
	rules.AddTransition(fsm.T{"pending", "started"})
	rules.AddTransition(fsm.T{"started", "finished"})

	some_thing := &Thing{State: "pending"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(some_thing, "finished")
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

	rules.AddRule(fsm.T{"started", "finished"}, func(subject fsm.Stater, goal fsm.State) error {
		return errors.New("Failing guard")
	})

	some_thing := &Thing{State: "started"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rules.Permitted(some_thing, "finished")
	}
}
