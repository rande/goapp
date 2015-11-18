package goapp

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

type User struct {
	name string
}

func Test_App(t *testing.T) {

	app := NewApp()

	app.Set("logger", func(app *App) interface{} {
		return log.New(os.Stdout, "prefix", log.LstdFlags)
	})

	app.Set("map", func(app *App) interface{} {
		m := make(map[string]string)
		m["test"] = "test"

		return m
	})

	app.Set("key", func(app *App) interface{} {
		return "Salut"
	})

	assert.IsType(t, app.Get("logger").(*log.Logger), new(log.Logger))

	m := app.Get("map").(map[string]string)
	m["foo"] = "bar"
	m["test"] = "hello"

	a := app.Get("map").(map[string]string)
	assert.Equal(t, a["test"], "hello")

	assert.Equal(t, "Salut", app.GetString("key"))
}

func Test_NoReference(t *testing.T) {
	app := NewApp()

	app.Set("user", func(app *App) interface{} {
		return User{
			name: "Thomas",
		}
	})

	app.Set("user.reference", func(app *App) interface{} {
		return &User{
			name: "Thomas",
		}
	})

	u := app.Get("user").(User)
	u.name = "Fred"

	u = app.Get("user").(User)
	assert.NotEqual(t, "Fred", u.name)

	u_r := app.Get("user.reference").(*User)
	u_r.name = "Fred"

	u_r = app.Get("user.reference").(*User)
	assert.Equal(t, "Fred", u_r.name)
}

func Test_GetKeys(t *testing.T) {
	app := NewApp()
	app.Set("user", func(app *App) interface{} { return nil })
	app.Set("user.reference", func(app *App) interface{} { return nil })

	assert.Contains(t, app.GetKeys(), "user")
	assert.Contains(t, app.GetKeys(), "user.reference")
}
