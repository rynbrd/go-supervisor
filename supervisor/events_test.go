package supervisor

import (
	"bufio"
	"io"
	"strconv"
	"testing"
)

// Compare two string/string maps.
func cmpMap(m1 map[string]string, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m2 {
		if v != m1[k] {
			return false
		}
	}
	return true
}

// Compare two byte arrays.
func cmpBytes(p1 []byte, p2 []byte) bool {
	if p1 == nil {
		p1 = []byte{}
	}
	if p2 == nil {
		p2 = []byte{}
	}

	if len(p1) != len(p2) {
		return false
	}
	for i, v := range p2 {
		if v != p1[i] {
			return false
		}
	}
	return true
}

// Compare two events
func cmpEvents(e1 *Event, e2 *Event) bool {
	switch {
	case e1 == nil && e2 == nil:
		return true
	case e1 == nil || e2 == nil:
		return false
	default:
		return cmpMap(e1.Header, e2.Header) && cmpMap(e1.Meta, e2.Meta) && cmpBytes(e1.Payload, e2.Payload)
	}
}

// Construct an event.
func createEvent(serial int, eventname string, processname string, payload []byte) Event {
	serialstr := strconv.Itoa(serial)
	return Event{
		map[string]string{
			"ver":        "3.0",
			"server":     "supervisor",
			"eventname":  eventname,
			"serial":     serialstr,
			"pool":       "listener",
			"poolserial": serialstr,
		},
		map[string]string{
			"processname": processname,
			"groupname":   processname,
		},
		payload,
	}
}

// Test the ReadEvent function.
func TestReadEvent(t *testing.T) {
	reader, writer := io.Pipe()
	bufReader := bufio.NewReader(reader)
	serial := 0

	sendAndVerify := func(eventname string, payload []byte) {
		sentEvent := createEvent(serial, eventname, "test", payload)
		serial++

		go func() {
			_, err := writer.Write(sentEvent.ToBytes())
			if err != nil {
				t.Error(err)
			}
		}()

		receiveEvent, err := ReadEvent(bufReader)
		if err != nil {
			t.Error(err)
		}

		if !cmpEvents(&sentEvent, &receiveEvent) {
			t.Error("invalid event received")
		}
	}

	sendAndVerify("EVENT_EMPTY_PAYLOAD", []byte{})
	sendAndVerify("EVENT_FULL_PAYLOAD", []byte("this is a payload test"))
}
