package sync

import (
	"context"
	"fmt"
	"golang.org/x/net/webdav"
	"io/fs"
	"os"
)

type FilesystemTargetConfig struct {
	Root string
}

type FilesystemTarget struct {
	root fs.FS
	tree []*Object
}

func NewFilesystemTarget(cfg *FilesystemTargetConfig) *FilesystemTarget {
	root := os.DirFS(cfg.Root)
	return &FilesystemTarget{root: root}
}

func (t *FilesystemTarget) Inventory() ([]*Object, error) {
	t.tree = nil
	if err := fs.WalkDir(t.root, ".", t.recurse); err != nil {
		return nil, err
	}
	return t.tree, nil
}

func (t *FilesystemTarget) recurse(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		fi, err := d.Info()
		if err != nil {
			return err
		}
		etag := ""
		if v, ok := fi.(webdav.ETager); ok {
			etag, err = v.ETag(context.Background())
			if err != nil {
				return err
			}
		} else {
			etag = fmt.Sprintf(`"%x%x"`, fi.ModTime().UnixNano(), fi.Size())
		}
		t.tree = append(t.tree, &Object{path, fi.Size(), fi.ModTime(), etag})
	}
	return nil
}
