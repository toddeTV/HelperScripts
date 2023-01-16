package greet

import (
	"flag"
	"fmt"
)

type Command struct {
	fs *flag.FlagSet

	name string
}

func Cmd() *Command {
	gc := &Command{
		fs: flag.NewFlagSet("greet", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.name, "name", "World", "name of the person to be greeted")

	return gc
}

func (g *Command) Name() string {
	return g.fs.Name()
}

func (g *Command) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *Command) Run() error {
	fmt.Println("Hello", g.name, "!")
	return nil
}
