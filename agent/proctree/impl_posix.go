//go:build !windows

package proctree

import (
	"os/exec"
	"sync"
)

func Init(_ string) error {
	return nil
}

func StartChild(tail TailFunction, args ...string) (*Child, error) {
	cmd := exec.Command(args[0], args[1:]...)

	cld := &Child{
		TailFunction: tail,
		cmd:          cmd,
		outStream:    make(chan []byte),
		errStream:    make(chan []byte),
		wg:           new(sync.WaitGroup),
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	cld.wg.Add(3)
	go reader(stdout, cld.outStream, cld.wg)
	go reader(stderr, cld.errStream, cld.wg)
	go cld.combiner(cld.wg)

	return cld, nil
}

func WaitChild(c *Child) error {
	c.wg.Wait()
	if err := c.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func StopChild(c *Child) error {
	if err := c.cmd.Process.Kill(); err != nil {
		return err
	}
	return nil
}
