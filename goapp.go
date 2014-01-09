package goapp

type App struct {
    // contains the original closure to generate the service
    values map[string]AppFunc

    // contains the instanciated services
    services map[string]interface{}
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
        panic("The service does not exist")
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
        values: make(map[string]AppFunc),
    }

    return &app
}