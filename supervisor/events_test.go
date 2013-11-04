package supervisor

import (
	"strings"
	"bufio"
	"io"
	"strconv"
	"testing"
)

func TestRead(t *testing.T) {
	reader, writer := io.Pipe()
	bufReader := bufio.NewReader(reader)
	serial := 0
	header := map[string]string{
		"ver":    "3.0",
		"server": "supervisor",
		"pool":   "listener",
	}

	send := func(eventname string, meta map[string]string, payload []byte) {
		header["serial"] = strconv.Itoa(serial)
		header["poolserial"] = header["serial"]
		header["eventname"] = eventname
		event := new(Event)
		event.Header = header
		event.Meta = meta
		event.Payload = payload
		writer.Write(event.ToBytes())
		serial += 1
	}

	send_and_verify := func(eventname string, meta map[string]string, payload []byte) {
		go send("TEST", meta, payload)
		event, err := ReadEvent(bufReader)
		if err != nil {
			t.Error(err)
		}

		cmpMap := func(m1 map[string]string, m2 map[string]string) bool {
			for k, v := range m2 {
				if v != m1[k] {
					return false
				}
			}
			return true
		}

		cmpPayload := func(p1 []byte, p2 []byte) bool {
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

		switch {
		case !cmpMap(event.Header, header):
			t.Error("invalid event header")
		case !cmpMap(event.Meta, meta):
			t.Error("invalid event meta")
		case !cmpPayload(event.Payload, payload):
			t.Errorf("invalid event payload")
		}
	}

	meta := map[string]string{"processname": "test", "moredata": "strictlyextra"}
	send_and_verify("EVENT_EMPTY_PAYLOAD", meta, []byte{})
	send_and_verify("EVENT_FULL_PAYLOAD", meta, []byte("this is a payload test"))
}

func TestWrite(t *testing.T) {
	reader, writer := io.Pipe()
	bufReader := bufio.NewReader(reader)
	bufWriter := bufio.NewWriter(writer)

	read_and_verify := func(expected string) {
		data, err := bufReader.ReadBytes('\n')
		if err != nil {
			t.Error(err)
		}

		header := string(data)
		tokens := strings.SplitN(header, " ", 2)
		if len(tokens) != 2 || tokens[0] != "RESULT" {
			t.Errorf("Result header invalid: %s", header)
		}

		length, err := strconv.Atoi(tokens[1])
		if err != nil {
			t.Errorf("Result length invalid: %s", tokens[2])
		}

		data = make([]byte, length)
		_, err = bufReader.Read(data)
		if err != nil {
			t.Error(err)
		}

		payload := string(data)
		if payload != expected {
			t.Errorf("Payload result invalid: %s != %s", payload, expected)
		}
	}

	payload := "some arbitrary data"
	WriteResult(bufWriter, []byte(payload))
	go read_and_verify(payload)

	WriteResultOK(bufWriter)
	go read_and_verify("OK")

	WriteResultFail(bufWriter)
	go read_and_verify("FAIL")
}
