package traversal

// TODO template in Go2.0 is expected
// Depth First Search
func TraversalDFS(val interface{}, processFn func(ele interface{}, depth int) (gotoNextLayer bool)) {
	// Nodes already visited at an earlier level.
	visited := map[interface{}]bool{}
	traversalDFS(val, 0, visited, processFn)
}

func traversalDFS(ele interface{}, depth int, visited map[interface{}]bool, processFn func(ele interface{}, depth int) (gotoNextLayer bool)) (gotoNextLayer bool) {
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

	// Scan node for nodes to include.
	for _, e := range node.Lefts() {
		traversalDFS(e, depth+1, visited, processFn)
	}
	for _, e := range node.Middles() {
		traversalDFS(e, depth+1, visited, processFn)
	}
	for _, e := range node.Rights() {
		traversalDFS(e, depth+1, visited, processFn)
	}
	return true
}
