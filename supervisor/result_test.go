package supervisor

import (
	"io"
	"testing"
)

// Test the ReadResult and WriteResult functions.
func TestResult(t *testing.T) {
	reader, writer := io.Pipe()

	readAndVerify := func(expected string) {
		payload, err := ReadResult(reader)
		switch {
		case err != nil:
			t.Errorf(`ReadResult() => error{"%v"}, want payload "%s"`, err, expected)
		case string(payload) != expected:
			t.Errorf(`ReadResult() => "%s", want "%s"`, payload, expected)
		}
	}

	payload := "some arbitrary data"
	go WriteResult(writer, []byte(payload))
	readAndVerify(payload)
}
