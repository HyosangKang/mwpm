# mwpm
GO implementation of minimum weight perfect matching algorithm

# Reference
Kolmogorov, V. Blossom V: a new implementation of a minimum cost perfect matching algorithm. <em>Math. Prog. Comp.</em> 1, 43–67 (2009). https://doi.org/10.1007/s12532-009-0002-8
 
# Usage
The function `Run` is the main function that finds the perfect matching. It takes a graph that implements the interface `WeightedUndirectedGraph` and returns a list of pairs of IDs of nodes that are matched. (Each pair of IDs is sorted in ascending order.) The second output argument is the boolean value that indicates whether the matching is successful or not. Currently, it is `false` only when there are odd number of nodes in the graph.

```go
package main

import (
    "github.com/hyosangkang/mwpm"
    "gonum.org/v1/gonum/graph/simple"
)

func main() {
	adj := [][]int{
		{0, 1, 1, 0, 0, 0},
		{1, 0, 1, 1, 0, 0},
		{1, 1, 0, 1, 1, 0},
		{0, 1, 1, 0, 1, 1},
		{0, 0, 1, 1, 0, 1},
		{0, 0, 0, 1, 1, 0},
	}
	wg := simple.NewWeightedUndirectedGraph(0, 0)
	for i := 0; i < len(adj); i++ {
		wg.AddNode(wg.NewNode())
	}
	for i := 0; i < len(adj); i++ {
		for j := 0; j < len(adj[i]); j++ {
			if adj[i][j] != 0 {
				wg.SetWeightedEdge(wg.NewWeightedEdge(wg.Node(int64(i)), wg.Node(int64(j)), float64(adj[i][j])))
			}
		}
	}
    m, _ := mwpm.Run(wg)
    fmt.Println(m)
}
```

Output
```
[[0 2] [1 3] [4 5]]
```
