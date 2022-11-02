package build

import "fmt"

var Version string
var Hash string

func String() string {
	if Version != "" {
		return fmt.Sprintf("%v [%v]", Version, Hash)
	} else {
		return "<developer_build>"
	}
}
