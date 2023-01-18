# mwpm
minimum weight perfect matching

# Referece
Kolmogorov, V. Blossom V: a new implementation of a minimum cost perfect matching algorithm. <em>Math. Prog. Comp.</em> 1, 43–67 (2009). https://doi.org/10.1007/s12532-009-0002-8

# Usage
```go
package main

import (
    "github.com/hyosangkang/mwpm"
    "gonum.org/v1/gonum/graph/simple"
)

func main() {
    wg := simple.NewWeightedUndirected(0.0, 0.0)
    m, _ := mwpm.MinimumWeightPerfectMatching(wg)
    fmt.Println(m)
}
```
