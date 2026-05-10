package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 1 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("initialize result missing: %#v", m)
		}
		caps, ok := res["capabilities"].(map[string]interface{})
		if !ok {
			t.Fatalf("initialize capabilities missing: %#v", res)
		}
		if _, ok := caps["completionProvider"].(map[string]interface{}); !ok {
			t.Fatalf("completionProvider missing in capabilities: %#v", caps)
		}
		if dh, ok := caps["documentHighlightProvider"].(bool); !ok || !dh {
			t.Fatalf("expected documentHighlightProvider=true, got %#v", caps["documentHighlightProvider"])
		}
		if rp, ok := caps["referencesProvider"].(bool); !ok || !rp {
			t.Fatalf("expected referencesProvider=true, got %#v", caps["referencesProvider"])
		}
		rn, ok := caps["renameProvider"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected renameProvider object, got %#v", caps["renameProvider"])
		}
		if pp, _ := rn["prepareProvider"].(bool); !pp {
			t.Fatalf("expected renameProvider.prepareProvider=true, got %#v", rn["prepareProvider"])
		}
		cp, _ := caps["completionProvider"].(map[string]interface{})
		if rp, _ := cp["resolveProvider"].(bool); !rp {
			t.Fatalf("expected resolveProvider=true, got %#v", cp["resolveProvider"])
		}
		if capVal, ok := caps["codeActionProvider"].(bool); !ok || !capVal {
			t.Fatalf("expected codeActionProvider=true, got %#v", caps["codeActionProvider"])
		}
		if shp, ok := caps["signatureHelpProvider"].(map[string]interface{}); !ok || shp == nil {
			t.Fatalf("expected signatureHelpProvider, got %#v", caps["signatureHelpProvider"])
		}
		return
	}
	t.Fatal("initialize response not found")
}

func TestServerExitBeforeShutdownReturnsError(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err == nil {
		t.Fatal("expected exit-before-shutdown error")
	}
}

func TestServerUnknownRequestReturnsMethodNotFound(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","id":9,"method":"does/not/exist","params":{}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	found := false
	for _, m := range msgs {
		errObj, ok := m["error"].(map[string]interface{})
		if !ok {
			continue
		}
		if code, ok := errObj["code"].(float64); ok && int(code) == -32601 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected method-not-found response")
	}
}

