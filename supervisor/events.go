package supervisor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Event represents a Supervisor event. An event consists of a header, metadata (meta) and an optional payload.
type Event struct {
	Header  map[string]string
	Meta    map[string]string
	Payload []byte
}

// getHeaderString returns the requested header value as a string or an empty string if missing.
func (event Event) getHeaderString(key string) string {
	if value, ok := event.Header[key]; ok {
		return value
	} else {
		return ""
	}
}

// getHeaderInt returns the requested header value as an int or 0 if missing.
func (event Event) getHeaderInt(key string) int {
	strval, ok := event.Header[key]
	if !ok {
		return 0
	}

	intval, err := strconv.Atoi(strval)
	if err != nil {
		return 0
	}

	return intval
}

// Name returns the name of the event.
func (event Event) Name() string {
	return event.getHeaderString("eventname")
}

// Serial returns the event serial number.
func (event Event) Serial() int {
	return event.getHeaderInt("serial")
}

// Pool returns the event pool where the event originated.
func (event Event) Pool() string {
	return event.getHeaderString("pool")
}

// PoolSerial returns the serial of the event in the event pool where the event originated.
func (event Event) PoolSerial() int {
	return event.getHeaderInt("poolserial")
}

// Version returns the version of the Supervisor instance that sent the event.
func (event Event) Version() string {
	return event.getHeaderString("ver")
}

// ToBytes converts the event to a byte array suitable for parsing.
func (event Event) ToBytes() []byte {
	mapser := func(data map[string]string) []byte {
		parts := make([]string, 0, len(data))
		for k, v := range data {
			parts = append(parts, fmt.Sprintf("%v:%v", k, v))
		}
		return []byte(strings.Join(parts, " "))
	}

	meta := mapser(event.Meta)
	payload := make([]byte, 0, len(meta)+len(event.Payload)+1)
	payload = append(payload, meta...)
	if event.Payload == nil || len(event.Payload) > 0 {
		payload = append(payload, '\n')
		payload = append(payload, event.Payload...)
	}

	event.Header["len"] = strconv.Itoa(len(payload))
	header := mapser(event.Header)
	message := make([]byte, 0, len(header)+len(payload))
	message = append(message, header...)
	message = append(message, '\n')
	message = append(message, payload...)
	return message
}

// parseMap parses a header or metadata string into a string/string map.
func parseMap(data []byte) map[string]string {
	str := strings.TrimSpace(string(data))
	mapping := make(map[string]string)
	for _, token := range strings.Split(str, " ") {
		if token == "" {
			continue
		}
		pair := strings.SplitN(token, ":", 2)
		switch len(pair) {
		case 2:
			mapping[pair[0]] = pair[1]
		case 1:
			mapping[pair[0]] = ""
		}
	}
	return mapping
}

// parsePayload parses a raw event payload into a metadata and payload.
func parsePayload(data []byte) (meta map[string]string, payload []byte) {
	if index := bytes.IndexByte(data, '\n'); index > 0 {
		meta = parseMap(data[:index])
		payload = data[index+1:]
	} else {
		meta = parseMap(data)
	}
	return
}

// ReadEvent waits for a Supervisor event and returns the parsed event. The err
// will be non-nil if an error occurred. This may include io.EOF which should
// preceed closing of the reader.
func ReadEvent(reader *bufio.Reader) (event *Event, err error) {
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}

	header := parseMap(data)
	length, err := strconv.Atoi(header["len"])
	if err != nil {
		return
	}

	rawPayload := make([]byte, length)
	_, err = reader.Read(rawPayload)
	if err != nil {
		return
	}

	event = new(Event)
	event.Header = header
	event.Meta, event.Payload = parsePayload(rawPayload)
	return
}

// WriteResult writes an event result to the stream and returns the number of
// byte written and an error if one occurs.
func WriteResult(writer *bufio.Writer, result []byte) (n int, err error) {
	n, err = fmt.Fprintf(writer, "RESULT %v\n", len(result))
	if err != nil {
		return
	}

	n2, err := writer.Write(result)
	n += n2
	writer.Flush()
	return
}

// WriteResultOK is a shortcut to write an OK result to the stream.
func WriteResultOK(writer *bufio.Writer) (n int, err error) {
	return WriteResult(writer, []byte("OK"))
}

// WriteResultFail is a shortcut to write a Fail result to the stream.
func WriteResultFail(writer *bufio.Writer) (n int, err error) {
	return WriteResult(writer, []byte("FAIL"))
}

// ReadResult reads an event result and returns the payload.
func ReadResult(reader *bufio.Reader) (payload []byte, err error) {
	header, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}

	tokens := strings.SplitN(string(header), " ", 2)
	if len(tokens) != 2 || tokens[0] != "RESULT" {
		err = errors.New(fmt.Sprintf("result header invalid: %s", header))
		return
	}

	length, err := strconv.Atoi(strings.TrimRight(tokens[1], "\n"))
	if err != nil {
		return
	}

	payload = make([]byte, length)
	_, err = reader.Read(payload)
	return
}

// Listen is a simple Supervisor event listener that sends received events over
// the provided channel. It responds to Supervisor with an OK after queuing an
// event in the channel. It returns an error if one occurs or nil if the reader
// encounters an EOF.
func Listen(in *bufio.Reader, out *bufio.Writer, ch chan *Event) error {
	var event *Event
	var err error

	for {
		event, err = ReadEvent(in)
		if err != nil {
			break
		}
		ch <- event
		WriteResultOK(out)
	}

	if err == io.EOF {
		err = nil
	}
	return err
}
