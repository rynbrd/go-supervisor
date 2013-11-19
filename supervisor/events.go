package supervisor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	parentEvents []string = []string{
		"PROCESS_COMMUNICATION",
		"PROCESS_LOG",
		"PROCESS_STATE",
		"SUPERVISOR_STATE_CHANGE",
		"TICK",
	}
)

// mapInt retrieves a value from a map as an int.
func mapInt(data map[string]string, key string) int {
	if strval, ok := data[key]; ok {
		if intval, err := strconv.Atoi(strval); err == nil {
			return intval
		}
	}
	return 0
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

// Event represents a Supervisor event. An event consists of a header, metadata (meta) and an optional payload.
type Event struct {
	Header  map[string]string
	Meta    map[string]string
	Payload []byte
}

// ReadEvent waits for a Supervisor event and returns the parsed event. The err
// will be non-nil if an error occurred. This may include io.EOF which should
// preceed closing of the reader.
func ReadEvent(reader io.Reader) (event Event, err error) {
	buf := bufio.NewReader(reader)
	data, err := buf.ReadBytes('\n')
	if err != nil {
		return
	}

	header := parseMap(data)
	length, err := strconv.Atoi(header["len"])
	if err != nil {
		return
	}

	rawPayload := make([]byte, length)
	_, err = buf.Read(rawPayload)
	if err != nil {
		return
	}

	event = Event{}
	event.Header = header
	event.Meta, event.Payload = parsePayload(rawPayload)
	return
}

// String returns the event as a human readable string.
func (event Event) String() string {
	return fmt.Sprintf("Event{%d:%s}", event.Serial(), event.Name())
}

// HeaderInt returns the requested header value as an int or 0 if missing or broken.
func (event Event) HeaderInt(key string) int {
	return mapInt(event.Header, key)
}

// MetaInt returns the requested meta value as an int or 0 if missing or broken.
func (event Event) MetaInt(key string) int {
	return mapInt(event.Meta, key)
}

// Name returns the name of the event.
func (event Event) Name() string {
	return event.Header["eventname"]
}

// Parent returns the parent type of the event.
func (event Event) Parent() string {
	name := event.Name()
	for _, parent := range parentEvents {
		prefix := parent + "_"
		if len(prefix) <= len(name) && prefix == name[:len(prefix)] {
			return parent
		}
	}
	return name
}

// State determines the state of the process or supervisor instance based on the name.
func (event Event) State() string {
	parent := event.Parent()
	switch parent {
	case "PROCESS_STATE":
		fallthrough
	case "SUPERVISOR_STATE_CHANGE":
		return event.Name()[len(parent)+1:]
	}
	return ""
}

// Serial returns the event serial number.
func (event Event) Serial() int {
	return event.HeaderInt("serial")
}

// Pool returns the event pool where the event originated.
func (event Event) Pool() string {
	return event.Header["pool"]
}

// PoolSerial returns the serial of the event in the event pool where the event originated.
func (event Event) PoolSerial() int {
	return event.HeaderInt("poolserial")
}

// Version returns the version of the Supervisor instance that sent the event.
func (event Event) Version() string {
	return event.Header["ver"]
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
