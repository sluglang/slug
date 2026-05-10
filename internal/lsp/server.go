package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Analyzer func(path, src string) ([]string, []string)

type Server struct {
	in          *bufio.Reader
	out         io.Writer
	analyze     Analyzer
	docs        map[string]document
	shutdown    bool
	initialized bool
}

type document struct {
	URI     string
	Text    string
	Version int
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
			return nil
		}
	}
}

func (s *Server) handle(req rpcRequest) error {
	switch req.Method {
	case "initialize":
		s.initialized = true
		return s.writeResult(req.ID, initializeResult{Capabilities: serverCapabilities{TextDocumentSync: textDocumentSyncOptions{OpenClose: true, Change: 1}}})
	case "initialized":
		return nil
	case "shutdown":
		s.shutdown = true
		return s.writeResult(req.ID, nil)
	case "exit":
		return nil
	case "textDocument/didOpen":
		var p didOpenParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
		}
		s.docs[p.TextDocument.URI] = document{URI: p.TextDocument.URI, Text: p.TextDocument.Text, Version: p.TextDocument.Version}
		return s.publishDiagnosticsFor(p.TextDocument.URI)
	case "textDocument/didChange":
		var p didChangeParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
		}
		doc, ok := s.docs[p.TextDocument.URI]
		if !ok {
			return nil
		}
		if len(p.ContentChanges) > 0 {
			doc.Text = p.ContentChanges[len(p.ContentChanges)-1].Text
		}
		doc.Version = p.TextDocument.Version
		s.docs[p.TextDocument.URI] = doc
		return s.publishDiagnosticsFor(p.TextDocument.URI)
	case "textDocument/didClose":
		var p didCloseParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
		}
		delete(s.docs, p.TextDocument.URI)
		return s.publishDiagnostics(p.TextDocument.URI, nil)
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
	errs, warns := s.analyze(uri, doc.Text)
	diags := make([]lspDiagnostic, 0, len(errs)+len(warns))
	for _, e := range errs {
		diags = append(diags, parseDiagnostic(e, 1, "slug-semantic"))
	}
	for _, w := range warns {
		diags = append(diags, parseDiagnostic(w, 2, "slug-semantic"))
	}
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
			parts := strings.Split(ln, ":")
			if len(parts) >= 3 {
				if l, err := strconv.Atoi(parts[len(parts)-2]); err == nil {
					line = maxInt(0, l-1)
				}
				if c, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					col = maxInt(0, c-1)
				}
			}
			break
		}
	}
	clean := strings.TrimSpace(msg)
	if i := strings.Index(clean, "ParseError:"); i >= 0 {
		clean = strings.TrimSpace(clean[i+len("ParseError:"):])
		if j := strings.Index(clean, "\n"); j >= 0 {
			clean = strings.TrimSpace(clean[:j])
		}
	}
	if i := strings.Index(clean, "TypeWarning:"); i >= 0 {
		clean = strings.TrimSpace(clean[i+len("TypeWarning:"):])
		if j := strings.Index(clean, "\n"); j >= 0 {
			clean = strings.TrimSpace(clean[:j])
		}
	}
	return lspDiagnostic{
		Range: lspRange{
			Start: lspPosition{Line: line, Character: col},
			End:   lspPosition{Line: line, Character: col + 1},
		},
		Severity: severity,
		Source:   source,
		Message:  clean,
	}
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
			v := strings.TrimSpace(line[len("Content-Length:"):])
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
