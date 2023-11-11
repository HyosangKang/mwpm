package mwpm

import (
	"fmt"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestMain(m *testing.M) {
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
	for i := 0; i < 10000; i++ {
		fmt.Println()
		fmt.Println("----------------NEW----------------")
		pair, _ := Run(wg)
		fmt.Println(pair)
	}
}
