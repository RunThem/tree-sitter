package main

import (
	"fmt"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"log"
	"os"
)

var (
	input       []byte
	staticFuncs []string
	globalFuncs []string

	err error
)

func main() {

	if len(os.Args) != 2 {
		log.Println("usage: cmd {file}")
		os.Exit(-1)
	}

	input, err = os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(c.GetLanguage())
	queryDSL, err := sitter.NewQuery([]byte("(function_definition) @func"), c.GetLanguage())
	if err != nil {
		log.Fatalln(err)
	}

	tree := parser.Parse(nil, input)
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(queryDSL, tree.RootNode())

	for {
		m, ok := queryCursor.NextMatch()
		if !ok || len(m.Captures) == 0 {
			break
		}
		node := m.Captures[0].Node

		if isMainFunc(node) {
			break
		}

		if isStaticFunc(node) {
			staticFuncs = append(staticFuncs, functionDeclartor(node))
		} else {
			globalFuncs = append(globalFuncs, functionDeclartor(node))
		}
	}

	fmt.Println(funcNameCat("/* static function */", staticFuncs))
	fmt.Println(funcNameCat("/* global function */", globalFuncs))
}

func functionDeclartor(node *sitter.Node) string {
	startIdx := node.StartByte()
	endIdx := node.Child(int(node.ChildCount() - 2)).EndByte()

	if endIdx <= startIdx {
		return ""
	}

	return string(input[startIdx:endIdx])
}

func isStaticFunc(node *sitter.Node) bool {
	for i := uint32(0); i < node.ChildCount()-1; i++ {
		n := node.Child(int(i))
		if n.Type() == "storage_class_specifier" && n.Content(input) == "static" {
			return true
		}
	}

	return false
}

func isMainFunc(node *sitter.Node) bool {
	queryCursor := sitter.NewQueryCursor()
	queryDSL, err := sitter.NewQuery([]byte("(identifier) @id"), c.GetLanguage())
	if err != nil {
		return false
	}

	queryCursor.Exec(queryDSL, node)

	idNode, ok := queryCursor.NextMatch()
	if !ok {
		return false
	}

	n := idNode.Captures[0].Node
	return n.Content(input) == "main"
}

func funcNameCat(comment string, code []string) string {
	codes := comment + "\n"

	for _, v := range code {
		codes += fmt.Sprintf("%s;\n", v)
	}

	return codes
}