func TestServerCanceledUnknownRequestSuppressesErrorResponse(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":9}}`) +
			frame(`{"jsonrpc":"2.0","id":9,"method":"does/not/exist","params":{}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 9 {
			continue
		}
		t.Fatalf("expected canceled unknown request id=9 to be suppressed, got %#v", m)
	}
}

func TestServerCanceledUnknownRequestStringIDSuppressesErrorResponse(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":"req-unknown"}}`) +
			frame(`{"jsonrpc":"2.0","id":"req-unknown","method":"does/not/exist","params":{}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(string)
		if !ok {
			continue
		}
		if id == "req-unknown" {
			t.Fatalf("expected canceled unknown request id=req-unknown to be suppressed, got %#v", m)
		}
	}
}

func TestServerRejectsRequestsAfterShutdown(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","id":3,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	found := false
	for _, m := range msgs {
		errObj, ok := m["error"].(map[string]interface{})
		if !ok {
			continue
		}
		if code, ok := errObj["code"].(float64); ok && int(code) == -32600 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected invalid-request response after shutdown")
	}
}

func TestServerDidChangeBeforeOpenIgnored(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///a.slug","version":2},"contentChanges":[{"text":"ok"}]}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		if method, _ := m["method"].(string); method == "textDocument/publishDiagnostics" {
			t.Fatal("didChange-before-open should not publish diagnostics")
		}
	}
}

func TestServerDebouncesRapidDidChangeBursts(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"a"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///a.slug","version":2},"contentChanges":[{"text":"b"}]}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///a.slug","version":3},"contentChanges":[{"text":"c"}]}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) {
		return []string{"ParseError: e\n    --> /tmp/a.slug:1:1\n"}, nil
	})
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	count := 0
	for _, m := range msgs {
		if method, _ := m["method"].(string); method == "textDocument/publishDiagnostics" {
			count++
		}
	}
	if count > 4 {
		t.Fatalf("expected coalesced diagnostics notifications, got %d", count)
	}
}

func TestServerPublishesDiagnosticsOnDidOpenAndChange(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"bad"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///a.slug","version":2},"contentChanges":[{"text":"ok"}]}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
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

func TestServerDedupesDuplicateDiagnostics(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"bad"}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) {
		msg := "ParseError: boom\n    --> /tmp/a.slug:2:3\n"
		return []string{msg, msg}, nil
	})
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		if method, _ := m["method"].(string); method != "textDocument/publishDiagnostics" {
			continue
		}
		params, ok := m["params"].(map[string]interface{})
		if !ok {
			continue
		}
		diags, ok := params["diagnostics"].([]interface{})
		if !ok {
			continue
		}
		if len(diags) != 1 {
			t.Fatalf("expected deduped diagnostics length 1, got %d", len(diags))
		}
		return
	}
	t.Fatal("expected diagnostics notification")
}

func TestServerCancelRequestNotificationIsAccepted(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":99}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !s.canceledReqs["99"] {
		t.Fatal("expected canceled request id to be recorded")
	}
}

func TestServerFlushesDeferredDiagnosticsOnShutdown(t *testing.T) {
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"a"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///a.slug","version":2},"contentChanges":[{"text":"b"}]}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) {
		if src == "b" {
			return []string{"ParseError: changed\n    --> /tmp/a.slug:1:1\n"}, nil
		}
		return nil, nil
	})
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	found := false
	for _, m := range msgs {
		if method, _ := m["method"].(string); method != "textDocument/publishDiagnostics" {
			continue
		}
		params, ok := m["params"].(map[string]interface{})
		if !ok {
			continue
		}
		diags, ok := params["diagnostics"].([]interface{})
		if !ok {
			continue
		}
		if len(diags) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("expected diagnostics to be flushed on shutdown")
	}
}

func TestDebounceWindowConstantIsSane(t *testing.T) {
	if diagnosticsDebounceWindow <= 0 {
		t.Fatal("debounce window must be > 0")
	}
	if diagnosticsDebounceWindow > 500*time.Millisecond {
		t.Fatalf("debounce window unexpectedly high: %s", diagnosticsDebounceWindow)
	}
}

func TestServerHoverReturnsSymbolInfo(t *testing.T) {
	src := "val answer = 42\\nval use = answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":7,"method":"textDocument/hover","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":11}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 7 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("hover result missing: %#v", m)
		}
		contents, ok := res["contents"].(map[string]interface{})
		if !ok {
			t.Fatalf("hover contents missing: %#v", res)
		}
		v, _ := contents["value"].(string)
		if !strings.Contains(v, "answer") {
			t.Fatalf("unexpected hover value: %q", v)
		}
		return
	}
	t.Fatal("hover response not found")
}

func TestServerHoverDistinguishesValConstantAndVarVariable(t *testing.T) {
	src := "val c = 1\nvar v = 2\nc\nv\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":54,"method":"textDocument/hover","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":2,"character":0}}}`) +
			frame(`{"jsonrpc":"2.0","id":55,"method":"textDocument/hover","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":3,"character":0}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	seen54 := false
	seen55 := false
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok {
			continue
		}
		switch int(id) {
		case 54:
			seen54 = true
			res, _ := m["result"].(map[string]interface{})
			contents, _ := res["contents"].(map[string]interface{})
			v, _ := contents["value"].(string)
			if !strings.Contains(v, "`c` (constant)") {
				t.Fatalf("expected val hover to show constant, got %q", v)
			}
		case 55:
			seen55 = true
			res, _ := m["result"].(map[string]interface{})
			contents, _ := res["contents"].(map[string]interface{})
			v, _ := contents["value"].(string)
			if !strings.Contains(v, "`v` (variable)") {
				t.Fatalf("expected var hover to show variable, got %q", v)
			}
		}
	}
	if !seen54 || !seen55 {
		t.Fatalf("missing hover responses: seen54=%v seen55=%v", seen54, seen55)
	}
}

func TestServerHoverIncludesDocComment(t *testing.T) {
	src := "/**\n * Answer docs.\n */\nval answer = 42\nval use = answer\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":47,"method":"textDocument/hover","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":4,"character":11}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 47 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("hover result missing: %#v", m)
		}
		contents, ok := res["contents"].(map[string]interface{})
		if !ok {
			t.Fatalf("hover contents missing: %#v", res)
		}
		v, _ := contents["value"].(string)
		if !strings.Contains(v, "Answer docs.") {
			t.Fatalf("expected hover to include doc comment, got: %q", v)
		}
		return
	}
	t.Fatal("hover response not found")
}

func TestServerHoverIncludesImportedAliasDocComment(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "/**\n * Reduce docs from std hover.\n */\n@export val reduce = fn(vs, f, init){ init }\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	src := "val { reduce } = import(\"slug.std\")\nreduce([1,2], fn(a,b){a}, 0)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	playPath := filepath.Join(tmp, "playground.slug")
	if err := os.WriteFile(playPath, []byte(src), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":50,"method":"textDocument/hover","params":{"textDocument":{"uri":"`+playURI+`"},"position":{"line":1,"character":2}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 50 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("hover result missing: %#v", m)
		}
		contents, ok := res["contents"].(map[string]interface{})
		if !ok {
			t.Fatalf("hover contents missing: %#v", res)
		}
		v, _ := contents["value"].(string)
		if !strings.Contains(v, "`reduce` (function)") {
			t.Fatalf("expected hover kind=function for imported alias, got: %q", v)
		}
		if !strings.Contains(v, "Reduce docs from std hover.") {
			t.Fatalf("expected hover to include imported doc comment, got: %q", v)
		}
		return
	}
	t.Fatal("hover response not found")
}

func TestServerDocumentSymbolReturnsTopLevel(t *testing.T) {
	src := "val f = fn(x){ x }\\nval A = struct { name, }\\nvar n = 1\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":8,"method":"textDocument/documentSymbol","params":{"textDocument":{"uri":"file:///a.slug"}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 8 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("documentSymbol result missing: %#v", m)
		}
		if len(res) < 3 {
			t.Fatalf("expected >=3 symbols, got %d", len(res))
		}
		return
	}
	t.Fatal("documentSymbol response not found")
}

func TestServerDefinitionReturnsLocalDeclaration(t *testing.T) {
	src := "val answer = 42\\nval use = answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":9,"method":"textDocument/definition","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":11}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 9 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("definition result missing: %#v", m)
		}
		uri, _ := res["uri"].(string)
		if !strings.HasPrefix(uri, "file://") {
			t.Fatalf("unexpected definition uri: %q", uri)
		}
		return
	}
	t.Fatal("definition response not found")
}

func TestServerDefinitionReturnsNilOutsideIdentifier(t *testing.T) {
	src := "val alpha = 1\\nalpha\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":10,"method":"textDocument/definition","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":0,"character":3}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 10 {
			continue
		}
		if m["result"] != nil {
			t.Fatalf("expected nil definition result outside identifier, got %#v", m["result"])
		}
		return
	}
	t.Fatal("definition response not found")
}

func TestServerDefinitionReturnsNilInsideStringLiteral(t *testing.T) {
	src := "val alpha = 1\\n\"alpha\"\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":11,"method":"textDocument/definition","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":2}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 11 {
			continue
		}
		if m["result"] != nil {
			t.Fatalf("expected nil definition result inside string literal, got %#v", m["result"])
		}
		return
	}
	t.Fatal("definition response not found")
}

func TestServerDefinitionResolvesWildcardImportedExportFromModuleFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "@export val reduce = fn(a,b){ a }\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	playPath := filepath.Join(tmp, "playground.slug")
	playSrc := "var {*} = import(\\\"slug.std\\\")\\n[1,2] /> reduce(0, fn(a,b){a})\\n"
	if err := os.WriteFile(playPath, []byte(strings.ReplaceAll(playSrc, "\\\\", "")), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)

	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+playSrc+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":30,"method":"textDocument/definition","params":{"textDocument":{"uri":"`+playURI+`"},"position":{"line":1,"character":10}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 30 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("definition result missing: %#v", m)
		}
		uri, _ := res["uri"].(string)
		if !strings.HasSuffix(uri, "/lib/slug/std.slug") {
			t.Fatalf("expected definition uri in std module, got %q", uri)
		}
		return
	}
	t.Fatal("definition response not found")
}

func TestServerCodeActionSuggestsImportForUnresolvedSymbolFromSlugHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "@export val reduce = fn(a,b){ a }\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	playPath := filepath.Join(tmp, "playground.slug")
	playSrc := "[1,2] /> reduce(0, fn(a,b){a})\\n"
	if err := os.WriteFile(playPath, []byte(strings.ReplaceAll(playSrc, "\\\\", "")), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+playSrc+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":31,"method":"textDocument/codeAction","params":{"textDocument":{"uri":"`+playURI+`"},"range":{"start":{"line":0,"character":10},"end":{"line":0,"character":16}}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 31 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok || len(res) == 0 {
			t.Fatalf("codeAction result missing: %#v", m)
		}
		act, _ := res[0].(map[string]interface{})
		title, _ := act["title"].(string)
		if !strings.Contains(title, "reduce") || !strings.Contains(title, "slug.std") {
			t.Fatalf("unexpected code action title: %q (%#v)", title, act)
		}
		return
	}
	t.Fatal("codeAction response not found")
}

func TestServerCodeActionSourceOnlyReturnsSourceKindImportAction(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	timePath := filepath.Join(libDir, "time.slug")
	timeSrc := "@export val sleep = fn(ms) { nil }\n"
	if err := os.WriteFile(timePath, []byte(timeSrc), 0o644); err != nil {
		t.Fatalf("write module failed: %v", err)
	}
	playPath := filepath.Join(tmp, "playground.slug")
	playSrc := "val { now } = import(\"slug.time\")\nsleep(1)\n"
	playSrcJSON := strings.ReplaceAll(playSrc, "\"", "\\\"")
	playSrcJSON = strings.ReplaceAll(playSrcJSON, "\n", "\\n")
	if err := os.WriteFile(playPath, []byte(playSrc), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+playSrcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":53,"method":"textDocument/codeAction","params":{"textDocument":{"uri":"`+playURI+`"},"range":{"start":{"line":1,"character":1},"end":{"line":1,"character":1}},"context":{"only":["source"]}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 53 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok || len(res) == 0 {
			t.Fatalf("codeAction result missing: %#v", m)
		}
		act, _ := res[0].(map[string]interface{})
		kind, _ := act["kind"].(string)
		if kind != "source.organizeImports" {
			t.Fatalf("expected source.organizeImports kind for source-only request, got %q (%#v)", kind, act)
		}
		return
	}
	t.Fatal("codeAction response not found")
}

func TestServerCodeActionExtendsExistingDestructuredImport(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "@export val map = fn(v){v}\n@export val reduce = fn(v){v}\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	playPath := filepath.Join(tmp, "playground.slug")
	playSrc := "val { map } = import(\"slug.std\")\n[1,2] /> reduce(0, fn(a,b){a})\n"
	playSrcJSON := strings.ReplaceAll(playSrc, "\"", "\\\"")
	playSrcJSON = strings.ReplaceAll(playSrcJSON, "\n", "\\n")
	if err := os.WriteFile(playPath, []byte(playSrc), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+playSrcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":32,"method":"textDocument/codeAction","params":{"textDocument":{"uri":"`+playURI+`"},"range":{"start":{"line":1,"character":9},"end":{"line":1,"character":15}}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 32 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok || len(res) == 0 {
			t.Fatalf("codeAction result missing: %#v", m)
		}
		act, _ := res[0].(map[string]interface{})
		edit, _ := act["edit"].(map[string]interface{})
		changes, _ := edit["changes"].(map[string]interface{})
		edits, _ := changes[playURI].([]interface{})
		if len(edits) != 1 {
			t.Fatalf("expected single edit, got %#v", edits)
		}
		e0, _ := edits[0].(map[string]interface{})
		newText, _ := e0["newText"].(string)
		if newText != ", reduce" {
			t.Fatalf("expected import extension edit ', reduce', got %q", newText)
		}
		return
	}
	t.Fatal("codeAction response not found")
}

func TestServerCodeActionSuggestsQualifyWithExistingImportAlias(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "@export val reduce = fn(v){v}\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	playPath := filepath.Join(tmp, "playground.slug")
	playSrc := "val std = import(\"slug.std\")\n[1,2] /> reduce(0, fn(a,b){a})\n"
	playSrcJSON := strings.ReplaceAll(playSrc, "\"", "\\\"")
	playSrcJSON = strings.ReplaceAll(playSrcJSON, "\n", "\\n")
	if err := os.WriteFile(playPath, []byte(playSrc), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+playSrcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":33,"method":"textDocument/codeAction","params":{"textDocument":{"uri":"`+playURI+`"},"range":{"start":{"line":1,"character":9},"end":{"line":1,"character":15}}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 33 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok || len(res) == 0 {
			t.Fatalf("codeAction result missing: %#v", m)
		}
		first, _ := res[0].(map[string]interface{})
		firstTitle, _ := first["title"].(string)
		if !strings.Contains(firstTitle, "std.reduce") {
			t.Fatalf("expected first ranked action to qualify with std.reduce, got %q", firstTitle)
		}
		found := false
		for _, item := range res {
			act, _ := item.(map[string]interface{})
			title, _ := act["title"].(string)
			if !strings.Contains(title, "std.reduce") {
				continue
			}
			edit, _ := act["edit"].(map[string]interface{})
			changes, _ := edit["changes"].(map[string]interface{})
			edits, _ := changes[playURI].([]interface{})
			if len(edits) != 1 {
				continue
			}
			e0, _ := edits[0].(map[string]interface{})
			newText, _ := e0["newText"].(string)
			if newText == "std.reduce" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected qualify quick-fix std.reduce, got %#v", res)
		}
		return
	}
	t.Fatal("codeAction response not found")
}

func TestServerCodeActionPrefersExtendImportOverQualifyWhenBothAvailable(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "@export val map = fn(v){v}\n@export val reduce = fn(v){v}\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	playPath := filepath.Join(tmp, "playground.slug")
	playSrc := "val std = import(\"slug.std\")\nval { map } = import(\"slug.std\")\n[1,2] /> reduce(0, fn(a,b){a})\n"
	playSrcJSON := strings.ReplaceAll(playSrc, "\"", "\\\"")
	playSrcJSON = strings.ReplaceAll(playSrcJSON, "\n", "\\n")
	if err := os.WriteFile(playPath, []byte(playSrc), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+playSrcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":34,"method":"textDocument/codeAction","params":{"textDocument":{"uri":"`+playURI+`"},"range":{"start":{"line":2,"character":9},"end":{"line":2,"character":15}}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 34 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok || len(res) == 0 {
			t.Fatalf("codeAction result missing: %#v", m)
		}
		first, _ := res[0].(map[string]interface{})
		title, _ := first["title"].(string)
		if !strings.Contains(title, "Extend import") {
			t.Fatalf("expected first ranked action to extend existing import, got %q", title)
		}
		edit, _ := first["edit"].(map[string]interface{})
		changes, _ := edit["changes"].(map[string]interface{})
		edits, _ := changes[playURI].([]interface{})
		if len(edits) != 1 {
			t.Fatalf("expected single edit, got %#v", edits)
		}
		e0, _ := edits[0].(map[string]interface{})
		newText, _ := e0["newText"].(string)
		if newText != ", reduce" {
			t.Fatalf("expected extension edit ', reduce', got %q", newText)
		}
		return
	}
	t.Fatal("codeAction response not found")
}

func TestServerSignatureHelpLocalFunctionAndActiveParam(t *testing.T) {
	src := "val sum = fn(a, b) { a + b }\n[1,2] /> sum(1, 2)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":35,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":17}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 35 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing: %#v", m)
		}
		ap, _ := res["activeParameter"].(float64)
		if int(ap) != 1 {
			t.Fatalf("expected activeParameter=1, got %#v", res["activeParameter"])
		}
		sigs, _ := res["signatures"].([]interface{})
		if len(sigs) == 0 {
			t.Fatalf("expected signatures, got %#v", res)
		}
		s0, _ := sigs[0].(map[string]interface{})
		lbl, _ := s0["label"].(string)
		if !strings.Contains(lbl, "sum(a, b)") {
			t.Fatalf("unexpected signature label: %q", lbl)
		}
		return
	}
	t.Fatal("signatureHelp response not found")
}

func TestServerSignatureHelpImportedFunction(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "@export val reduce = fn(vs, f, init){ init }\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	src := "val { reduce } = import(\"slug.std\")\nreduce([1,2], fn(a,b){a}, 0)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	playPath := filepath.Join(tmp, "playground.slug")
	if err := os.WriteFile(playPath, []byte(src), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":36,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"`+playURI+`"},"position":{"line":1,"character":24}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 36 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing: %#v", m)
		}
		sigs, _ := res["signatures"].([]interface{})
		if len(sigs) == 0 {
			t.Fatalf("expected signatures, got %#v", res)
		}
		s0, _ := sigs[0].(map[string]interface{})
		lbl, _ := s0["label"].(string)
		if !strings.Contains(lbl, "reduce(vs, f, init)") {
			t.Fatalf("unexpected imported signature label: %q", lbl)
		}
		return
	}
	t.Fatal("signatureHelp response not found")
}

func TestServerSignatureHelpIncludesDocAndParamDocs(t *testing.T) {
	src := "/**\n * Adds values.\n * @param a left side\n * @param b right side\n */\n@testWith([1, 2], 3)\n@testWith([\"1\"], 1)\nval add = fn(a, b) { a + b }\nadd(1, 2)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":48,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":8,"character":7}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 48 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing: %#v", m)
		}
		sigs, _ := res["signatures"].([]interface{})
		if len(sigs) == 0 {
			t.Fatalf("expected signatures, got %#v", res)
		}
		s0, _ := sigs[0].(map[string]interface{})
		doc, _ := s0["documentation"].(map[string]interface{})
		dv, _ := doc["value"].(string)
		if !strings.Contains(dv, "Adds values.") {
			t.Fatalf("expected signature documentation to include function docs, got %q", dv)
		}
		if !strings.Contains(dv, "#### Examples") || !strings.Contains(dv, "add(1, 2)  // => 3") {
			t.Fatalf("expected signature documentation to include @testWith examples, got %q", dv)
		}
		if !strings.Contains(dv, "add(\"1\")  // => 1") {
			t.Fatalf("expected string examples to keep quotes, got %q", dv)
		}
		params, _ := s0["parameters"].([]interface{})
		if len(params) != 2 {
			t.Fatalf("expected 2 parameters, got %#v", s0["parameters"])
		}
		p0, _ := params[0].(map[string]interface{})
		p0doc, _ := p0["documentation"].(map[string]interface{})
		p0v, _ := p0doc["value"].(string)
		if !strings.Contains(p0v, "left side") {
			t.Fatalf("expected first parameter doc from @param, got %q", p0v)
		}
		return
	}
	t.Fatal("signatureHelp response not found")
}

func TestServerSignatureHelpImportedWildcardIncludesDocsFromDiskModule(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "/**\n * Reduce wildcard docs.\n * @param vs values\n * @param f reducer\n * @param init seed\n */\n@export val reduce = fn(vs, f, init){ init }\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	src := "var {*} = import(\"slug.std\")\nreduce([1,2], fn(a,b){a}, 0)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	playPath := filepath.Join(tmp, "playground.slug")
	if err := os.WriteFile(playPath, []byte(src), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":51,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"`+playURI+`"},"position":{"line":1,"character":24}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 51 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing: %#v", m)
		}
		sigs, _ := res["signatures"].([]interface{})
		if len(sigs) == 0 {
			t.Fatalf("expected signatures, got %#v", res)
		}
		s0, _ := sigs[0].(map[string]interface{})
		doc, _ := s0["documentation"].(map[string]interface{})
		dv, _ := doc["value"].(string)
		if !strings.Contains(dv, "Reduce wildcard docs.") {
			t.Fatalf("expected wildcard signature docs from module, got %q", dv)
		}
		params, _ := s0["parameters"].([]interface{})
		p1, _ := params[1].(map[string]interface{})
		p1doc, _ := p1["documentation"].(map[string]interface{})
		p1v, _ := p1doc["value"].(string)
		if !strings.Contains(p1v, "reducer") {
			t.Fatalf("expected wildcard parameter docs from @param, got %q", p1v)
		}
		return
	}
	t.Fatal("signatureHelp response not found")
}

func TestServerSignatureHelpImportedMemberCallIncludesDocsFromDiskModule(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "/**\n * Reduce member docs.\n * @param vs values\n * @param f reducer\n * @param init seed\n */\n@export val reduce = fn(vs, f, init){ init }\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	src := "val std = import(\"slug.std\")\nstd.reduce([1,2], fn(a,b){a}, 0)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	playPath := filepath.Join(tmp, "playground.slug")
	if err := os.WriteFile(playPath, []byte(src), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":52,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"`+playURI+`"},"position":{"line":1,"character":28}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 52 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing: %#v", m)
		}
		sigs, _ := res["signatures"].([]interface{})
		if len(sigs) == 0 {
			t.Fatalf("expected signatures, got %#v", res)
		}
		s0, _ := sigs[0].(map[string]interface{})
		doc, _ := s0["documentation"].(map[string]interface{})
		dv, _ := doc["value"].(string)
		if !strings.Contains(dv, "Reduce member docs.") {
			t.Fatalf("expected member-call signature docs from module, got %q", dv)
		}
		return
	}
	t.Fatal("signatureHelp response not found")
}

func TestServerSignatureHelpIgnoresCommasInsideStringsAndNestedCalls(t *testing.T) {
	src := "val sum = fn(a, b, c) { a }\nsum(\"x,y\", sum(1,2), 3)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":37,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":21}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 37 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing: %#v", m)
		}
		ap, _ := res["activeParameter"].(float64)
		if int(ap) != 2 {
			t.Fatalf("expected activeParameter=2, got %#v", res["activeParameter"])
		}
		return
	}
	t.Fatal("signatureHelp response not found")
}

func TestServerSignatureHelpIgnoresCommasInsideMapAndListLiterals(t *testing.T) {
	src := "val sum = fn(a, b, c) { a }\nsum({\"k\":[1,2]}, [3,4], 5)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":38,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":24}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 38 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing: %#v", m)
		}
		ap, _ := res["activeParameter"].(float64)
		if int(ap) != 2 {
			t.Fatalf("expected activeParameter=2, got %#v", res["activeParameter"])
		}
		return
	}
	t.Fatal("signatureHelp response not found")
}

func TestServerSignatureHelpMultilineNestedLiteralCall(t *testing.T) {
	src := "val sum = fn(a, b, c) { a }\nsum(\n  {\"k\": [1, 2]},\n  [3, {\"x\": 4}],\n  5\n)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":40,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":4,"character":3}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 40 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing: %#v", m)
		}
		ap, _ := res["activeParameter"].(float64)
		if int(ap) != 2 {
			t.Fatalf("expected activeParameter=2, got %#v", res["activeParameter"])
		}
		return
	}
	t.Fatal("signatureHelp response not found")
}

func TestServerSignatureHelpReturnsNilOutsideCallContext(t *testing.T) {
	src := "val sum = fn(a, b, c) { a }\nsum(1, 2, 3)\nval x = 42\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":41,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":2,"character":0}}}`) +
			frame(`{"jsonrpc":"2.0","id":42,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":12}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	seen41 := false
	seen42 := false
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok {
			continue
		}
		switch int(id) {
		case 41:
			seen41 = true
			if v, exists := m["result"]; !exists || v != nil {
				t.Fatalf("expected nil signatureHelp result for id=41, got %#v", m["result"])
			}
		case 42:
			seen42 = true
			if v, exists := m["result"]; !exists || v != nil {
				t.Fatalf("expected nil signatureHelp result for id=42, got %#v", m["result"])
			}
		}
	}
	if !seen41 {
		t.Fatal("signatureHelp response id=41 not found")
	}
	if !seen42 {
		t.Fatal("signatureHelp response id=42 not found")
	}
}

func TestServerSignatureHelpStableAcrossTriggerSequence(t *testing.T) {
	src := "val sum = fn(a, b, c) { a }\nsum(1, 2, 3)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":43,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":4}}}`) +
			frame(`{"jsonrpc":"2.0","id":44,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":7}}}`) +
			frame(`{"jsonrpc":"2.0","id":45,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":10}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	want := map[int]int{
		43: 0,
		44: 1,
		45: 2,
	}
	seen := map[int]bool{}
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok {
			continue
		}
		i := int(id)
		expected, tracked := want[i]
		if !tracked {
			continue
		}
		seen[i] = true
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("signatureHelp result missing for id=%d: %#v", i, m)
		}
		as, ok := res["activeSignature"].(float64)
		if !ok || int(as) != 0 {
			t.Fatalf("id=%d expected activeSignature=0, got %#v", i, res["activeSignature"])
		}
		sigs, ok := res["signatures"].([]interface{})
		if !ok || len(sigs) == 0 {
			t.Fatalf("id=%d expected non-empty signatures, got %#v", i, res["signatures"])
		}
		ap, _ := res["activeParameter"].(float64)
		if int(ap) != expected {
			t.Fatalf("id=%d expected activeParameter=%d, got %#v", i, expected, res["activeParameter"])
		}
	}
	for id := range want {
		if !seen[id] {
			t.Fatalf("signatureHelp response id=%d not found", id)
		}
	}
}

func TestServerSignatureHelpCanceledRequestSuppressesResponse(t *testing.T) {
	src := "val sum = fn(a, b, c) { a }\nsum(1, 2, 3)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":46}}`) +
			frame(`{"jsonrpc":"2.0","id":46,"method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":7}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok {
			continue
		}
		if int(id) == 46 {
			t.Fatalf("expected canceled signatureHelp id=46 to be suppressed, got %#v", m)
		}
	}
}

func TestServerSignatureHelpCanceledStringIDSuppressesResponse(t *testing.T) {
	src := "val sum = fn(a, b, c) { a }\nsum(1, 2, 3)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":"sig-1"}}`) +
			frame(`{"jsonrpc":"2.0","id":"sig-1","method":"textDocument/signatureHelp","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":7}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(string)
		if !ok {
			continue
		}
		if id == "sig-1" {
			t.Fatalf("expected canceled signatureHelp id=sig-1 to be suppressed, got %#v", m)
		}
	}
}

func TestFindCallContextHandlesEscapedQuotesInStringArgs(t *testing.T) {
	src := "val sum = fn(a, b, c) { a }\nsum(\"a\\\",b\", 2, 3)\n"
	off := strings.Index(src, ", 3)")
	if off < 0 {
		t.Fatalf("expected marker not found in source: %q", src)
	}
	callee, param, ok := findCallContext(src, off+2)
	if !ok {
		t.Fatal("expected call context to be found")
	}
	if callee != "sum" {
		t.Fatalf("expected callee=sum, got %q", callee)
	}
	if param != 2 {
		t.Fatalf("expected active parameter=2, got %d", param)
	}
}

func TestServerCompletionReturnsKeywordsAndSymbols(t *testing.T) {
	src := "val answer = 42\\nval an = ans\\nv\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":12,"method":"textDocument/completion","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":12}}}`) +
			frame(`{"jsonrpc":"2.0","id":13,"method":"textDocument/completion","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":2,"character":1}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	seen12 := false
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 12 {
			continue
		}
		seen12 = true
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("completion result missing: %#v", m)
		}
		foundAnswer := false
		foundAnd := false
		for _, it := range res {
			item, _ := it.(map[string]interface{})
			label, _ := item["label"].(string)
			if label == "answer" {
				foundAnswer = true
			}
			if label == "and" {
				foundAnd = true
			}
		}
		if !foundAnswer {
			t.Fatalf("expected symbol completion 'answer', got %#v", res)
		}
		if foundAnd {
			t.Fatalf("unexpected non-prefix completion 'and' for prefix 'ans': %#v", res)
		}
		break
	}
	if !seen12 {
		t.Fatal("completion response id=12 not found")
	}
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 13 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("completion result missing: %#v", m)
		}
		foundVal := false
		foundVar := false
		foundAnswer := false
		for _, it := range res {
			item, _ := it.(map[string]interface{})
			label, _ := item["label"].(string)
			if label == "val" {
				foundVal = true
			}
			if label == "var" {
				foundVar = true
			}
			if label == "answer" {
				foundAnswer = true
			}
		}
		if !foundVal || !foundVar {
			t.Fatalf("expected keyword completions 'val' and 'var', got %#v", res)
		}
		if foundAnswer {
			t.Fatalf("unexpected symbol completion 'answer' for prefix 'v': %#v", res)
		}
		return
	}
	t.Fatal("completion response(s) not found")
}

func TestServerCompletionResolveEnrichesItem(t *testing.T) {
	src := "val answer = 42\\nval an = ans\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":14,"method":"completionItem/resolve","params":{"label":"answer","kind":6,"data":{"uri":"file:///a.slug","label":"answer","kind":"variable"}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 14 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("resolve result missing: %#v", m)
		}
		detail, _ := res["detail"].(string)
		if detail == "" {
			t.Fatalf("expected detail to be populated: %#v", res)
		}
		doc, _ := res["documentation"].(map[string]interface{})
		if doc == nil {
			t.Fatalf("expected documentation in resolved item: %#v", res)
		}
		value, _ := doc["value"].(string)
		if !strings.Contains(value, "answer") {
			t.Fatalf("expected documentation mentioning symbol name: %#v", res)
		}
		return
	}
	t.Fatal("completionItem/resolve response not found")
}

func TestServerCompletionResolveEnrichesImportedItemFromModuleDocs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SLUG_HOME", tmp)
	libDir := filepath.Join(tmp, "lib", "slug")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	stdPath := filepath.Join(libDir, "std.slug")
	stdSrc := "/**\n * Reduce docs from std.\n */\n@testWith([[1,2], fn(a,b){a+b}, 0], 3)\n@export val reduce = fn(vs, f, init){ init }\n"
	if err := os.WriteFile(stdPath, []byte(stdSrc), 0o644); err != nil {
		t.Fatalf("write std module failed: %v", err)
	}
	src := "val { reduce } = import(\"slug.std\")\nreduce([1,2], fn(a,b){a}, 0)\n"
	srcJSON := strings.ReplaceAll(src, "\"", "\\\"")
	srcJSON = strings.ReplaceAll(srcJSON, "\n", "\\n")
	playPath := filepath.Join(tmp, "playground.slug")
	if err := os.WriteFile(playPath, []byte(src), 0o644); err != nil {
		t.Fatalf("write playground failed: %v", err)
	}
	playURI, _ := normalizeURI(playPath)
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+playURI+`","version":1,"text":"`+srcJSON+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":49,"method":"completionItem/resolve","params":{"label":"reduce","kind":6,"data":{"uri":"`+playURI+`","label":"reduce","kind":"variable"}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 49 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("resolve result missing: %#v", m)
		}
		detail, _ := res["detail"].(string)
		if detail != "function" {
			t.Fatalf("expected imported symbol kind=function, got %#v", res["detail"])
		}
		doc, _ := res["documentation"].(map[string]interface{})
		if doc == nil {
			t.Fatalf("expected documentation in resolved imported item: %#v", res)
		}
		value, _ := doc["value"].(string)
		if !strings.Contains(value, "Reduce docs from std.") {
			t.Fatalf("expected imported module docs in completion resolve, got %#v", value)
		}
		if !strings.Contains(value, "#### Examples") || !strings.Contains(value, "reduce(") || !strings.Contains(value, "// => 3") {
			t.Fatalf("expected imported module @testWith examples in completion resolve, got %#v", value)
		}
		return
	}
	t.Fatal("completionItem/resolve response not found")
}

func TestServerDocumentHighlightReturnsAllIdentifierMatches(t *testing.T) {
	src := "val answer = 42\\nval use = answer\\nanswer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":15,"method":"textDocument/documentHighlight","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":11}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 15 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("documentHighlight result missing: %#v", m)
		}
		if len(res) != 3 {
			t.Fatalf("expected 3 highlights for 'answer', got %d (%#v)", len(res), res)
		}
		return
	}
	t.Fatal("documentHighlight response not found")
}

func TestServerDocumentHighlightReturnsEmptyOutsideIdentifier(t *testing.T) {
	src := "val answer = 42\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":16,"method":"textDocument/documentHighlight","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":0,"character":3}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 16 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("documentHighlight result missing: %#v", m)
		}
		if len(res) != 0 {
			t.Fatalf("expected empty highlights outside identifier, got %#v", res)
		}
		return
	}
	t.Fatal("documentHighlight response not found")
}

func TestServerReferencesIncludeDeclaration(t *testing.T) {
	src := "val answer = 42\\nval use = answer\\nanswer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":17,"method":"textDocument/references","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":11},"context":{"includeDeclaration":true}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 17 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("references result missing: %#v", m)
		}
		if len(res) != 3 {
			t.Fatalf("expected 3 references including declaration, got %d (%#v)", len(res), res)
		}
		return
	}
	t.Fatal("references response not found")
}

func TestServerReferencesExcludeDeclaration(t *testing.T) {
	src := "val answer = 42\\nval use = answer\\nanswer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":18,"method":"textDocument/references","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":11},"context":{"includeDeclaration":false}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 18 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("references result missing: %#v", m)
		}
		if len(res) != 2 {
			t.Fatalf("expected 2 references excluding declaration, got %d (%#v)", len(res), res)
		}
		return
	}
	t.Fatal("references response not found")
}

func TestServerReferencesIncludeOpenDocumentsForTopLevelSymbol(t *testing.T) {
	srcA := "@export val answer = 42\\nanswer\\n"
	srcB := "val { answer } = import(\\\"a\\\")\\nanswer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/a.slug","version":1,"text":"`+srcA+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/b.slug","version":1,"text":"`+srcB+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":22,"method":"textDocument/references","params":{"textDocument":{"uri":"file:///lib/a.slug"},"position":{"line":1,"character":2},"context":{"includeDeclaration":true}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 22 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("references result missing: %#v", m)
		}
		if len(res) != 4 {
			t.Fatalf("expected 4 module-aware references across exporter/importer docs, got %d (%#v)", len(res), res)
		}
		return
	}
	t.Fatal("references response not found")
}

func TestServerPrepareRenameReturnsRangeAndPlaceholder(t *testing.T) {
	src := "val answer = 42\\nval use = answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":19,"method":"textDocument/prepareRename","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":11}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 19 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("prepareRename result missing: %#v", m)
		}
		placeholder, _ := res["placeholder"].(string)
		if placeholder != "answer" {
			t.Fatalf("expected placeholder=answer, got %q (%#v)", placeholder, res)
		}
		return
	}
	t.Fatal("prepareRename response not found")
}

func TestServerRenameReturnsScopedWorkspaceEdits(t *testing.T) {
	src := "val answer = 42\\nval use = answer\\nanswer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":20,"method":"textDocument/rename","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":11},"newName":"result"}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 20 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("rename result missing: %#v", m)
		}
		changes, ok := res["changes"].(map[string]interface{})
		if !ok {
			t.Fatalf("rename changes missing: %#v", res)
		}
		editsRaw, ok := changes["file:///a.slug"].([]interface{})
		if !ok {
			t.Fatalf("rename edits for uri missing: %#v", changes)
		}
		if len(editsRaw) != 3 {
			t.Fatalf("expected 3 rename edits, got %d (%#v)", len(editsRaw), editsRaw)
		}
		return
	}
	t.Fatal("rename response not found")
}

func TestServerRenameAppliesEditsAcrossOpenDocumentsForTopLevelSymbol(t *testing.T) {
	srcA := "@export val answer = 42\\nanswer\\n"
	srcB := "val { answer } = import(\\\"a\\\")\\nanswer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/a.slug","version":1,"text":"`+srcA+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/b.slug","version":1,"text":"`+srcB+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":23,"method":"textDocument/rename","params":{"textDocument":{"uri":"file:///lib/a.slug"},"position":{"line":1,"character":2},"newName":"result"}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 23 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("rename result missing: %#v", m)
		}
		changes, ok := res["changes"].(map[string]interface{})
		if !ok {
			t.Fatalf("rename changes missing: %#v", res)
		}
		editsA, okA := changes["file:///lib/a.slug"].([]interface{})
		editsB, okB := changes["file:///lib/b.slug"].([]interface{})
		if !okA || !okB {
			t.Fatalf("expected rename edits for both docs, got %#v", changes)
		}
		if len(editsA) != 2 || len(editsB) != 2 {
			t.Fatalf("expected 2 edits per file, got a=%d b=%d", len(editsA), len(editsB))
		}
		return
	}
	t.Fatal("rename response not found")
}

func TestServerRenameRejectsInvalidIdentifier(t *testing.T) {
	src := "val answer = 42\\nval use = answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///a.slug","version":1,"text":"`+src+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":21,"method":"textDocument/rename","params":{"textDocument":{"uri":"file:///a.slug"},"position":{"line":1,"character":11},"newName":"bad-name"}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 21 {
			continue
		}
		errObj, ok := m["error"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected rename error response, got %#v", m)
		}
		code, _ := errObj["code"].(float64)
		if int(code) != -32602 {
			t.Fatalf("expected invalid params error code -32602, got %#v", errObj)
		}
		return
	}
	t.Fatal("rename error response not found")
}

