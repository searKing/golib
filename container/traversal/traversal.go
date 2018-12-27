package traversal

// A field represents a single Node found in a structure.
type Node struct {
	ele   interface{}
	depth int
}

func (n *Node) Lefts() []interface{} {
	if n.ele == nil {
		return nil
	}
	lefters, ok := n.ele.(Leftser)
	if ok {
		return lefters.Lefts()
	}
	lefter, ok := n.ele.(Lefter)
	if ok {
		return []interface{}{lefter.Left()}
	}
	return nil
}

func (n *Node) Middles() []interface{} {
	if n.ele == nil {
		return nil
	}
	middleers, ok := n.ele.(Middleser)
	if ok {
		return middleers.Middles()
	}
	middleer, ok := n.ele.(Middleer)
	if ok {
		return []interface{}{middleer.Middle()}
	}
	return nil

}
func (n *Node) Rights() []interface{} {
	if n.ele == nil {
		return nil
	}
	righters, ok := n.ele.(Rightser)
	if ok {
		return righters.Rights()
	}
	righter, ok := n.ele.(Righter)
	if ok {
		return []interface{}{righter.Right()}
	}
	return nil

}

type Lefter interface {
	// Left returns the left list element or nil.
	Left() interface{}
}
type Middleer interface {
	// Middle returns the middle list element or nil.
	Middle() interface{}
}
type Righter interface {
	// Right returns the middle list element or nil.
	Right() interface{}
}

type Leftser interface {
	// Left returns the left list element or nil.
	Lefts() []interface{}
}
type Middleser interface {
	// Middle returns the middle list element or nil.
	Middles() []interface{}
}
type Rightser interface {
	// Right returns the middle list element or nil.
	Rights() []interface{}
}
