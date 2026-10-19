package commands

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const indexFileName = "index.json"

type FunctionEntry struct {
	Name     string `json:"name"`
	Receiver string `json:"receiver,omitempty"`
	Line     int    `json:"line"`
}

type IndexNode struct {
	Name      string          `json:"name"`
	Path      string          `json:"path"`
	Type      string          `json:"type"`
	Functions []FunctionEntry `json:"functions,omitempty"`
	Children  []*IndexNode    `json:"children,omitempty"`
}

func HandleIndexCommand() error {
	root, err := buildIndexNode(".")
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(indexFileName, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Written: %s\n", indexFileName)
	return nil
}

func buildIndexNode(dir string) (*IndexNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	node := &IndexNode{
		Name: filepath.Base(dir),
		Path: filepath.ToSlash(dir),
		Type: "dir",
	}
	for _, entry := range entries {
		if shouldSkipIndexEntry(entry.Name()) {
			continue
		}
		childPath := filepath.Join(dir, entry.Name())
		child, err := buildChildIndexNode(childPath, entry)
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}
	return node, nil
}

func shouldSkipIndexEntry(name string) bool {
	return strings.HasPrefix(name, ".") ||
		name == "vendor" ||
		name == "node_modules" ||
		name == indexFileName
}

func buildChildIndexNode(path string, entry os.DirEntry) (*IndexNode, error) {
	if entry.IsDir() {
		return buildNonEmptyDirNode(path)
	}
	if strings.HasSuffix(entry.Name(), ".go") {
		return buildGoFileNode(path, entry.Name())
	}
	return nil, nil
}

func buildNonEmptyDirNode(path string) (*IndexNode, error) {
	child, err := buildIndexNode(path)
	if err != nil {
		return nil, err
	}
	if len(child.Children) == 0 {
		return nil, nil
	}
	return child, nil
}

func buildGoFileNode(path string, name string) (*IndexNode, error) {
	functions, err := extractFunctions(path)
	if err != nil {
		return nil, err
	}
	return &IndexNode{
		Name:      name,
		Path:      filepath.ToSlash(path),
		Type:      "file",
		Functions: functions,
	}, nil
}

func extractFunctions(path string) ([]FunctionEntry, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	var functions []FunctionEntry
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		functions = append(functions, FunctionEntry{
			Name:     funcDecl.Name.Name,
			Receiver: receiverOfFunc(funcDecl),
			Line:     fileSet.Position(funcDecl.Pos()).Line,
		})
	}
	return functions, nil
}

func receiverOfFunc(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return ""
	}
	return receiverTypeName(funcDecl.Recv.List[0].Type)
}

func receiverTypeName(expr ast.Expr) string {
	switch receiverType := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(receiverType.X)
	case *ast.Ident:
		return receiverType.Name
	case *ast.IndexExpr:
		return receiverTypeName(receiverType.X)
	case *ast.IndexListExpr:
		return receiverTypeName(receiverType.X)
	}
	return ""
}