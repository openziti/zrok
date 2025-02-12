package build

import "fmt"

var Version string
var Hash string

const Series = "v1.0"

func String() string {
	if Version != "" {
		return fmt.Sprintf("%v [%v]", Version, Hash)
	} else {
		return Series + ".x [developer build]"
	}
}
