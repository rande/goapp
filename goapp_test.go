package goapp

import (
	"container/list"
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
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

func Test_LifeCycle(t *testing.T) {
	l := NewLifecycle()

	l.Add(Init, func(app *App) error {
		app.Set("hello", func(app *App) interface{} {
			return "world"
		})

		return nil
	})

	l.Add(Register, func(app *App) error {
		// The register need to value from the initial loop
		hello := app.GetString("hello") // so the hello service will be initialized

		assert.Equal(t, hello, "world")

		app.Set("testing", func(app *App) interface{} {
			return t
		})

		return nil
	})

	l.Add(Config, func(app *App) error {
		tester := app.Get("testing").(*testing.T) // always need to cast

		assert.Equal(t, tester, t)

		app.Set("messages", func(app *App) interface{} {
			return list.New()
		})

		return nil
	})

	l.Add(Run, func(app *App) error {
		messages := app.Get("messages").(*list.List)

		messages.PushBack("Run1: sleep for 0.5s")
		time.Sleep(500 * time.Millisecond)
		messages.PushBack("Run1: Wake up ...")

		return nil
	})

	l.Add(Run, func(app *App) error {
		messages := app.Get("messages").(*list.List)

		messages.PushBack("Run2: sleep for 0.5s")
		time.Sleep(500 * time.Millisecond)
		messages.PushBack("Run2: Wake up ...")

		return errors.New("this is a error")
	})

	l.Add(Run, func(app *App) error {
		panic("take this !!")
	})

	app := NewApp()

	r := l.Go(app)

	messages := app.Get("messages").(*list.List)
	assert.Equal(t, 4, messages.Len())

	assert.Equal(t, r, 1) // an error is present
}
