package foreign

import (
	"bufio"
	"io"
	"os"
	"slug/internal/object"
	"strings"
	"sync"
)

var (
	stdinReadLinesOnce    sync.Once
	stdinReadLinesChannel *object.Channel
)

func fnIoStdinReadLines() *object.Foreign {
	return &object.Foreign{
		Name: "readLines",
		Fn: func(ctx object.EvaluatorContext, args ...object.Object) object.Object {
			if len(args) != 0 {
				return ctx.NewError("wrong number of arguments to `readLines`, got=%d, want=0", len(args))
			}

			stdinReadLinesOnce.Do(func() {
				stdinReadLinesChannel = object.NewChannel(0)
				go readStdinLines(stdinReadLinesChannel)
			})

			return stdinReadLinesChannel
		},
	}
}

func readStdinLines(ch *object.Channel) {
	defer ch.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err == nil {
			ch.GoChan() <- &object.String{Value: normalizeLine(line)}
			continue
		}

		if err == io.EOF {
			if len(line) > 0 {
				ch.GoChan() <- &object.String{Value: normalizeLine(line)}
			}
			return
		}

		// Unexpected stdin read errors terminate the stream.
		return
	}
}

func normalizeLine(line string) string {
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line
}
