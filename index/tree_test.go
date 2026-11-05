package index

import "testing"

func TestEnsureDirectoryNodeReusesSlashSegments(t *testing.T) {
	root := &IndexNode{Name: ".", Type: "dir"}

	if got := ensureDirectoryNode(root, "."); got != root {
		t.Fatalf("ensureDirectoryNode(root, \".\") did not return the root")
	}

	first := ensureDirectoryNode(root, "pkg/sub")
	if first.Name != "sub" {
		t.Fatalf("ensureDirectoryNode(root, \"pkg/sub\").Name = %q, want %q", first.Name, "sub")
	}
	if len(root.Children) != 1 || root.Children[0].Name != "pkg" {
		t.Fatalf("root has %d children, want one directory named pkg", len(root.Children))
	}
	pkg := root.Children[0]
	if len(pkg.Children) != 1 || pkg.Children[0] != first {
		t.Fatalf("pkg does not contain the sub directory node")
	}

	second := ensureDirectoryNode(root, "pkg/sub")
	if second != first {
		t.Errorf("second ensureDirectoryNode call returned a new node instead of the existing one")
	}
	if len(root.Children) != 1 || len(pkg.Children) != 1 {
		t.Errorf("second ensureDirectoryNode call duplicated directory nodes")
	}
}
