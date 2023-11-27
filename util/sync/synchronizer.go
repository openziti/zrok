package sync

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
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

	for _, target := range copyList {
		logrus.Infof("+> %v", target.Path)
		ss, err := src.ReadStream(target.Path)
		if err != nil {
			return err
		}
		if err := dst.WriteStream(target.Path, ss, os.ModePerm); err != nil {
			return err
		}
		logrus.Infof("=> %v", target.Path)
	}

	return nil
}
