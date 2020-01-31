package fsm

import (
	"errors"
	"fmt"
)

// Guard provides protection against transitioning to the goal State.
// Returning an error if the transition is not permitted
type Guard func(start *State, goal *State) error

const (
	errTransitionFormat  = "Cannot transition from %s to %s"
	errNoRulesFormat     = "No rules found for %s to %s"
	errGuardFailedFormat = "Guard failed from %s to %s: %s"
)

var (
	// ErrInvalidTransition describes the errors when doing an invalid transition
	ErrInvalidTransition = errors.New("invalid transition")
)

// Transition is the change between States
type Transition interface {
	Origin() ID
	Exit() ID
}

// T implements the Transition interface; it provides a default
// implementation of a Transition. It should only be instantiated
// with NewTransition.
type T struct {
	O, E ID
}

// Origin returns the starting state
func (t T) Origin() ID { return t.O }

// Exit returns the ending state
func (t T) Exit() ID { return t.E }

// NewTransition let's you create a new transition and apply some rules
func NewTransition(i1 IDer, i2 IDer) T {
	return T{
		O: i1.ID(),
		E: i2.ID(),
	}
}

// Ruleset stores the rules for the state machine.
type Ruleset map[ID][]Guard

// AddRule adds Guards for the given Transition
func (r Ruleset) AddRule(t Transition, guards ...Guard) {
	for _, guard := range guards {
		r[t] = append(r[t], guard)
	}
}

// AddTransition adds a transition with a default rule
func (r Ruleset) AddTransition(t Transition) {
	r.AddRule(t, func(start *State, goal *State) error {
		if start.ID() != t.Origin() {
			return fmt.Errorf(errTransitionFormat, start.ID(), goal.ID())
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

// Permitted determines if a transition is allowed.
// This occurs in parallel.
// NOTE: Guards are not halted if they are short-circuited for some
// transition. They may continue running *after* the outcome is determined.
func (r Ruleset) Permitted(start *State, goal *State) error {
	attempt := T{start.ID(), goal.ID()}

	if guards, ok := r[attempt]; ok {

		for _, guard := range guards {
			err := guard(start, goal)
			if err != nil {
				return fmt.Errorf(errGuardFailedFormat, start.ID(), goal.ID(), err.Error())
			}

			start.id = start.ID()
			goal.id = goal.ID()
		}
		return nil
	}
	return fmt.Errorf(errNoRulesFormat, start.ID(), goal.ID())
}

// Machine is a pairing of Rules and a State.
// The state or rules may be changed at any time within
// the machine's lifecycle.
type Machine struct {
	Rules *Ruleset
	State State
}

// Transition attempts to move the Subject to the Goal state.
func (m *Machine) Transition(goal State) (err error) {
	if err = m.Rules.Permitted(&m.State, &goal); err == nil {
		m.State = goal
		return nil
	}

	return err
}

// New initializes a machine
func New(opts ...func(*Machine)) Machine {
	var m Machine

	for _, opt := range opts {
		opt(&m)
	}

	return m
}
