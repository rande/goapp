package goapp

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
)

func ExampleLifecycle_BasicUsage() {
	l := NewLifecycle()

	l.Run(func(app *App, state *GoroutineState) error {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello world")
		})

		http.ListenAndServe(":8080", nil)

		return nil
	})

	os.Exit(l.Go(NewApp()))
}

func Test_LifeCycle_With_Success(t *testing.T) {
	l := NewLifecycle()

	l.Run(func(app *App, state *GoroutineState) error {

		return nil
	})

	r := l.Go(NewApp())

	assert.Equal(t, r, 0) // an error is present due to panic
}

func Test_LifeCycle_With_Panic(t *testing.T) {
	l := NewLifecycle()

	l.Run(func(app *App, state *GoroutineState) error {
		panic("take this !!")
	})

	r := l.Go(NewApp())

	assert.Equal(t, r, 1) // an error is present due to panic
}

func Test_LifeCycle_With_Error(t *testing.T) {
	l := NewLifecycle()

	l.Run(func(app *App, state *GoroutineState) error {
		return nil
	})

	l.Run(func(app *App, state *GoroutineState) error {
		return errors.New("this is a error")
	})

	l.Run(func(app *App, state *GoroutineState) error {
		return errors.New("this is a second error")
	})

	r := l.Go(NewApp())

	assert.Equal(t, r, 1) // an error is present
}

func Test_LifeCycle_Wait_Run(t *testing.T) {
	l := NewLifecycle()

	l.Init(func(app *App) error {
		app.Set("hello", func(app *App) interface{} {
			return "world"
		})

		return nil
	})

	l.Register(func(app *App) error {
		// The register need to value from the initial loop
		hello := app.GetString("hello") // so the hello service will be initialized

		assert.Equal(t, hello, "world")

		app.Set("testing", func(app *App) interface{} {
			return t
		})

		return nil
	})

	l.Config(func(app *App) error {
		tester := app.Get("testing").(*testing.T) // always need to cast

		assert.Equal(t, tester, t)

		app.Set("messages", func(app *App) interface{} {
			return list.New()
		})

		return nil
	})

	l.Run(func(app *App, state *GoroutineState) error {
		messages := app.Get("messages").(*list.List)

		messages.PushBack("Run1: sleep for 0.5s")
		time.Sleep(500 * time.Millisecond)
		messages.PushBack("Run1: Wake up ...")

		return nil
	})

	l.Run(func(app *App, state *GoroutineState) error {
		messages := app.Get("messages").(*list.List)

		messages.PushBack("Run2: sleep for 0.5s")
		time.Sleep(500 * time.Millisecond)
		messages.PushBack("Run2: Wake up ...")

		return nil
	})

	app := NewApp()

	r := l.Go(app)

	messages := app.Get("messages").(*list.List)
	assert.Equal(t, 4, messages.Len())

	assert.Equal(t, r, 0) // an error is not present
}

func Test_LifeCycle_Stop_Channel(t *testing.T) {

	l := NewLifecycle()

	l.Config(func(app *App) error {
		app.Set("messages", func(app *App) interface{} {
			return list.New()
		})

		return nil
	})

	getExit := false

	f := func(app *App, state *GoroutineState) error {
		for {
			select {
			case <-state.In:
				getExit = true
				return nil

			default:
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	l.Run(f)
	l.Run(f)

	l.Run(func(app *App, state *GoroutineState) error {
		time.Sleep(1000 * time.Millisecond)

		state.Out <- 1

		return nil
	})

	l.Go(NewApp())

	assert.True(t, getExit) // an error is present
}
