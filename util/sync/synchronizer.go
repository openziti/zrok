package sync

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Synchronize(src, dst Target) error {
	srcTree, err := src.Inventory()
	if err != nil {
		return errors.Wrap(err, "error creating source inventory")
	}

	dstTree, err := dst.Inventory()
	if err != nil {
		return errors.Wrap(err, "error creating destination inventory")
	}

	dstIndex := make(map[string]*Object)
	for _, f := range dstTree {
		dstIndex[f.Path] = f
	}

	var copyList []*Object
	for _, srcF := range srcTree {
		if dstF, found := dstIndex[srcF.Path]; found {
			if dstF.ETag != srcF.ETag {
				copyList = append(copyList, srcF)
			}
		} else {
			copyList = append(copyList, srcF)
		}
	}

	logrus.Infof("files to copy:")
	for _, copy := range copyList {
		logrus.Infof("-> %v", copy.Path)
	}

	return nil
}
