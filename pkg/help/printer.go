package help

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/kong"
	"go/doc"
	"strings"
)

const (
	defaultIndent        = 2
	defaultColumnPadding = 4
)

// Printer prints the help messages given the context.
func Printer(options kong.HelpOptions, ctx *kong.Context) (err error) {
	if ctx.Empty() {
		options.Summary = false
	}
	w := newHelpWriter(ctx, options)
	selected := ctx.Selected()
	if selected == nil {
		printApp(w, ctx.Model)
	} else {
		return kong.DefaultHelpPrinter(options, ctx)
	}
	return w.Write(ctx.Stdout)
}

func printApp(w *helpWriter, app *kong.Application) {
	w.Printf("Usage: %s%s", app.Name, app.Summary())
	printNodeDetail(w, app.Node, true)
	cmds := app.Leaves(true)
	if len(cmds) > 0 && app.HelpFlag != nil {
		w.Print("")
		if w.Summary {
			w.Printf(`Run "%s --help" for more information.`, app.Name)
		} else {
			w.Printf(`Run "%s <command> --help" for more information on a command.`, app.Name)
		}
	}
}

func printNodeDetail(w *helpWriter, node *kong.Node, hide bool) {
	if node.Help != "" {
		w.Print("")
		w.Wrap(node.Help)
	}
	if w.Summary {
		return
	}
	if node.Detail != "" {
		w.Print("")
		w.Wrap(node.Detail)
	}
	if len(node.Positional) > 0 {
		w.Print("")
		w.Print("Arguments:")
		writePositionals(w.Indent(), node.Positional)
	}
	if flags := node.AllFlags(true); len(flags) > 0 {
		w.Print("")
		w.Print("Flags:")
		writeFlags(w.Indent(), flags)
	}
	cmds := node.Leaves(hide)
	if len(cmds) > 0 {
		w.Print("")
		w.Print("Commands:")
		iw := w.Indent()
		if w.Compact {
			rows := [][2]string{}
			for _, cmd := range cmds {
				rows = append(rows, [2]string{cmd.Path(), cmd.Help})
			}
			writeTwoColumns(iw, defaultColumnPadding, rows)
		} else {
			for i, cmd := range cmds {
				printCommandSummary(iw, cmd)
				if i != len(cmds)-1 {
					iw.Print("")
				}
			}
		}
	}
}

func printCommandSummary(w *helpWriter, cmd *kong.Command) {
	w.Print(cmd.Summary())
	if cmd.Help != "" {
		w.Indent().Wrap(cmd.Help)
	}
}

func writePositionals(w *helpWriter, args []*kong.Positional) {
	rows := [][2]string{}
	for _, arg := range args {
		rows = append(rows, [2]string{arg.Summary(), arg.Help})
	}
	writeTwoColumns(w, defaultColumnPadding, rows)
}

func writeFlags(w *helpWriter, groups [][]*kong.Flag) {
	rows := [][2]string{}
	haveShort := false
	for _, group := range groups {
		for _, flag := range group {
			if flag.Short != 0 {
				haveShort = true
				break
			}
		}
	}
	for i, group := range groups {
		if i > 0 {
			rows = append(rows, [2]string{"", ""})
		}
		for _, flag := range group {
			if !flag.Hidden {
				rows = append(rows, [2]string{formatFlag(haveShort, flag), flag.Help})
			}
		}
	}
	writeTwoColumns(w, defaultColumnPadding, rows)
}

func writeTwoColumns(w *helpWriter, padding int, rows [][2]string) {
	maxLeft := 375 * w.width / 1000
	if maxLeft < 30 {
		maxLeft = 30
	}
	// Find size of first column.
	leftSize := 0
	for _, row := range rows {
		if c := len(row[0]); c > leftSize && c < maxLeft {
			leftSize = c
		}
	}

	offsetStr := strings.Repeat(" ", leftSize+padding)

	for _, row := range rows {
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, row[1], "", strings.Repeat(" ", defaultIndent), w.width-leftSize-padding)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")

		line := fmt.Sprintf("%-*s", leftSize, row[0])
		if len(row[0]) < maxLeft {
			line += fmt.Sprintf("%*s%s", padding, "", lines[0])
			lines = lines[1:]
		}
		w.Print(line)
		for _, line := range lines {
			w.Printf("%s%s", offsetStr, line)
		}
	}
}

// haveShort will be true if there are short flags present at all in the help. Useful for column alignment.
func formatFlag(haveShort bool, flag *kong.Flag) string {
	flagString := ""
	name := flag.Name
	isBool := flag.IsBool()
	if flag.Short != 0 {
		flagString += fmt.Sprintf("-%c, --%s", flag.Short, name)
	} else {
		if haveShort {
			flagString += fmt.Sprintf("    --%s", name)
		} else {
			flagString += fmt.Sprintf("--%s", name)
		}
	}
	if !isBool {
		flagString += fmt.Sprintf("=%s", flag.FormatPlaceHolder())
	}
	return flagString
}
