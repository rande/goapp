// The package try to normalize how an application should start without
// providing any convention about how each steps should be used.
//
// The process will go through 5 steps:
//  - Init: Initialize application: register flags, no logic should be done here
//  - Register : Register components that does not required configuration settings
//  - Config   : Read configuration and defined main services from this configuration
//  - Run      : Run the main program loop
//  - Exit     : (not yet implemented
package goapp

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

const (
	Init       = 10 // Initialize application: register flags, no logic should be done here
	Register   = 20 // Register components that does not required configuration settings
	Config     = 30 // Read configuration
	Prepare    = 40 // Defined main services from configuration
	Run        = 50 // Run the main program loop
	Exit       = 60 // Exit the program
	Terminated = 70 // Exit the program
)

// Represents the function used to change states or start a process.
// Please note the process should be blocking as it will run inside a goroutine
type LifecycleFun func(app *App) error

type Lifecycle struct {
	init     []LifecycleFun
	register []LifecycleFun
	config   []LifecycleFun
	prepare  []LifecycleFun
	run      []LifecycleFun
	exit     []LifecycleFun
}

// register a new LifecycleFun function to a step
func (l *Lifecycle) add(t int, f LifecycleFun) {
	switch t {
	case Init:
		l.init = append(l.init, f)
	case Register:
		l.register = append(l.register, f)
	case Config:
		l.config = append(l.config, f)
	case Prepare:
		l.prepare = append(l.prepare, f)
	case Run:
		l.run = append(l.run, f)
	case Exit:
		l.exit = append(l.exit, f)
	}
}

func (l *Lifecycle) Init(f LifecycleFun) {
	l.add(Init, f)
}

func (l *Lifecycle) Register(f LifecycleFun) {
	l.add(Register, f)
}

func (l *Lifecycle) Config(f LifecycleFun) {
	l.add(Config, f)
}

func (l *Lifecycle) Prepare(f LifecycleFun) {
	l.add(Config, f)
}

func (l *Lifecycle) Run(f LifecycleFun) {
	l.add(Run, f)
}

func (l *Lifecycle) Exit(f LifecycleFun) {
	l.add(Exit, f)
}

func (l *Lifecycle) execute(fs []LifecycleFun, app *App) {
	for _, f := range fs {
		f(app)
	}
}

// Start the different step, each LifecycleFun defined in the "run" step will be started in a dedicated goroutine.
// The Go function will return the exit code of the program
func (l *Lifecycle) Go(app *App) int {

	app.state = Init
	l.execute(l.init, app)

	app.state = Register
	l.execute(l.register, app)

	app.state = Config
	l.execute(l.config, app)

	app.state = Prepare
	l.execute(l.prepare, app)

	app.state = Run
	var wg sync.WaitGroup

	results := make([]chan error, len(l.run)) // create a pool of channel to handle each return

	for p, f := range l.run {
		results[p] = make(chan error, 1)

		wg.Add(1)
		go func(f LifecycleFun, c chan error) {
			defer func() {
				if r := recover(); r != nil {
					message := fmt.Sprintf("Panic recovered, message=%s\n", r)
					c <- errors.New(message + string(debug.Stack()[:]))
				}

				wg.Done()
			}()

			c <- f(app)
		}(f, results[p])
	}

	wg.Wait()

	hasError := false
	for _, c := range results {
		err := <-c

		if err != nil {
			fmt.Printf(">>> Error: %s\n", err)
			hasError = true
		}
	}

	app.state = Exit
	l.execute(l.exit, app)

	app.state = Terminated

	// check for errors
	if hasError {
		return 1
	}

	return 0
}

func NewLifecycle() *Lifecycle {
	return &Lifecycle{
		init:     make([]LifecycleFun, 0),
		register: make([]LifecycleFun, 0),
		config:   make([]LifecycleFun, 0),
		run:      make([]LifecycleFun, 0),
		exit:     make([]LifecycleFun, 0),
	}
}
