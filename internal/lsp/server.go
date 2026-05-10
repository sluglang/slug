package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type Analyzer func(path, src string) ([]string, []string)

type Server struct {
	in           *bufio.Reader
	out          io.Writer
	analyze      Analyzer
	docs         map[string]document
	shutdown     bool
	initialized  bool
	seenShutdown bool
}

type document struct {
	URI      string
	Path     string
	Text     string
	Version  int
	Language string
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type initializeResult struct {
	Capabilities serverCapabilities `json:"capabilities"`
}

type serverCapabilities struct {
	TextDocumentSync textDocumentSyncOptions `json:"textDocumentSync"`
}

type textDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	Change    int  `json:"change"`
}

type didOpenParams struct {
	TextDocument struct {
		URI        string `json:"uri"`
		LanguageID string `json:"languageId"`
		Version    int    `json:"version"`
		Text       string `json:"text"`
	} `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type didCloseParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type publishDiagnosticsParams struct {
	URI         string          `json:"uri"`
	Diagnostics []lspDiagnostic `json:"diagnostics"`
}

type lspDiagnostic struct {
	Range    lspRange `json:"range"`
	Severity int      `json:"severity"`
	Source   string   `json:"source"`
	Message  string   `json:"message"`
}

type lspRange struct {
	Start lspPosition `json:"start"`
	End   lspPosition `json:"end"`
}

type lspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

var diagLocRe = regexp.MustCompile(`-->\s+(.+):(\d+):(\d+)`)

func NewServer(in io.Reader, out io.Writer, analyze Analyzer) *Server {
	return &Server{
		in:      bufio.NewReader(in),
		out:     out,
		analyze: analyze,
		docs:    map[string]document{},
	}
}

func (s *Server) Run() error {
	for {
		body, err := readFramedMessage(s.in)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		var req rpcRequest
		if err := json.Unmarshal(body, &req); err != nil {
			_ = s.writeError(nil, -32700, "Parse error", err.Error())
			continue
		}
		if req.Method == "" {
			continue
		}
		if err := s.handle(req); err != nil {
			return err
		}
		if req.Method == "exit" {
			if !s.seenShutdown {
				return fmt.Errorf("lsp exit received before shutdown")
			}
			return nil
		}
	}
}

func (s *Server) handle(req rpcRequest) error {
	if req.Method != "initialize" && !s.initialized {
		if len(req.ID) > 0 {
			return s.writeError(idOrNil(req.ID), -32002, "Server not initialized", req.Method)
		}
		return nil
	}
	if s.shutdown && req.Method != "exit" {
		if len(req.ID) > 0 {
			return s.writeError(idOrNil(req.ID), -32600, "Invalid request", "server is shut down")
		}
		return nil
	}

	switch req.Method {
	case "initialize":
		s.initialized = true
		return s.writeResult(req.ID, initializeResult{Capabilities: serverCapabilities{TextDocumentSync: textDocumentSyncOptions{OpenClose: true, Change: 1}}})
	case "initialized":
		return nil
	case "shutdown":
		s.shutdown = true
		s.seenShutdown = true
		return s.writeResult(req.ID, nil)
	case "exit":
		return nil
	case "textDocument/didOpen":
		var p didOpenParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
		}
		normURI, normPath := normalizeURI(p.TextDocument.URI)
		s.docs[normURI] = document{URI: normURI, Path: normPath, Text: p.TextDocument.Text, Version: p.TextDocument.Version, Language: p.TextDocument.LanguageID}
		return s.publishDiagnosticsFor(normURI)
	case "textDocument/didChange":
		var p didChangeParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
		}
		normURI, _ := normalizeURI(p.TextDocument.URI)
		doc, ok := s.docs[normURI]
		if !ok {
			// Invalid transition: didChange before didOpen; ignore safely.
			return nil
		}
		if len(p.ContentChanges) > 0 {
			doc.Text = p.ContentChanges[len(p.ContentChanges)-1].Text
		}
		doc.Version = p.TextDocument.Version
		s.docs[normURI] = doc
		return s.publishDiagnosticsFor(normURI)
	case "textDocument/didClose":
		var p didCloseParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
		}
		normURI, _ := normalizeURI(p.TextDocument.URI)
		delete(s.docs, normURI)
		return s.publishDiagnostics(normURI, nil)
	default:
		if len(req.ID) > 0 {
			return s.writeError(idOrNil(req.ID), -32601, "Method not found", req.Method)
		}
		return nil
	}
}

func (s *Server) publishDiagnosticsFor(uri string) error {
	doc, ok := s.docs[uri]
	if !ok {
		return nil
	}
	analyzePath := doc.Path
	if analyzePath == "" {
		analyzePath = doc.URI
	}
	errs, warns := s.analyze(analyzePath, doc.Text)
	diags := make([]lspDiagnostic, 0, len(errs)+len(warns))
	for _, e := range errs {
		diags = append(diags, parseDiagnostic(e, 1, "slug-semantic"))
	}
	for _, w := range warns {
		diags = append(diags, parseDiagnostic(w, 2, "slug-semantic"))
	}
	diags = dedupeDiagnostics(diags)
	return s.publishDiagnostics(uri, diags)
}

func (s *Server) publishDiagnostics(uri string, diags []lspDiagnostic) error {
	msg := rpcRequest{JSONRPC: "2.0", Method: "textDocument/publishDiagnostics"}
	params := publishDiagnosticsParams{URI: uri, Diagnostics: diags}
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	msg.Params = b
	return writeFramedMessage(s.out, msg)
}

func parseDiagnostic(msg string, severity int, source string) lspDiagnostic {
	line := 0
	col := 0
	for _, ln := range strings.Split(msg, "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "-->") {
			matches := diagLocRe.FindStringSubmatch(ln)
			if len(matches) == 4 {
				if l, err := strconv.Atoi(matches[2]); err == nil {
					line = maxInt(0, l-1)
				}
				if c, err := strconv.Atoi(matches[3]); err == nil {
					col = maxInt(0, c-1)
				}
			}
			break
		}
	}
	clean := normalizeDiagnosticMessage(msg)
	endCol := col + 1
	if clean != "" {
		endCol = col + len(clean)
	}
	if endCol < col+1 {
		endCol = col + 1
	}
	return lspDiagnostic{
		Range: lspRange{
			Start: lspPosition{Line: line, Character: col},
			End:   lspPosition{Line: line, Character: endCol},
		},
		Severity: severity,
		Source:   source,
		Message:  clean,
	}
}

func normalizeDiagnosticMessage(msg string) string {
	clean := strings.TrimSpace(msg)
	if i := strings.Index(clean, "ParseError:"); i >= 0 {
		clean = strings.TrimSpace(clean[i+len("ParseError:"):])
	}
	if i := strings.Index(clean, "TypeWarning:"); i >= 0 {
		clean = strings.TrimSpace(clean[i+len("TypeWarning:"):])
	}
	if j := strings.Index(clean, "\n"); j >= 0 {
		clean = strings.TrimSpace(clean[:j])
	}
	return clean
}

func dedupeDiagnostics(in []lspDiagnostic) []lspDiagnostic {
	if len(in) <= 1 {
		return in
	}
	seen := map[string]bool{}
	out := make([]lspDiagnostic, 0, len(in))
	for _, d := range in {
		k := fmt.Sprintf("%d:%d:%d:%d:%d:%s:%s", d.Range.Start.Line, d.Range.Start.Character, d.Range.End.Line, d.Range.End.Character, d.Severity, d.Source, d.Message)
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, d)
	}
	return out
}

func normalizeURI(uri string) (normalizedURI string, localPath string) {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme == "" {
		p := filepath.Clean(uri)
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		return uriFromPath(p), p
	}
	if u.Scheme != "file" {
		return uri, ""
	}
	path := u.Path
	if runtime.GOOS == "windows" && strings.HasPrefix(path, "/") && len(path) > 2 && path[2] == ':' {
		path = path[1:]
	}
	path = filepath.FromSlash(path)
	path = filepath.Clean(path)
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return uriFromPath(path), path
}

func uriFromPath(path string) string {
	p := filepath.ToSlash(path)
	if runtime.GOOS == "windows" {
		if len(p) >= 2 && p[1] == ':' {
			return "file:///" + p
		}
	}
	if strings.HasPrefix(p, "/") {
		return "file://" + p
	}
	return "file:///" + p
}

func readFramedMessage(r *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "content-length:") {
			idx := strings.Index(line, ":")
			if idx < 0 {
				continue
			}
			v := strings.TrimSpace(line[idx+1:])
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("invalid content-length: %w", err)
			}
			contentLength = n
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

func writeFramedMessage(w io.Writer, v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewBufferString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))))
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

func idOrNil(id json.RawMessage) interface{} {
	if len(id) == 0 {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(id, &v); err != nil {
		return nil
	}
	return v
}

func (s *Server) writeResult(id json.RawMessage, result interface{}) error {
	return writeFramedMessage(s.out, rpcResponse{JSONRPC: "2.0", ID: idOrNil(id), Result: result})
}

func (s *Server) writeError(id interface{}, code int, message string, data interface{}) error {
	return writeFramedMessage(s.out, rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message, Data: data}})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
