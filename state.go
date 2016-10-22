package fsm

// ID is the type of what will be used to compare between states
type ID interface{}

// State describes a node of the machine, see NewState for more info,
// use ID() and I() to get information for this state
type State struct {
	id ID
	i  interface{}
}

// NewState creates a new state where a dataset which can be IDed
// is passed (implements IDer), where the id of the State
// determines the transitions (e.g. id:'pending'->id:'started'),
// and you can optionally include other data in the IDer which
// can be associated with this state, this helps
// if you want to customize transition rules.
func NewState(i IDer) State {
	return State{
		id: i.ID(),
		i:  i,
	}
}

// ID returns the id of the state
func (s State) ID() ID {
	return s.id
}

// I returns the interface associated with the state
func (s State) I() interface{} {
	return s.i
}

// IDer describes an interface that can return an ID for
// the transitions to take place, (e.g. id:'pending'->id:'started').
type IDer interface {
	// ID returns the id of this node/state, to be used for state identification
	ID() ID
}

// String implements the IDer interface, in case you only want to base yourself
// on an ID for transitions
type String string

// ID is for the IDer interface
func (s String) ID() ID {
	return s
}
