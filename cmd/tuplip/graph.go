package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/emicklei/dot"
)

// graphCmd contains the options for the graph command.
type graphCmd struct {
	Command []string `arg:"" optional:"" help:"the commands for which to show the command graph" sep:" "`
}

// Run prints the help graph.
func (h *graphCmd) Run(realCtx *kong.Context) error {
	ctx, err := kong.Trace(realCtx.Kong, h.Command)
	if err != nil {
		return err
	}
	if ctx.Error != nil {
		return ctx.Error
	}
	graph := dot.NewGraph(dot.Directed)
	var node = ctx.Selected()
	if node == nil {
		node = ctx.Model.Node
	}
	commandGraph(node, nil, graph)
	_, err = fmt.Println(graph.String())
	return err
}

// commandGraph adds all sub nodes to the given graph.
func commandGraph(nextNode *kong.Node, parentNode *dot.Node, graph *dot.Graph) {
	var newNode dot.Node
	if nextNode == nil {
		return
	}
	switch nextNode.Type {
	case kong.CommandNode:
		newNode = graph.Node(nextNode.Name)
		newNode.Attr("shape", "cds")
	case kong.ArgumentNode:
		newNode = graph.Node("<" + nextNode.Name + ">").Box()
	default:
		newNode = graph.Node(nextNode.Name).Box()
	}
	dotNode := &newNode
	if parentNode != nil && len(parentNode.EdgesTo(newNode)) == 0 {
		parentNode.Edge(newNode)
	}
	for _, arg := range nextNode.Positional {
		argNode := graph.Node(arg.Summary())
		if len(dotNode.EdgesTo(argNode)) == 0 {
			dotNode.Edge(argNode)
		}
	}
	for _, subCmd := range nextNode.Children {
		commandGraph(subCmd, dotNode, graph)
	}
	return
}
