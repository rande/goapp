package goapp

import (
	"container/list"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_LifeCycle_With_Success(t *testing.T) {
	l := NewLifecycle()

	l.Run(func(app *App) error {

		return nil
	})

	r := l.Go(NewApp())

	assert.Equal(t, r, 0) // an error is present due to panic
}

func Test_LifeCycle_With_Panic(t *testing.T) {
	l := NewLifecycle()

	l.Run(func(app *App) error {
		panic("take this !!")
	})

	r := l.Go(NewApp())

	assert.Equal(t, r, 1) // an error is present due to panic
}

func Test_LifeCycle_With_Error(t *testing.T) {
	l := NewLifecycle()

	l.Run(func(app *App) error {
		return nil
	})

	l.Run(func(app *App) error {
		return errors.New("this is a error")
	})

	l.Run(func(app *App) error {
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

	l.Run(func(app *App) error {
		messages := app.Get("messages").(*list.List)

		messages.PushBack("Run1: sleep for 0.5s")
		time.Sleep(500 * time.Millisecond)
		messages.PushBack("Run1: Wake up ...")

		return nil
	})

	l.Run(func(app *App) error {
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

	assert.Equal(t, r, 0) // an error is present
}
