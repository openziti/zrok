package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

var testA = "[13.845] DEBUG sdk-golang/ziti/edge/impl. (*edgeConn) .Accept: {edgeSeq= [2] vid=[invalid- vid-size-of-0-bytes] connId= [2147483694] type= [EdgeDataType] chSeq= [7]} receivedddddd 567 bytes (msg type: 60786)"

var testB = "[] Hello this is a test to see if word wrapping works propery. {Need to pad space. WhathappensifIputareallylongwordwithnobreakswillitcutthewordorwillitproperlyloverflow and then (a new line)} ()ß"

func init() {
	testCmd.AddCommand(newShareTestCommand().cmd)
}

type shareTestCommand struct {
	headless bool
	cmd      *cobra.Command
}

func newShareTestCommand() *shareTestCommand {
	cmd := &cobra.Command{
		Use:   "share <target>",
		Short: "Share a target resource publicly",
		Args:  cobra.ExactArgs(1),
	}
	command := &shareTestCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")

	cmd.Run = command.run
	return command
}

func (cmd *shareTestCommand) run(_ *cobra.Command, args []string) {
	w := 167
	if !cmd.headless {
		mdl := newShareModel("token", []string{"Endpoints"}, "shareMode", "backendMode")
		prg := tea.NewProgram(mdl, tea.WithAltScreen())
		mdl.prg = prg

		go func() {
			mdl.Write([]byte(testA))
			mdl.Write([]byte(testB))
		}()

		if _, err := prg.Run(); err != nil {
			tui.Error("An error occurred", err)
		}
	} else {
		for i := w; i >= 0; i = i - 10 {
			fmt.Println("-----------------------------")
			fmt.Printf("Width: %d\n", i)
			out, n := plainTextRender(i, testA)
			fmt.Println(out)
			fmt.Printf("Expected lines: %d\n", n)
			fmt.Println("-----------------------------")
		}
	}
}

func plainTextRender(width int, logs ...string) (string, int) {
	var splitLines []string
	for _, line := range logs {
		wrapped := wordwrap.String(line, width)

		wrappedLines := strings.Split(wrapped, "\n")
		for _, wrappedLine := range wrappedLines {
			splitLine := strings.ReplaceAll(wrappedLine, "\n", "")
			if splitLine != "" {
				splitLines = append(splitLines, splitLine)
			}
		}
	}
	maxRows := shareLogStyle.GetHeight()
	maxRows = 999
	startRow := 0
	if len(splitLines) > maxRows {
		startRow = len(splitLines) - maxRows
	}
	out := ""
	n := 0
	for i := startRow; i < len(splitLines); i++ {
		outLine := splitLines[i]
		n++
		if i < len(splitLines)-1 {
			outLine += "\n"
		}
		out += outLine
	}
	return out, n

}
