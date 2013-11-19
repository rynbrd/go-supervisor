package supervisor

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

// Test the Listen function.
func TestListen(t *testing.T) {
	stdin, stdinWriter := io.Pipe()
	stdoutReader, stdout := io.Pipe()

	ch := make(chan Event, 1)
	reader := bufio.NewReader(stdoutReader)
	listener := NewListener(stdin, stdout)

	go func() {
		if err := listener.Run(ch); err != nil {
			t.Errorf(`Listen() => error{"%v"}, want nil`, err)
		}
	}()

	serial := 0

	readAndVerifyState := func(state string) {
		realState := strings.ToUpper(state)
		data := make([]byte, len(realState)+1)
		if _, err := stdoutReader.Read(data); err != nil {
			t.Errorf(`Listener.%s() => error{"%v"}`, state, err)
		} else if string(data) != realState+"\n" {
			t.Errorf(`Listener.%s() => "%s", want "%s"`, state, data, realState)
		}
	}

	sendAndVerifyEvent := func(eventname string, payload []byte) {
		sentEvent := createEvent(serial, eventname, "test", payload)
		serial++

		bytes := sentEvent.ToBytes()
		if _, err := stdinWriter.Write(bytes); err != nil {
			t.Errorf(`stdin.Write() => error{"%v"}, want n=%d`, err, len(bytes))
		}

		if result, err := ReadResult(reader); err != nil {
			t.Errorf(`ReadResult() => error{"%v"}, want result="OK"`, err)
		} else if string(result) != "OK" {
			t.Errorf(`ReadResult() => "%s", want "OK"`, result)
		}

		readAndVerifyState("Ready")

		receiveEvent, ok := <-ch
		if !ok {
			t.Errorf(`(event, ok := <-ch) => channel closed, want event`)
		} else if !cmpEvents(&sentEvent, &receiveEvent) {
			t.Errorf(`(event, ok := <-ch) => got %s, want %s`, receiveEvent, sentEvent)
		}
	}

	readAndVerifyState("Ready")
	sendAndVerifyEvent("PROCESS_STATE_RUNNING", []byte{})
	sendAndVerifyEvent("PROCESS_LOG_STDERR", []byte("some pretend log data"))
}
