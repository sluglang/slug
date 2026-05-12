package object

import (
	"bytes"
	"fmt"
	"slug/internal/util"
	"strings"
)

func RenderError(err *Error) string {
	if err == nil {
		return ""
	}
	kind := strings.TrimSpace(err.Kind)
	if kind == "" {
		kind = "Error"
	}
	if err.Position > 0 && strings.TrimSpace(err.Src) != "" {
		line, col := util.GetLineAndColumn(err.Src, err.Position)
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%s: %s\n\n", kind, err.Message)
		if strings.TrimSpace(err.Path) != "" {
			fmt.Fprintf(&buf, "    --> %s:%d:%d\n", err.Path, line, col)
		} else {
			fmt.Fprintf(&buf, "    --> %d:%d\n", line, col)
		}
		buf.WriteString(util.GetContextLines(err.Src, line, col))
		return buf.String()
	}
	return fmt.Sprintf("%s: %s", kind, err.Message)
}

func RenderStacktrace(rtErr *RuntimeError) string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "RuntimeError: %s\n\n", rtErr.Payload.Inspect())

	if len(rtErr.StackTrace) > 0 {
		l, c := util.GetLineAndColumn(rtErr.StackTrace[0].Src, rtErr.StackTrace[0].Position)
		if strings.TrimSpace(rtErr.StackTrace[0].File) != "" {
			fmt.Fprintf(&buf, "    --> %s:%d:%d\n", rtErr.StackTrace[0].File, l, c)
		} else {
			fmt.Fprintf(&buf, "    --> %d:%d\n", l, c)
		}
		buf.WriteString(util.GetContextLines(rtErr.StackTrace[0].Src, l, c))
		buf.WriteString("\n")
	}

	// Start with the payload itself
	fmt.Fprintf(&buf, "Stacktrace: %s", rtErr.Payload.Inspect())
	buf.WriteString(formatRuntimeErrorStack(rtErr))

	return buf.String()
}

// Helper: turn a RuntimeError's stack trace into a human-readable string.
func formatRuntimeErrorStack(rtErr *RuntimeError) string {
	var buf bytes.Buffer

	for _, frame := range rtErr.StackTrace {
		l, c := util.GetLineAndColumn(frame.Src, frame.Position)
		fmt.Fprintf(&buf, "\n  at [%3d:%3d] %s - %s", l, c, frame.File, frame.Function)
	}

	// Optionally include chained causes
	if rtErr.Cause != nil {
		fmt.Fprintf(&buf, "\n\nCaused by: %s", rtErr.Cause.Payload.Inspect())
		buf.WriteString(formatRuntimeErrorStack(rtErr.Cause))
	}

	return buf.String()
}
