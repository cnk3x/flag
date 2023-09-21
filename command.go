package flag

import (
	"bytes"
	"fmt"
	"os"
)

type App struct {
	Name        string
	Description string
	Commands    []*Command
}

type Command struct {
	Name   string
	Usage  string
	Flags  *FlagSet
	Action func(flags *FlagSet)
}

func New(name, description string) *App {
	return &App{Name: name, Description: description}
}

func (app *App) AddCommand(name, usage string, options ...Option) *Command {
	cmd := &Command{Name: name, Usage: usage}
	cmd.Flags = NewFlagSet(app.Name+" "+name, ContinueOnError)
	for _, fOpt := range options {
		fOpt.Apply(cmd)
	}
	app.Commands = append(app.Commands, cmd)
	return cmd
}

func (app *App) GetUsage() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s - %s\n", app.Name, app.Description)
	buf.WriteByte('\n')
	fmt.Fprintf(&buf, "Usage: \n  %s [Command] [...Options] [...Args]\n", app.Name)
	buf.WriteByte('\n')
	fmt.Fprintf(&buf, "Commands:\n")
	var nw int
	for _, cmd := range app.Commands {
		nw = max(nw, len(cmd.Name))
	}
	for _, cmd := range app.Commands {
		fmt.Fprintf(&buf, "  %*s  %s\n", -nw, cmd.Name, cmd.Usage)
	}
	return buf.String()
}

func (cmd *Command) GetUsage() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Usage:\n  %s [...options]\n\nOptions:\n", cmd.Flags.Name())
	fmt.Fprintln(&buf, cmd.Flags.FlagUsages())
	buf.Truncate(buf.Len() - 1)
	return buf.String()
}

func (app *App) RunWith(args []string) {
	var name string
	if len(args) > 0 {
		name = args[0]
	}

	if name == "help" || name == "" || name == "-h" || name == "--help" {
		fmt.Fprintln(os.Stderr, app.GetUsage())
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "Run '%s COMMAND --help' for more information on a command.\n", app.Name)
		os.Exit(2)
		return
	}

	var cmd *Command
	for _, c := range app.Commands {
		if c.Name == name {
			cmd = c
		}
	}
	if cmd == nil {
		fmt.Fprintln(os.Stderr, app.GetUsage())
		fmt.Fprintf(os.Stderr, "unknown command %s\n", name)
		fmt.Fprintln(os.Stderr)
		os.Exit(2)
		return
	}

	if cmd.Flags != nil {
		cmd.Flags.Init(app.Name+" "+cmd.Name, ContinueOnError)
		if err := cmd.Flags.Parse(args[1:]); err != nil {
			if err == ErrHelp {
				fmt.Fprintln(os.Stderr)
				os.Exit(0)
			} else {
				fmt.Fprintln(os.Stderr, cmd.GetUsage())
				fmt.Fprintln(os.Stderr)
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(2)
			}
			return
		}
	}

	if cmd.Action == nil {
		fmt.Fprintln(os.Stderr, cmd.GetUsage())
		fmt.Fprintln(os.Stderr)
	} else {
		cmd.Action(cmd.Flags)
	}
}

func (app *App) Run() {
	app.RunWith(os.Args[1:])
}

type Option func(cmd *Command)

func (fOpt Option) Apply(cmd *Command) { fOpt(cmd) }

func Action(fRun func(flags *FlagSet)) Option { return func(cmd *Command) { cmd.Action = fRun } }
func Flags(applyFlag ...func(flags *FlagSet)) Option {
	return func(cmd *Command) {
		for _, af := range applyFlag {
			af(cmd.Flags)
		}
	}
}
