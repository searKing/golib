package traversal

// TODO template in Go2.0 is expected
// Breadth First Search
func TraversalBFS(val interface{}, processFn func(ele interface{}, depth int) (gotoNextLayer bool)) {
	// Nodes already visited at an earlier level.
	visited := map[interface{}]bool{}
	traversalBFS(val, 0, visited, processFn)
}

func traversalBFS(ele interface{}, depth int, visited map[interface{}]bool, processFn func(ele interface{}, depth int) (gotoNextLayer bool)) (gotoNextLayer bool) {
	if visited[ele] {
		return true
	}
	visited[ele] = true
	node := Node{
		ele:   ele,
		depth: 0}
	if !processFn(node.ele, depth) {
		return false
	}
	// Anonymous fields to explore at the current level and the next.
	next := []interface{}{}
	// Scan node for nodes to include.
	for _, e := range node.Lefts() {
		if !processFn(e, depth) {
			continue
		}
		next = append(next, e)
	}
	for _, e := range node.Middles() {
		if !processFn(e, depth) {
			continue
		}
		next = append(next, e)
	}
	for _, e := range node.Rights() {
		if !processFn(e, depth) {
			continue
		}
		next = append(next, e)
	}
	for _, e := range next {
		traversalBFS(e, depth, visited, processFn)
	}
	return true
}
