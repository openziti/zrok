package main

import "os/exec"

func openBrowser(url string) error {
	return exec.Command("open", url).Run()
}
