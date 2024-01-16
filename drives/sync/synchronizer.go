package sync

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

func OneWay(src, dst Target, sync bool) error {
	srcTree, err := src.Inventory()
	if err != nil {
		return errors.Wrap(err, "error creating source inventory")
	}

	var dstTree []*Object
	if sync {
		dstTree, err = dst.Inventory()
		if err != nil {
			return errors.Wrap(err, "error creating destination inventory")
		}
	}

	dstIndex := make(map[string]*Object)
	for _, f := range dstTree {
		dstIndex[f.Path] = f
	}

	var copyList []*Object
	for _, srcF := range srcTree {
		if dstF, found := dstIndex[srcF.Path]; found {
			if !srcF.IsDir && (dstF.Size != srcF.Size || dstF.Modified.Unix() != srcF.Modified.Unix()) {
				logrus.Debugf("%v <- dstF.Size = '%d', srcF.Size = '%d', dstF.Modified.UTC = '%d', srcF.Modified.UTC = '%d'", srcF.Path, dstF.Size, srcF.Size, dstF.Modified.Unix(), srcF.Modified.Unix())
				copyList = append(copyList, srcF)
			}
		} else {
			logrus.Debugf("%v <- !found", srcF.Path)
			copyList = append(copyList, srcF)
		}
	}

	for _, copyPath := range copyList {
		if copyPath.IsDir {
			if err := dst.Mkdir(copyPath.Path); err != nil {
				return err
			}
		} else {
			ss, err := src.ReadStream(copyPath.Path)
			if err != nil {
				return err
			}
			if err := dst.WriteStreamWithModTime(copyPath.Path, ss, os.ModePerm, copyPath.Modified); err != nil {
				return err
			}
		}
		logrus.Infof("=> %v", copyPath.Path)
	}

	return nil
}
