package requester

import (
	"flag"
	"fmt"
)

type Action struct {
	Handler func() error
	Name    string
	Help    string
}

func Actions(actions []*Action) {
	flags := make([]*bool, len(actions))
	for i, action := range actions {
		if action.Name == "" {
			panic("all actions should have names")
		}
		if action.Help == "" {
			panic("all actions should have help")
		}
		if action.Handler == nil {
			panic("all actions should have helpers")
		}

		flags[i] = flag.Bool(action.Name, false, action.Help)
	}
	flag.Parse()

	found := false
	for i, action := range actions {
		if *flags[i] {
			found = true
			if err := action.Handler(); err != nil {
				fmt.Println(err)
			}
			break
		}
	}

	if !found {
		flag.PrintDefaults()
		return
	}

	if err := CloseLibrary(); err != nil {
		fmt.Println(err)
	}

	// TODO: Print stats on exit
}
