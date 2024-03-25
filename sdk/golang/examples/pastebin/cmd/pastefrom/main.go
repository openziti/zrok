package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"os"
)

const MAX_PASTE_SIZE = 64 * 1024

func main() {
	if len(os.Args) < 2 {
		panic("usage: pastefrom <shrToken>")
	}
	shrToken := os.Args[1]

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	acc, err := sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: shrToken})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := sdk.DeleteAccess(root, acc); err != nil {
			panic(err)
		}
	}()

	conn, err := sdk.NewDialer(shrToken, root)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	buf := make([]byte, MAX_PASTE_SIZE)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	fmt.Printf(string(buf[:n]))
}
