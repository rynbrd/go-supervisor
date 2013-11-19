package supervisor

import (
	"fmt"
	"io"
)

type Listener struct {
	in  io.Reader
	out io.Writer
}

// NewListener creates a new event listener with the given in and out streams. The listener will
// never close the streams.
func NewListener(in io.Reader, out io.Writer) Listener {
	return Listener{in, out}
}

// Read waits for and returns an event from supervisor. An error is returned if the read fails. If
// EOF is encountered the error will be io.EOF.
func (l Listener) Read() (event Event, err error) {
	return ReadEvent(l.in)
}

// Ready puts the listener into the READY state.
func (l Listener) Ready() error {
	_, err := fmt.Fprintf(l.out, "READY\n")
	return err
}

// Ack puts the listener into the ACKNOWLEDGED state.
func (l Listener) Ack() error {
	_, err := fmt.Fprintf(l.out, "ACKNOWLEDGED\n")
	return err
}

// Busy puts the listener into the BUSY state.
func (l Listener) Busy() error {
	_, err := fmt.Fprintf(l.out, "BUSY\n")
	return err
}

// Result sends a result payload to Supervisor.
func (l Listener) Result(result []byte) error {
	_, err := WriteResult(l.out, result)
	return err
}

// OK sends an OK result to Supervisor.
func (l Listener) Ok() error {
	return l.Result([]byte("OK"))
}

// Fail saends a FAIL result to Supervisor.
func (l Listener) Fail() error {
	return l.Result([]byte("FAIL"))
}

// Run starts the listener and sends recieved events over the provided channel. It will listen for
// events until EOF is recieved. This is a simple implementation that will send an OK result
// followed by a READY after every event is recieved and parsed. If parsing or reading fails for
// any reason then Run will exit with an error.
func (l Listener) Run(events chan Event) error {
	var event Event
	var err error

	l.Ready()
	for {
		event, err = l.Read()
		if err != nil {
			break
		}
		events <- event
		l.Ok()
		l.Ready()
	}

	if err == io.EOF {
		err = nil
	}
	return err
}