func TestServerReferencesIncludeModuleObjectMemberUsages(t *testing.T) {
	srcA := "@export val answer = 42\\nanswer\\n"
	srcB := "val m = import(\\\"a\\\")\\nm.answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/a.slug","version":1,"text":"`+srcA+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/b.slug","version":1,"text":"`+srcB+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":24,"method":"textDocument/references","params":{"textDocument":{"uri":"file:///lib/a.slug"},"position":{"line":1,"character":2},"context":{"includeDeclaration":true}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 24 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("references result missing: %#v", m)
		}
		if len(res) != 3 {
			t.Fatalf("expected 3 references including module-object member usage, got %d (%#v)", len(res), res)
		}
		return
	}
	t.Fatal("references response not found")
}

func TestServerRenameFromModuleObjectMemberEditsExporterAndUsage(t *testing.T) {
	srcA := "@export val answer = 42\\nanswer\\n"
	srcB := "val m = import(\\\"a\\\")\\nm.answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/a.slug","version":1,"text":"`+srcA+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/b.slug","version":1,"text":"`+srcB+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":25,"method":"textDocument/rename","params":{"textDocument":{"uri":"file:///lib/b.slug"},"position":{"line":1,"character":3},"newName":"result"}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 25 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("rename result missing: %#v", m)
		}
		changes, ok := res["changes"].(map[string]interface{})
		if !ok {
			t.Fatalf("rename changes missing: %#v", res)
		}
		editsA, okA := changes["file:///lib/a.slug"].([]interface{})
		editsB, okB := changes["file:///lib/b.slug"].([]interface{})
		if !okA || !okB {
			t.Fatalf("expected rename edits for exporter and module-object importer docs, got %#v", changes)
		}
		if len(editsA) != 2 || len(editsB) != 1 {
			t.Fatalf("expected edits a=2 b=1, got a=%d b=%d", len(editsA), len(editsB))
		}
		return
	}
	t.Fatal("rename response not found")
}

