package sync

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func unsafePathError(p string) error {
	return fmt.Errorf("unsafe path '%s'", p)
}

func cleanVirtualPath(p string) (string, error) {
	if p == "" {
		p = "/"
	}
	if strings.Contains(p, "\x00") {
		return "", unsafePathError(p)
	}
	for _, segment := range strings.Split(p, "/") {
		if segment == ".." {
			return "", unsafePathError(p)
		}
	}
	return path.Clean("/" + p), nil
}

func objectPathFromVirtualPath(p string, isDir bool) (string, error) {
	clean, err := cleanVirtualPath(p)
	if err != nil {
		return "", err
	}
	if clean == "/" {
		return "/", nil
	}

	rel := strings.TrimPrefix(clean, "/")
	if _, err := filepath.Localize(rel); err != nil {
		return "", unsafePathError(p)
	}

	if isDir {
		return clean + "/", nil
	}
	return clean, nil
}

func localNameFromVirtualPath(p string) (string, error) {
	clean, err := cleanVirtualPath(p)
	if err != nil {
		return "", err
	}
	if clean == "/" {
		return ".", nil
	}

	localName, err := filepath.Localize(strings.TrimPrefix(clean, "/"))
	if err != nil {
		return "", unsafePathError(p)
	}
	return localName, nil
}

func remoteObjectPath(rootPath, hrefPath string, isDir bool) (string, error) {
	root, err := cleanVirtualPath(rootPath)
	if err != nil {
		return "", err
	}
	href, err := cleanVirtualPath(hrefPath)
	if err != nil {
		return "", err
	}

	if root != "/" && href != root && !strings.HasPrefix(href, root+"/") {
		return "", unsafePathError(hrefPath)
	}

	rel := href
	if root != "/" {
		if href == root {
			rel = "/"
		} else {
			rel = "/" + strings.TrimPrefix(href, root+"/")
		}
	}

	return objectPathFromVirtualPath(rel, isDir)
}

func remoteFileObjectPath(requestPath, hrefPath string) (string, error) {
	request, err := cleanVirtualPath(requestPath)
	if err != nil {
		return "", err
	}
	href, err := cleanVirtualPath(hrefPath)
	if err != nil {
		return "", err
	}
	if href != request {
		return "", unsafePathError(hrefPath)
	}
	return remoteObjectPath(path.Dir(request), href, false)
}

func joinRemotePath(rootPath, objectPath string) (string, error) {
	root, err := cleanVirtualPath(rootPath)
	if err != nil {
		return "", err
	}
	object, err := cleanVirtualPath(objectPath)
	if err != nil {
		return "", err
	}
	if object == "/" {
		return root, nil
	}
	return path.Join(root, object), nil
}
