package main

import (
	"golang.org/x/net/webdav"
	"log"
	"net/http"
)

func main() {
	dav := &webdav.Handler{
		FileSystem: webdav.Dir("."),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
			}
		},
	}
	http.Handle("/", dav)
	if err := http.ListenAndServe("0.0.0.0:8800", nil); err != nil {
		log.Fatalf("error serving: ")
	}
}
