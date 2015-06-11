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
	Init     = 10 // Initialize application: register flags, no logic should be done here
	Register = 20 // Register components that does not required configuration settings
	Config   = 30 // Read configuration and defined main services from this configuration
	Run      = 40 // Run the main program loop
	Exit     = 50 // Exit the program
)

// The structure contains all services build by the AppFunc function, the service is initialized when get Get
// method is called.
type App struct {
	state    int
	values   map[string]AppFunc     // contains the original closure to generate the service
	services map[string]interface{} // contains the instantiated services
}

type AppFunc func(app *App) interface{}

func (app *App) Set(name string, f AppFunc) {
	if _, ok := app.services[name]; ok {
		panic("Cannot overwrite initialized service")
	}

	app.values[name] = f
}

func (app *App) Get(name string) interface{} {
	if _, ok := app.values[name]; !ok {
		panic(fmt.Sprintf("The service does not exist: %s", name))
	}

	if _, ok := app.services[name]; !ok {
		app.services[name] = app.values[name](app)
	}

	return app.services[name]
}

func (app *App) GetString(name string) interface{} {
	return app.Get(name).(string)
}

func NewApp() *App {
	app := App{
		services: make(map[string]interface{}),
		values:   make(map[string]AppFunc),
	}

	return &app
}

type LifecycleFun func(app *App) error

type Lifecycle struct {
	init     []LifecycleFun
	register []LifecycleFun
	config   []LifecycleFun
	run      []LifecycleFun
	exit     []LifecycleFun
}

// register a new LifecycleFun function to a step
func (l *Lifecycle) Add(t int, f LifecycleFun) {
	switch t {
	case Init:
		l.init = append(l.init, f)
	case Register:
		l.register = append(l.register, f)
	case Config:
		l.config = append(l.config, f)
	case Run:
		l.run = append(l.run, f)
	case Exit:
		l.exit = append(l.exit, f)
	}
}

// Start the different step, the each LifecycleFun defined in the run step will be started in a dedicated goroutine.
// The Go function will exit the program
func (l *Lifecycle) Go(app *App) int {

	fmt.Printf(">>> GoApp vX.X\n")
	fmt.Printf(">>> GoApp - Init\n")
	app.state = Init
	for _, f := range l.init {
		f(app)
	}

	fmt.Printf(">>> GoApp - Register\n")
	app.state = Register
	for _, f := range l.register {
		f(app)
	}

	fmt.Printf(">>> GoApp - Configuration\n")
	app.state = Config
	for _, f := range l.config {
		f(app)
	}

	fmt.Printf(">>> GoApp - Run\n")
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
			fmt.Printf(">>> GoApp - Error: %s\n", err)
			hasError = true
		}
	}

	fmt.Printf(">>> GoApp - Exit\n")
	app.state = Exit
	for _, f := range l.exit {
		f(app)
	}

	fmt.Printf(">>> GoApp - done!\n")

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
