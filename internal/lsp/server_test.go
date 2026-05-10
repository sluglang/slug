package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
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
		res, ok := m["result"].([]interface{})
		if !ok || len(res) == 0 {
			t.Fatalf("definition result missing: %#v", m)
		}
		loc, _ := res[0].(map[string]interface{})
		uri, _ := loc["uri"].(string)
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
