package supervisor

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ReadResult reads an event result and returns the payload.
func ReadResult(reader io.Reader) (payload []byte, err error) {
	buf := bufio.NewReader(reader)
	header, err := buf.ReadBytes('\n')
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
	_, err = buf.Read(payload)
	return
}

// WriteResult writes an event result to the stream and returns the number of
// byte written and an error if one occurs.
func WriteResult(writer io.Writer, result []byte) (n int, err error) {
	n, err = fmt.Fprintf(writer, "RESULT %v\n", len(result))
	if err != nil {
		return
	}

	n2, err := writer.Write(result)
	n += n2
	return
}
