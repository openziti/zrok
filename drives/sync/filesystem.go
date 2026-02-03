package sync

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/v2/drives/davServer"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
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
	return os.MkdirAll(filepath.Join(t.cfg.Root, path), os.ModePerm)
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
	return os.Open(filepath.Join(t.cfg.Root, path))
}

func (t *FilesystemTarget) WriteStream(path string, stream io.Reader, mode os.FileMode) error {
	targetPath := filepath.Join(t.cfg.Root, path)

	if err := os.MkdirAll(filepath.Dir(targetPath), mode); err != nil {
		return err
	}
	f, err := os.Create(targetPath)
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
	return t.WriteStream(path, stream, mode)
}

func (t *FilesystemTarget) Move(src, dest string) error {
	return os.Rename(filepath.Join(t.cfg.Root, src), filepath.Join(filepath.Dir(t.cfg.Root), dest))
}

func (t *FilesystemTarget) Rm(path string) error {
	return os.RemoveAll(filepath.Join(t.cfg.Root, path))
}

func (t *FilesystemTarget) SetModificationTime(path string, mtime time.Time) error {
	targetPath := filepath.Join(t.cfg.Root, path)
	if err := os.Chtimes(targetPath, time.Now(), mtime); err != nil {
		return err
	}
	return nil
}