func TestServerReferencesIncludeInlineImportMemberUsages(t *testing.T) {
	srcA := "@export val answer = 42\\nanswer\\n"
	srcB := "import(\\\"a\\\").answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/a.slug","version":1,"text":"`+srcA+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/b.slug","version":1,"text":"`+srcB+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":26,"method":"textDocument/references","params":{"textDocument":{"uri":"file:///lib/a.slug"},"position":{"line":1,"character":2},"context":{"includeDeclaration":true}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 26 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("references result missing: %#v", m)
		}
		if len(res) != 3 {
			t.Fatalf("expected 3 references including inline import member usage, got %d (%#v)", len(res), res)
		}
		return
	}
	t.Fatal("references response not found")
}

func TestServerRenameFromInlineImportMemberEditsExporterAndUsage(t *testing.T) {
	srcA := "@export val answer = 42\\nanswer\\n"
	srcB := "import(\\\"a\\\").answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/a.slug","version":1,"text":"`+srcA+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/b.slug","version":1,"text":"`+srcB+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":27,"method":"textDocument/rename","params":{"textDocument":{"uri":"file:///lib/b.slug"},"position":{"line":0,"character":13},"newName":"result"}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 27 {
			continue
		}
		res, ok := m["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("rename result missing: %#v", m)
		}
		changes, ok := res["changes"].(map[string]interface{})
		if !ok {
			t.Fatalf("rename changes missing: %#v", res)
		}
		editsA, okA := changes["file:///lib/a.slug"].([]interface{})
		editsB, okB := changes["file:///lib/b.slug"].([]interface{})
		if !okA || !okB {
			t.Fatalf("expected rename edits for exporter and inline importer docs, got %#v", changes)
		}
		if len(editsA) != 2 || len(editsB) != 1 {
			t.Fatalf("expected edits a=2 b=1, got a=%d b=%d", len(editsA), len(editsB))
		}
		return
	}
	t.Fatal("rename response not found")
}

