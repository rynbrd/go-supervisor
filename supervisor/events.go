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

type Event struct {
	Name string
	Meta map[string]string
	Payload []byte
}

func parseMap(data []byte) map[string]string {
	str := strings.TrimSpace(string(data))
	mapping := make(map[string]string)
	for _, token := range strings.Split(str, " ") {
		if token == "" {
			continue
		}
		parts := strings.Split(token, ":")
		mapping[parts[0]] = parts[1]
	}
	return mapping
}

func parseHeader(data []byte) (name string, length int) {
	header := parseMap(data)
	name = header["eventname"]
	length, err := strconv.Atoi(header["len"])
	if err != nil {
		panic(err)
	}
	return
}

func parsePayload(data []byte) (meta map[string]string, payload []byte) {
	if index := bytes.IndexByte(data, '\n'); index > 0 {
		meta = parseMap(data[:index])
		payload = data[index:]
	} else {
		meta = parseMap(data)
	}
	return
}

func readEvent(reader *bufio.Reader) *Event {
	data, err := reader.ReadBytes('\n')
	if err != nil {
		panic(err)
	}

	name, length := parseHeader(data)
	rawPayload := make([]byte, length)
	n, err := reader.Read(rawPayload)
	if err != nil {
		panic(err)
	}
	if n < length {
		panic(errors.New("failed to read full event payload"))
	}

	meta, payload := parsePayload(rawPayload)
	return &Event{name, meta, payload}
}

func Listen(reader *bufio.Reader, ch chan *Event) (err error) {
	defer func() {
		close(ch)
		if r := recover(); r != nil && r != io.EOF {
			err = r.(error)
		}
	}()

	fmt.Print("READY\n")
	for {
		ch <- readEvent(reader)
		fmt.Print("ACKNOWLEDGED\n")
	}
}
