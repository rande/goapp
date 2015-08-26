package goapp

const (
	Control_End   = 0 // process is stopped
	Control_Error = 1 // process generated an error
	Control_Stop  = 2 // process required to stop the application
	Control_Kill  = 3 // process required to kill the application
)

type GoroutineState struct {
	In    chan int // allow a goroutine to receive a stop/exit signal
	Out   chan int // allow a goroutine to send an exit command
	Error error    // store any error available while running the Process
}

func (s *GoroutineState) Close() {
	//	close(s.In)
	//	close(s.Out)
}

func NewGoroutineState() *GoroutineState {
	return &GoroutineState{
		In:    make(chan int),
		Out:   make(chan int),
		Error: nil,
	}
}