func TestServerReferencesFromMultiModuleInlineImportResolvesWhenUnambiguous(t *testing.T) {
	srcA := "@export val answer = 42\\nanswer\\n"
	srcB := "@export val other = 7\\n"
	srcC := "import(\\\"a\\\",\\\"b\\\").answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/a.slug","version":1,"text":"`+srcA+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/b.slug","version":1,"text":"`+srcB+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/c.slug","version":1,"text":"`+srcC+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":28,"method":"textDocument/references","params":{"textDocument":{"uri":"file:///lib/a.slug"},"position":{"line":1,"character":2},"context":{"includeDeclaration":true}}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 28 {
			continue
		}
		res, ok := m["result"].([]interface{})
		if !ok {
			t.Fatalf("references result missing: %#v", m)
		}
		if len(res) != 3 {
			t.Fatalf("expected 3 references with unambiguous multi-module inline import, got %d (%#v)", len(res), res)
		}
		return
	}
	t.Fatal("references response not found")
}

func TestServerRenameFromAmbiguousMultiModuleInlineImportFailsSafely(t *testing.T) {
	srcA := "@export val answer = 42\\nanswer\\n"
	srcB := "@export val answer = 7\\n"
	srcC := "import(\\\"a\\\",\\\"b\\\").answer\\n"
	in := strings.NewReader(
		frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/a.slug","version":1,"text":"`+srcA+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/b.slug","version":1,"text":"`+srcB+`"}}}`) +
			frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///lib/c.slug","version":1,"text":"`+srcC+`"}}}`) +
			frame(`{"jsonrpc":"2.0","id":29,"method":"textDocument/rename","params":{"textDocument":{"uri":"file:///lib/c.slug"},"position":{"line":0,"character":21},"newName":"result"}}`) +
			frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`) +
			frame(`{"jsonrpc":"2.0","method":"exit"}`),
	)
	var out bytes.Buffer
	s := NewServer(in, &out, func(path, src string) ([]string, []string) { return nil, nil })
	if err := s.Run(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	msgs := readAllMessages(t, out.String())
	for _, m := range msgs {
		id, ok := m["id"].(float64)
		if !ok || int(id) != 29 {
			continue
		}
		errObj, ok := m["error"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected rename error response, got %#v", m)
		}
		code, _ := errObj["code"].(float64)
		if int(code) != -32602 {
			t.Fatalf("expected invalid params error code -32602, got %#v", errObj)
		}
		return
	}
	t.Fatal("rename ambiguous error response not found")
}

func TestNormalizeURIFileScheme(t *testing.T) {
	u, p := normalizeURI("file:///tmp/x.slug")
	if !strings.HasPrefix(u, "file://") {
		t.Fatalf("expected file uri, got %q", u)
	}
	if p == "" {
		t.Fatal("expected local path")
	}
}

func TestParseDiagnosticParsesLocation(t *testing.T) {
	d := parseDiagnostic("ParseError: bad\n    --> /tmp/x.slug:10:5\n", 1, "x")
	if d.Range.Start.Line != 9 || d.Range.Start.Character != 4 {
		t.Fatalf("unexpected range: %+v", d.Range)
	}
	if d.Range.End.Character <= d.Range.Start.Character {
		t.Fatalf("unexpected end character: %+v", d.Range)
	}
}
