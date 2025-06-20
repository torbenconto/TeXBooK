package datasources

import (
	"os"
	"path/filepath"
)

type DataSource interface {
	Connect() error
	Disconnect() error
	Type() string
	Path() string
	// uuid
	ID() uint32
	Metadata() map[string]any
	ListFiles() (FileNode, error)
}

type BaseDataSource struct {
	SourceID uint32
}

func (b *BaseDataSource) ID() uint32 {
	return b.SourceID
}

type LocalDataSource struct {
	BaseDataSource
	SourcePath string
}

func (l *LocalDataSource) Connect() error {
	return nil
}

func (l *LocalDataSource) Disconnect() error {
	return nil
}

func (l *LocalDataSource) Type() string {
	return "local"
}

func (l *LocalDataSource) Path() string {
	return l.SourcePath
}

func (l *LocalDataSource) Metadata() map[string]any {
	return map[string]any{"path": l.Path()}
}

type FileNode struct {
	Name     string     `json:"name"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

func buildTree(root string) (FileNode, error) {
	info, err := os.Stat(root)
	if err != nil {
		return FileNode{}, err
	}

	if !info.IsDir() {
		if filepath.Ext(info.Name()) == ".tex" {
			return FileNode{
				Name:  info.Name(),
				IsDir: false,
			}, nil
		}
		return FileNode{}, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return FileNode{}, err
	}

	node := FileNode{
		Name:  info.Name(),
		IsDir: true,
	}

	for _, entry := range entries {
		childPath := filepath.Join(root, entry.Name())
		childNode, err := buildTree(childPath)
		if err != nil {
			return FileNode{}, err
		}
		if childNode.Name != "" {
			node.Children = append(node.Children, childNode)
		}
	}

	if len(node.Children) == 0 {
		return FileNode{}, nil
	}

	return node, nil
}

func (l *LocalDataSource) ListFiles() (FileNode, error) {
	tree, err := buildTree(l.Path())
	if err != nil {
		return FileNode{}, err
	}

	return tree, nil
}
