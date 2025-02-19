package proctree

import (
	"fmt"
	_ "github.com/kolesnikovae/go-winjob"
	"io"
	"os/exec"
	"sync"
)

type Child struct {
	TailFunction TailFunction
	cmd          *exec.Cmd
	outStream    chan []byte
	errStream    chan []byte
	wg           *sync.WaitGroup
}

type TailFunction func(data []byte)

func (c *Child) combiner(wg *sync.WaitGroup) {
	defer wg.Done()

	outDone := false
	errDone := false
	for {
		select {
		case data := <-c.outStream:
			if data != nil {
				if c.TailFunction != nil {
					c.TailFunction(data)
				}
			} else {
				outDone = true
			}
		case data := <-c.errStream:
			if data != nil {
				if c.TailFunction != nil {
					c.TailFunction(data)
				}
			} else {
				errDone = true
			}
		}
		if outDone && errDone {
			return
		}
	}
}

func reader(r io.ReadCloser, o chan []byte, wg *sync.WaitGroup) {
	defer close(o)
	defer wg.Done()

	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("error reading: %v", err)
			return
		}
		o <- buf[:n]
	}
}
