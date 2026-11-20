package index

import "testing"

func TestEnsureDirectoryNodeReusesSlashSegments(t *testing.T) {
	root := &indexDir{Name: "."}

	if got := ensureDirectoryNode(root, "."); got != root {
		t.Fatalf("ensureDirectoryNode(root, \".\") did not return the root")
	}

	first := ensureDirectoryNode(root, "pkg/sub")
	if first.Name != "sub" {
		t.Fatalf("ensureDirectoryNode(root, \"pkg/sub\").Name = %q, want %q", first.Name, "sub")
	}
	if len(root.Children) != 1 {
		t.Fatalf("root has %d children, want one directory", len(root.Children))
	}
	pkg, ok := root.Children[0].(*indexDir)
	if !ok || pkg.Name != "pkg" {
		t.Fatalf("root does not contain a directory named pkg")
	}
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
