package goapp

import (
	"fmt"
)

// The structure contains all services build by the AppFunc function, the service is initialized when get Get
// method is called.
type App struct {
	state    int
	values   map[string]AppFunc     // contains the original closure to generate the service
	services map[string]interface{} // contains the instantiated services
}

type AppFunc func(app *App) interface{}

func NewApp() *App {
	app := App{
		services: make(map[string]interface{}),
		values:   make(map[string]AppFunc),
	}

	return &app
}

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

func (app *App) GetString(name string) string {
	return app.Get(name).(string)
}

func (app *App) GetState() int {
	return app.state
}

func (app *App) IsTerminated() bool {
	return app.state == Terminated
}
