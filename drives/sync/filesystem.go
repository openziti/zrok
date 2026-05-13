package sync

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/openziti/zrok/v2/drives/davServer"
)

type FilesystemTargetConfig struct {
	Root string
}

type FilesystemTarget struct {
	cfg  *FilesystemTargetConfig
	root fs.FS
	tree []*Object
}

func NewFilesystemTarget(cfg *FilesystemTargetConfig) *FilesystemTarget {
	root := os.DirFS(cfg.Root)
	return &FilesystemTarget{cfg: cfg, root: root}
}

type filesystemReadCloser struct {
	io.ReadCloser
	root *os.Root
}

func (rc *filesystemReadCloser) Close() error {
	err := rc.ReadCloser.Close()
	rootErr := rc.root.Close()
	if err != nil {
		return err
	}
	return rootErr
}

func (t *FilesystemTarget) openRoot() (*os.Root, error) {
	return os.OpenRoot(t.cfg.Root)
}

func (t *FilesystemTarget) ensureRoot(mode os.FileMode) (*os.Root, error) {
	if err := os.MkdirAll(t.cfg.Root, mode); err != nil {
		return nil, err
	}
	return os.OpenRoot(t.cfg.Root)
}

func (t *FilesystemTarget) Inventory() ([]*Object, error) {
	fi, err := os.Stat(t.cfg.Root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		t.cfg.Root = filepath.Dir(t.cfg.Root)
		return []*Object{{
			Path:     "/" + fi.Name(),
			IsDir:    false,
			Size:     fi.Size(),
			Modified: fi.ModTime(),
		}}, nil
	}

	t.tree = nil
	if err := fs.WalkDir(t.root, ".", t.recurse); err != nil {
		return nil, err
	}
	return t.tree, nil
}

func (t *FilesystemTarget) Dir(path string) ([]*Object, error) {
	des, err := os.ReadDir(t.cfg.Root)
	if err != nil {
		return nil, err
	}
	var objects []*Object
	for _, de := range des {
		fi, err := de.Info()
		if err != nil {
			return nil, err
		}
		objects = append(objects, &Object{
			Path:     de.Name(),
			IsDir:    de.IsDir(),
			Size:     fi.Size(),
			Modified: fi.ModTime(),
		})
	}
	return objects, nil
}

func (t *FilesystemTarget) Mkdir(path string) error {
	localName, err := localNameFromVirtualPath(path)
	if err != nil {
		return err
	}

	root, err := t.ensureRoot(os.ModePerm)
	if err != nil {
		return err
	}
	defer root.Close()

	return root.MkdirAll(localName, os.ModePerm)
}

func (t *FilesystemTarget) recurse(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	fi, err := d.Info()
	if err != nil {
		return err
	}
	etag := ""
	if v, ok := fi.(davServer.ETager); ok {
		etag, err = v.ETag(context.Background())
		if err != nil {
			return err
		}
	} else {
		etag = fmt.Sprintf(`"%x%x"`, fi.ModTime().UTC().UnixNano(), fi.Size())
	}
	if path != "." {
		outPath := "/" + path
		if fi.IsDir() {
			outPath = outPath + "/"
		}
		t.tree = append(t.tree, &Object{
			Path:     outPath,
			IsDir:    fi.IsDir(),
			Size:     fi.Size(),
			Modified: fi.ModTime(),
			ETag:     etag,
		})
	}
	return nil
}

func (t *FilesystemTarget) ReadStream(path string) (io.ReadCloser, error) {
	localName, err := localNameFromVirtualPath(path)
	if err != nil {
		return nil, err
	}

	root, err := t.openRoot()
	if err != nil {
		return nil, err
	}
	f, err := root.Open(localName)
	if err != nil {
		root.Close()
		return nil, err
	}
	return &filesystemReadCloser{ReadCloser: f, root: root}, nil
}

func (t *FilesystemTarget) WriteStream(path string, stream io.Reader, mode os.FileMode) error {
	localName, err := localNameFromVirtualPath(path)
	if err != nil {
		return err
	}

	root, err := t.ensureRoot(mode)
	if err != nil {
		return err
	}
	defer root.Close()

	if err := root.MkdirAll(filepath.Dir(localName), mode); err != nil {
		return err
	}
	f, err := root.Create(localName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, stream)
	if err != nil {
		return err
	}
	return nil
}

func (t *FilesystemTarget) WriteStreamWithModTime(path string, stream io.Reader, mode os.FileMode, modTime time.Time) error {
	if err := t.WriteStream(path, stream, mode); err != nil {
		return err
	}
	return t.SetModificationTime(path, modTime)
}

func (t *FilesystemTarget) Move(src, dest string) error {
	srcName, err := localNameFromVirtualPath(src)
	if err != nil {
		return err
	}
	destName, err := localNameFromVirtualPath(dest)
	if err != nil {
		return err
	}

	if srcName == "." {
		parent := filepath.Dir(t.cfg.Root)
		parentRoot, err := os.OpenRoot(parent)
		if err != nil {
			return err
		}
		defer parentRoot.Close()

		rootName, err := filepath.Localize(filepath.Base(t.cfg.Root))
		if err != nil {
			return unsafePathError(t.cfg.Root)
		}
		return parentRoot.Rename(rootName, destName)
	}

	root, err := t.openRoot()
	if err != nil {
		return err
	}
	defer root.Close()
	return root.Rename(srcName, destName)
}

func (t *FilesystemTarget) Rm(path string) error {
	localName, err := localNameFromVirtualPath(path)
	if err != nil {
		return err
	}
	if localName == "." {
		return os.RemoveAll(t.cfg.Root)
	}

	root, err := t.openRoot()
	if err != nil {
		return err
	}
	defer root.Close()
	return root.RemoveAll(localName)
}

func (t *FilesystemTarget) SetModificationTime(path string, mtime time.Time) error {
	localName, err := localNameFromVirtualPath(path)
	if err != nil {
		return err
	}

	root, err := t.openRoot()
	if err != nil {
		return err
	}
	defer root.Close()

	if err := root.Chtimes(localName, time.Now(), mtime); err != nil {
		return err
	}
	return nil
}
