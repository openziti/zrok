package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
)

func helloZrok(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello zrok!\n")
}

func main() {
	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	shr, err := sdk.CreateShare(root, &sdk.ShareRequest{
		BackendMode:         sdk.ProxyBackendMode,
		ShareMode:           sdk.PublicShareMode,
		NamespaceSelections: []sdk.NamespaceSelection{{NamespaceToken: "public"}},
		Target:              "http-server",
	})

	if err != nil {
		panic(err)
	}
	defer func() {
		if err := sdk.DeleteShare(root, shr); err != nil {
			panic(err)
		}
	}()

	conn, err := sdk.NewListener(shr.Token, root)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Access server at the following endpoints: ", strings.Join(shr.FrontendEndpoints, "\n"))

	http.HandleFunc("/", helloZrok)

	if err := http.Serve(conn, nil); err != nil {
		panic(err)
	}
}
