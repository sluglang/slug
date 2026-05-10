package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func frame(msg string) string {
	return "Content-Length: " + itoa(len(msg)) + "\r\n\r\n" + msg
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func readAllMessages(t *testing.T, out string) []map[string]interface{} {
	t.Helper()
	r := bufio.NewReader(strings.NewReader(out))
	msgs := []map[string]interface{}{}
	for {
		body, err := readFramedMessage(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		var m map[string]interface{}
		if err := json.Unmarshal(body, &m); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		msgs = append(msgs, m)
	}
	return msgs
}

func TestServerInitializeShutdownExit(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	if len(msgs) < 2 {
		t.Fatalf("expected responses, got %d", len(msgs))
	}
}

func TestServerPublishesDiagnosticsOnDidOpenAndChange(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"bad"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///a.slug","version":2},"contentChanges":[{"text":"ok"}]}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) {
		if strings.Contains(src, "bad") {
			return []string{"ParseError: boom\n    --> file:///a.slug:2:3\n"}, nil
		}
		return nil, nil
	})
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	foundDiag := 0
	for _, m := range msgs {
		if method, _ := m["method"].(string); method == "textDocument/publishDiagnostics" {
			foundDiag++
		}
	}
	if foundDiag < 2 {
		t.Fatalf("expected at least 2 diagnostics notifications, got %d", foundDiag)
	}
}

func TestParseDiagnosticParsesLocation(t *testing.T) {
	d := parseDiagnostic("ParseError: bad\n    --> /tmp/x.slug:10:5\n", 1, "x")
	if d.Range.Start.Line != 9 || d.Range.Start.Character != 4 {
		t.Fatalf("unexpected range: %+v", d.Range)
	}
}
