package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/prometheus/promql/parser"
)

// This code helps to understand what the promql query parser is doing.
// Usage: go run dev/parse.go '1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes) > 0.8'
func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	expr, err := parser.ParseExpr(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	be, ok := expr.(*parser.BinaryExpr)
	if ok {
		fmt.Println("Query Left Hand Side:")
		fmt.Printf("  Type: %s\n", be.LHS.Type())
		fmt.Printf("  Expr: %s\n", be.LHS.String())
		fmt.Println("Query Right Hand Side:")
		fmt.Printf("  Type: %s\n", be.RHS.Type())
		fmt.Printf("  Expr: %s\n", be.RHS.String())
	} else {
		fmt.Println("Query is not a binary expression")
		return
	}

	fmt.Println("Variable:")
	fmt.Printf("%#v\n", expr)
}
