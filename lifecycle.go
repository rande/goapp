// The package try to normalize how an application should start without
// providing any convention about how each steps should be used.
//
// The process will go through 6 steps:
//  - Init: Initialize application: register flags, no logic should be done here
//  - Register : Register components that does not required configuration settings
//  - Config   : Read configuration
//  - Prepare  : Defined main services from this configuration
//  - Run      : Run the main program loop, each function will be run in a goroutine
//  - Exit     : Register function call when the program will exit
package goapp

import (
	"errors"
	"fmt"
	"reflect"
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

type LifecycleRunFun func(app *App, state *GoroutineState) error

type Lifecycle struct {
	init     []LifecycleFun
	register []LifecycleFun
	config   []LifecycleFun
	prepare  []LifecycleFun
	run      []LifecycleRunFun
	exit     []LifecycleFun
}

// register a new LifecycleFun function to a step
func (l *Lifecycle) add(t int, f interface{}) {
	switch f := f.(type) {
	case LifecycleRunFun:
		l.run = append(l.run, f)
	case LifecycleFun:
		switch t {
		case Run:
			panic("You must use a LifecycleRunFun type for the Run type")
		case Init:
			l.init = append(l.init, f)
		case Register:
			l.register = append(l.register, f)
		case Config:
			l.config = append(l.config, f)
		case Prepare:
			l.prepare = append(l.prepare, f)
		case Exit:
			l.exit = append(l.exit, f)
		}
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

func (l *Lifecycle) Run(f LifecycleRunFun) {
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

	// start a set of goroutine, and provide a GoroutingState struct to interact wih the In and Out channel
	states := make([]*GoroutineState, len(l.run)) // create a pool of channel to handle message from tasks
	for p, f := range l.run {
		states[p] = NewGoroutineState()

		wg.Add(1)

		go func(f LifecycleRunFun, state *GoroutineState) {
			defer func() {
				if r := recover(); r != nil {
					message := fmt.Sprintf("Panic recovered, message=%s\n", r)
					state.Error = errors.New(message + string(debug.Stack()[:]))

					state.Out <- Control_Stop
				}

				wg.Done()
			}()

			state.Error = f(app, state)
		}(f, states[p])
	}

	// Start listenning to Out channel
	go func(states []*GoroutineState) {
		cases := make([]reflect.SelectCase, len(states))
		for i, state := range states {
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(state.Out),
			}
		}

		chosen, _, _ := reflect.Select(cases)

		fmt.Println("Sending exit signal to goroutine")
		for pos, state := range states {
			if pos == chosen {
				continue
			}

			state.In <- Control_Stop
		}
	}(states)

	wg.Wait()

	hasError := false
	for _, state := range states {

		if state.Error != nil {

			fmt.Printf(">>> Error: %s\n", state.Error)
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
		run:      make([]LifecycleRunFun, 0),
		exit:     make([]LifecycleFun, 0),
	}
}
