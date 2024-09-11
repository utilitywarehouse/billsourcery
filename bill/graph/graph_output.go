package graph

import (
	"fmt"
	"slices"
)

type graphOutput interface {
	Start() error
	End() error
	AddNode(id string, name string, tags []string) error
	AddCall(from_id string, to_id string) error
}

type DotGraphOutput struct{}

func (o *DotGraphOutput) Start() error {
	fmt.Println("digraph calls {")
	return nil
}

func (o *DotGraphOutput) End() error {
	fmt.Println("}")
	return nil
}

func (o *DotGraphOutput) AddNode(id string, name string, tags []string) error {
	colour := ""

	if slices.Contains(tags, "form") {
		colour = "lightgreen"
	} else if slices.Contains(tags, "report") {
		colour = "orange"
	} else if slices.Contains(tags, "public_procedure") {
		colour = "yellow"
	} else if slices.Contains(tags, "method") {
		if slices.Contains(tags, "missing") {
			colour = "red"
		} else {
			colour = "lightblue"
		}
	}

	fmt.Printf("\t%s [label=\"%s\" style=\"filled\" fillcolor=\"%s\"]\n", id, name, colour)

	return nil
}

func (o *DotGraphOutput) AddCall(from string, to string) error {
	fmt.Printf("\t%s -> %s\n", from, to)
	return nil
}

type NeoGraphOutput struct{}

func (o *NeoGraphOutput) Start() error {
	return nil
}

func (o *NeoGraphOutput) End() error {
	return nil
}

func (o NeoGraphOutput) AddNode(id string, name string, tags []string) error {

	fmt.Printf("MERGE (%s:Node {id:\"%s\", name:\"%s\"})\n", id, id, name)

	if slices.Contains(tags, "form") {
		fmt.Printf("SET %s :Form;\n", id)
	} else if slices.Contains(tags, "report") {
		fmt.Printf("SET %s :Report;\n", id)
	} else if slices.Contains(tags, "export") {
		fmt.Printf("SET %s :Export;\n", id)
	} else if slices.Contains(tags, "import") {
		fmt.Printf("SET %s :Import;\n", id)
	} else if slices.Contains(tags, "public_procedure") {
		fmt.Printf("SET %s :PublicProcedure;\n", id)
	} else if slices.Contains(tags, "method") {
		if slices.Contains(tags, "missing") {
			fmt.Printf("SET %s :Method\nSET %s :Missing;\n", id, id)
		} else {
			fmt.Printf("SET %s :Method;\n", id)
		}
	}

	return nil
}

func (o *NeoGraphOutput) AddCall(from string, to string) error {
	fmt.Printf("MERGE (f:Node {id: \"%s\"}) MERGE (t:Node {id: \"%s\"}) MERGE (f)-[:calls]->(t);\n", from, to)
	return nil
}
