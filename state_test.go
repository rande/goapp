package goapp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

// Just small tests to check goroutine behaviors

func Test_Signal_BasicUsage(t *testing.T) {

	s1 := NewGoroutineState()
	s2 := NewGoroutineState()
	s3 := NewGoroutineState()

	p := []*GoroutineState{s1, s2, s3}

	go func(s *GoroutineState) {
		// do some long stuff
		stop := "."

		for {
			select {
			case <-s.In:
				fmt.Println("Receiving a exist request")
				return
			default:
				fmt.Print(stop)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}(s1)

	go func(s *GoroutineState) {
		// do some long stuff
		stop := "."

		for {
			select {
			case <-s.In:
				fmt.Println("Receiving a exist request")
				return
			default:
				fmt.Print(stop)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}(s3)

	go func(s *GoroutineState) {
		time.Sleep(100 * time.Millisecond)

		s.Out <- 1 // ask to stop

	}(s2)

	go func(s *GoroutineState) {
		time.Sleep(100 * time.Millisecond)

		s.Out <- 2 // ask to stop
	}(s2)

	// wait for an exit signal. Do we need a RunFunc with signal handling and dispatching ???
	cases := make([]reflect.SelectCase, len(p))
	for i, state := range p {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(state.Out),
		}
	}

	chosen, _, _ := reflect.Select(cases)

	for pos, state := range p {
		if pos == chosen {
			continue
		}

		state.In <- 1
	}

	//	fmt.Print(<-s1.Out)

	assert.Equal(t, "world", "world")
}
