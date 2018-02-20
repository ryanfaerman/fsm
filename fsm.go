package fsm

import (
	"errors"
	"fmt"
)

type State string

// Guard provides protection against transitioning to the goal State.
// Returning true/false indicates if the transition is permitted or not.
type Guard func(subject Stater, goal State) error

var (
	InvalidTransition = errors.New("invalid transition")
)

// Transition is the change between States
type Transition interface {
	Origin() State
	Exit() State
}

// T implements the Transition interface; it provides a default
// implementation of a Transition.
type T struct {
	O, E State
}

func (t T) Origin() State { return t.O }
func (t T) Exit() State   { return t.E }

// Ruleset stores the rules for the state machine.
type Ruleset map[Transition][]Guard

// AddRule adds Guards for the given Transition
func (r Ruleset) AddRule(t Transition, guards ...Guard) {
	for _, guard := range guards {
		r[t] = append(r[t], guard)
	}
}

// AddTransition adds a transition with a default rule
func (r Ruleset) AddTransition(t Transition) {
	r.AddRule(t, func(subject Stater, goal State) error {
		if subject.CurrentState() != t.Origin() {
			return fmt.Errorf("invalid transition from %s to %s", subject.CurrentState(), t.Origin())
		}
		return nil
	})
}

// CreateRuleset will establish a ruleset with the provided transitions.
// This eases initialization when storing within another structure.
func CreateRuleset(transitions ...Transition) Ruleset {
	r := Ruleset{}

	for _, t := range transitions {
		r.AddTransition(t)
	}

	return r
}

// IsValidTransition determines if a transition is allowed.
// This occurs in parallel.
// NOTE: Guards are not halted if they are short-circuited for some
// transition. They may continue running *after* the outcome is determined.
// Returns an error if the transition is invalid or nil otherwise
func (r Ruleset) IsValidTransition(subject Stater, goal State) []error {
	attempt := T{subject.CurrentState(), goal}

	if guards, ok := r[attempt]; ok {
		outcome := make(chan error)

		for _, guard := range guards {
			go func(g Guard) {
				outcome <- g(subject, goal)
			}(guard)
		}

		guardErrors := make([]error, 0)
		for range guards {
			select {
			case err := <-outcome:
				if err != nil {
					guardErrors = append(guardErrors, err)
				}
			}
		}

		if len(guardErrors) > 0 {
			return guardErrors
		}

		return nil // All guards passed
	}
	return []error{fmt.Errorf("no transition from %s to %s", subject.CurrentState(), goal)}
}

// Stater can be passed into the FSM. The Stater is reponsible for setting
// its own default state. Behavior of a Stater without a State is undefined.
type Stater interface {
	CurrentState() State
	SetState(State)
}

// Machine is a pairing of Rules and a Subject.
// The subject or rules may be changed at any time within
// the machine's lifecycle.
type Machine struct {
	Rules   *Ruleset
	Subject Stater
}

// Transition attempts to move the Subject to the Goal state.
func (m Machine) Transition(goal State) []error {
	transitionErrors := m.Rules.IsValidTransition(m.Subject, goal)
	if len(transitionErrors) == 0 {
		m.Subject.SetState(goal)
		return nil
	}

	return transitionErrors
}

// New initializes a machine
func New(opts ...func(*Machine)) Machine {
	var m Machine

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// WithSubject is intended to be passed to New to set the Subject
func WithSubject(s Stater) func(*Machine) {
	return func(m *Machine) {
		m.Subject = s
	}
}

// WithRules is intended to be passed to New to set the Rules
func WithRules(r Ruleset) func(*Machine) {
	return func(m *Machine) {
		m.Rules = &r
	}
}
