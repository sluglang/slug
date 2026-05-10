package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"slug/internal/ast"
	"slug/internal/lexer"
	"slug/internal/parser"
	"slug/internal/token"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Analyzer func(path, src string) ([]string, []string)

const diagnosticsDebounceWindow = 75 * time.Millisecond

type Server struct {
	in              *bufio.Reader
	out             io.Writer
	analyze         Analyzer
	docs            map[string]document
	shutdown        bool
	initialized     bool
	seenShutdown    bool
	lastPublishedAt map[string]time.Time
	dirtyDocs       map[string]bool
	canceledReqs    map[string]bool
	handlers        map[string]func(rpcRequest) error
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
	TextDocumentSync          textDocumentSyncOptions `json:"textDocumentSync"`
	HoverProvider             bool                    `json:"hoverProvider,omitempty"`
	DefinitionProvider        bool                    `json:"definitionProvider,omitempty"`
	DocumentSymbolProvider    bool                    `json:"documentSymbolProvider,omitempty"`
	DocumentHighlightProvider bool                    `json:"documentHighlightProvider,omitempty"`
	ReferencesProvider        bool                    `json:"referencesProvider,omitempty"`
	RenameProvider            *renameProvider         `json:"renameProvider,omitempty"`
	CompletionProvider        *completionProvider     `json:"completionProvider,omitempty"`
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

type textDocumentPositionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position lspPosition `json:"position"`
}

type referenceParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position lspPosition `json:"position"`
	Context  struct {
		IncludeDeclaration bool `json:"includeDeclaration"`
	} `json:"context"`
}

type renameParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position lspPosition `json:"position"`
	NewName  string      `json:"newName"`
}

type prepareRenameResult struct {
	Range       lspRange `json:"range"`
	Placeholder string   `json:"placeholder,omitempty"`
}

type completionProvider struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

type renameProvider struct {
	PrepareProvider bool `json:"prepareProvider,omitempty"`
}

type cancelRequestParams struct {
	ID json.RawMessage `json:"id"`
}

type publishDiagnosticsParams struct {
	URI         string          `json:"uri"`
	Diagnostics []lspDiagnostic `json:"diagnostics"`
}

type lspHover struct {
	Contents lspMarkupContent `json:"contents"`
	Range    *lspRange        `json:"range,omitempty"`
}

type lspMarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type lspLocation struct {
	URI   string   `json:"uri"`
	Range lspRange `json:"range"`
}

type lspDocumentSymbol struct {
	Name           string   `json:"name"`
	Kind           int      `json:"kind"`
	Range          lspRange `json:"range"`
	SelectionRange lspRange `json:"selectionRange"`
	Detail         string   `json:"detail,omitempty"`
}

type lspCompletionItem struct {
	Label         string                 `json:"label"`
	Kind          int                    `json:"kind,omitempty"`
	Detail        string                 `json:"detail,omitempty"`
	Documentation *lspMarkupContent      `json:"documentation,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

type lspDocumentHighlight struct {
	Range lspRange `json:"range"`
	Kind  int      `json:"kind,omitempty"`
}

type lspTextEdit struct {
	Range   lspRange `json:"range"`
	NewText string   `json:"newText"`
}

type lspWorkspaceEdit struct {
	Changes map[string][]lspTextEdit `json:"changes,omitempty"`
}

type symbolDef struct {
	Name       string
	Kind       string
	Detail     string
	Start      int
	End        int
	ScopeDepth int
}

type importBinding struct {
	LocalName    string
	SourceModule string
	SourceName   string
}

type moduleSymbolIdentity struct {
	Module string
	Name   string
	Kind   string
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
	s := &Server{
		in:              bufio.NewReader(in),
		out:             out,
		analyze:         analyze,
		docs:            map[string]document{},
		lastPublishedAt: map[string]time.Time{},
		dirtyDocs:       map[string]bool{},
		canceledReqs:    map[string]bool{},
	}
	s.handlers = map[string]func(rpcRequest) error{
		"initialize":                     s.handleInitialize,
		"initialized":                    s.handleInitialized,
		"shutdown":                       s.handleShutdown,
		"exit":                           s.handleExit,
		"textDocument/didOpen":           s.handleDidOpen,
		"textDocument/didChange":         s.handleDidChange,
		"textDocument/didClose":          s.handleDidClose,
		"$/cancelRequest":                s.handleCancelRequest,
		"textDocument/hover":             s.handleHover,
		"textDocument/definition":        s.handleDefinition,
		"textDocument/documentSymbol":    s.handleDocumentSymbol,
		"textDocument/documentHighlight": s.handleDocumentHighlight,
		"textDocument/references":        s.handleReferences,
		"textDocument/prepareRename":     s.handlePrepareRename,
		"textDocument/rename":            s.handleRename,
		"textDocument/completion":        s.handleCompletion,
		"completionItem/resolve":         s.handleCompletionResolve,
	}
	return s
}

func (s *Server) Run() error {
	for {
		_ = s.flushDirtyDocs(false)

		body, err := readFramedMessage(s.in)
		if err == io.EOF {
			_ = s.flushDirtyDocs(true)
			return nil
		}
		if err != nil {
			return err
		}
		slog.Debug("lsp.rpc.inbound", "payload", string(body))

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
			_ = s.flushDirtyDocs(true)
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

	h, ok := s.handlers[req.Method]
	if !ok {
		if len(req.ID) > 0 {
			return s.writeError(idOrNil(req.ID), -32601, "Method not found", req.Method)
		}
		return nil
	}
	if req.Method != "textDocument/didChange" {
		_ = s.flushDirtyDocs(false)
	}
	return h(req)
}

func (s *Server) handleInitialize(req rpcRequest) error {
	s.initialized = true
	return s.writeResult(req.ID, initializeResult{Capabilities: serverCapabilities{
		TextDocumentSync:          textDocumentSyncOptions{OpenClose: true, Change: 1},
		HoverProvider:             true,
		DefinitionProvider:        true,
		DocumentSymbolProvider:    true,
		DocumentHighlightProvider: true,
		ReferencesProvider:        true,
		RenameProvider:            &renameProvider{PrepareProvider: true},
		CompletionProvider:        &completionProvider{ResolveProvider: true},
	}})
}

func (s *Server) handleInitialized(_ rpcRequest) error {
	return nil
}

func (s *Server) handleShutdown(req rpcRequest) error {
	s.shutdown = true
	s.seenShutdown = true
	_ = s.flushDirtyDocs(true)
	return s.writeResult(req.ID, nil)
}

func (s *Server) handleExit(_ rpcRequest) error {
	return nil
}

func (s *Server) handleDidOpen(req rpcRequest) error {
	var p didOpenParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, normPath := normalizeURI(p.TextDocument.URI)
	s.docs[normURI] = document{URI: normURI, Path: normPath, Text: p.TextDocument.Text, Version: p.TextDocument.Version, Language: p.TextDocument.LanguageID}
	return s.publishDiagnosticsFor(normURI)
}

func (s *Server) handleDidChange(req rpcRequest) error {
	var p didChangeParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return nil
	}
	if len(p.ContentChanges) > 0 {
		doc.Text = p.ContentChanges[len(p.ContentChanges)-1].Text
	}
	doc.Version = p.TextDocument.Version
	s.docs[normURI] = doc

	if last, ok := s.lastPublishedAt[normURI]; ok && time.Since(last) < diagnosticsDebounceWindow {
		s.dirtyDocs[normURI] = true
		return nil
	}
	return s.publishDiagnosticsFor(normURI)
}

func (s *Server) handleDidClose(req rpcRequest) error {
	var p didCloseParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	delete(s.docs, normURI)
	delete(s.dirtyDocs, normURI)
	delete(s.lastPublishedAt, normURI)
	return s.publishDiagnostics(normURI, nil)
}

func (s *Server) handleCancelRequest(req rpcRequest) error {
	var p cancelRequestParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return nil
	}
	if len(p.ID) == 0 {
		return nil
	}
	id := string(bytes.TrimSpace(p.ID))
	s.canceledReqs[id] = true
	return nil
}

func (s *Server) handleHover(req rpcRequest) error {
	var p textDocumentPositionParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, nil)
	}
	offset := offsetFromPosition(doc.Text, p.Position.Line, p.Position.Character)
	name, start, end := identifierAtOffset(doc.Text, offset)
	if name == "" {
		return s.writeResult(req.ID, nil)
	}
	syms := collectSymbols(doc.Text)
	sym, found := resolveSymbolAt(name, offset, syms)
	if !found {
		return s.writeResult(req.ID, &lspHover{
			Contents: lspMarkupContent{Kind: "markdown", Value: "`" + name + "`"},
		})
	}
	rng := offsetRangeToLSP(doc.Text, start, end)
	return s.writeResult(req.ID, &lspHover{
		Contents: lspMarkupContent{Kind: "markdown", Value: fmt.Sprintf("`%s` (%s)%s", sym.Name, sym.Kind, hoverDetail(sym.Detail))},
		Range:    &rng,
	})
}

func hoverDetail(d string) string {
	if strings.TrimSpace(d) == "" {
		return ""
	}
	return "\n\n" + d
}

func (s *Server) handleDefinition(req rpcRequest) error {
	var p textDocumentPositionParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, nil)
	}
	offset := offsetFromPosition(doc.Text, p.Position.Line, p.Position.Character)
	name, _, _ := identifierAtOffset(doc.Text, offset)
	if name == "" {
		return s.writeResult(req.ID, nil)
	}
	syms := collectSymbols(doc.Text)
	sym, found := resolveSymbolAt(name, offset, syms)
	if !found {
		return s.writeResult(req.ID, nil)
	}
	defRange := offsetRangeToLSP(doc.Text, sym.Start, sym.End)
	loc := lspLocation{
		URI:   normURI,
		Range: defRange,
	}
	return s.writeResult(req.ID, []lspLocation{loc})
}

func (s *Server) handleDocumentSymbol(req rpcRequest) error {
	var p struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, []lspDocumentSymbol{})
	}
	syms := collectTopLevelSymbols(doc.Text)
	out := make([]lspDocumentSymbol, 0, len(syms))
	for _, sym := range syms {
		rng := offsetRangeToLSP(doc.Text, sym.Start, sym.End)
		out = append(out, lspDocumentSymbol{
			Name:           sym.Name,
			Kind:           toDocumentSymbolKind(sym.Kind),
			Range:          rng,
			SelectionRange: rng,
			Detail:         sym.Detail,
		})
	}
	return s.writeResult(req.ID, out)
}

func (s *Server) handleDocumentHighlight(req rpcRequest) error {
	var p textDocumentPositionParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, []lspDocumentHighlight{})
	}
	offset := offsetFromPosition(doc.Text, p.Position.Line, p.Position.Character)
	name, _, _ := identifierAtOffset(doc.Text, offset)
	if name == "" {
		return s.writeResult(req.ID, []lspDocumentHighlight{})
	}
	highlights := collectIdentifierHighlights(doc.Text, name)
	return s.writeResult(req.ID, highlights)
}

func (s *Server) handleReferences(req rpcRequest) error {
	var p referenceParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, []lspLocation{})
	}
	offset := offsetFromPosition(doc.Text, p.Position.Line, p.Position.Character)
	name, _, _ := identifierAtOffset(doc.Text, offset)
	if name == "" {
		return s.writeResult(req.ID, []lspLocation{})
	}
	syms := collectSymbols(doc.Text)
	target, found := resolveSymbolAt(name, offset, syms)
	if !found {
		return s.writeResult(req.ID, []lspLocation{})
	}
	includeDecl := p.Context.IncludeDeclaration
	refs := s.collectReferencesAcrossOpenDocs(normURI, name, target, includeDecl)
	return s.writeResult(req.ID, refs)
}

func (s *Server) handlePrepareRename(req rpcRequest) error {
	var p textDocumentPositionParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, nil)
	}
	offset := offsetFromPosition(doc.Text, p.Position.Line, p.Position.Character)
	name, start, end := identifierAtOffset(doc.Text, offset)
	if name == "" {
		return s.writeResult(req.ID, nil)
	}
	syms := collectSymbols(doc.Text)
	target, found := resolveSymbolAt(name, offset, syms)
	if !found {
		return s.writeResult(req.ID, nil)
	}
	rng := offsetRangeToLSP(doc.Text, start, end)
	return s.writeResult(req.ID, prepareRenameResult{Range: rng, Placeholder: target.Name})
}

func (s *Server) handleRename(req rpcRequest) error {
	var p renameParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	if !isValidIdentifierName(p.NewName) {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", "newName must be a valid identifier")
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, &lspWorkspaceEdit{Changes: map[string][]lspTextEdit{}})
	}
	offset := offsetFromPosition(doc.Text, p.Position.Line, p.Position.Character)
	name, _, _ := identifierAtOffset(doc.Text, offset)
	if name == "" {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", "cursor is not on a renameable symbol")
	}
	syms := collectSymbols(doc.Text)
	target, found := resolveSymbolAt(name, offset, syms)
	if !found {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", "could not resolve symbol for rename")
	}
	refs := s.collectReferencesAcrossOpenDocs(normURI, name, target, true)
	changes := map[string][]lspTextEdit{}
	for _, ref := range refs {
		changes[ref.URI] = append(changes[ref.URI], lspTextEdit{
			Range:   ref.Range,
			NewText: p.NewName,
		})
	}
	return s.writeResult(req.ID, &lspWorkspaceEdit{
		Changes: changes,
	})
}

func (s *Server) handleCompletion(req rpcRequest) error {
	var p textDocumentPositionParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, []lspCompletionItem{})
	}

	offset := offsetFromPosition(doc.Text, p.Position.Line, p.Position.Character)
	prefix := completionPrefixAtOffset(doc.Text, offset)
	syms := collectSymbols(doc.Text)
	seen := map[string]bool{}
	items := make([]lspCompletionItem, 0, 64)

	add := func(label string, kind int, detail string) {
		if label == "" {
			return
		}
		if prefix != "" && !strings.HasPrefix(label, prefix) {
			return
		}
		if seen[label] {
			return
		}
		seen[label] = true
		items = append(items, lspCompletionItem{
			Label:  label,
			Kind:   kind,
			Detail: detail,
			Data: map[string]interface{}{
				"uri":   normURI,
				"label": label,
				"kind":  detail,
			},
		})
	}

	for _, kw := range slugKeywords {
		add(kw, 14, "keyword")
	}
	for _, s := range syms {
		add(s.Name, toCompletionItemKind(s.Kind), s.Kind)
	}
	return s.writeResult(req.ID, items)
}

func (s *Server) handleCompletionResolve(req rpcRequest) error {
	var item lspCompletionItem
	if err := json.Unmarshal(req.Params, &item); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	if item.Data == nil {
		return s.writeResult(req.ID, item)
	}
	uri, _ := item.Data["uri"].(string)
	label, _ := item.Data["label"].(string)
	if uri == "" || label == "" {
		return s.writeResult(req.ID, item)
	}
	doc, ok := s.docs[uri]
	if !ok {
		return s.writeResult(req.ID, item)
	}

	syms := collectSymbols(doc.Text)
	best := symbolDef{}
	found := false
	for _, s := range syms {
		if s.Name != label {
			continue
		}
		if !found || s.ScopeDepth < best.ScopeDepth || (s.ScopeDepth == best.ScopeDepth && s.Start < best.Start) {
			best = s
			found = true
		}
	}
	if !found {
		if item.Detail == "" {
			item.Detail = "symbol"
		}
		item.Documentation = &lspMarkupContent{
			Kind:  "markdown",
			Value: fmt.Sprintf("`%s`", item.Label),
		}
		return s.writeResult(req.ID, item)
	}
	item.Kind = toCompletionItemKind(best.Kind)
	item.Detail = best.Kind
	docLines := []string{fmt.Sprintf("`%s` (%s)", best.Name, best.Kind)}
	if strings.TrimSpace(best.Detail) != "" {
		docLines = append(docLines, "", best.Detail)
	}
	item.Documentation = &lspMarkupContent{
		Kind:  "markdown",
		Value: strings.Join(docLines, "\n"),
	}
	return s.writeResult(req.ID, item)
}

func toDocumentSymbolKind(kind string) int {
	switch kind {
	case "function":
		return 12
	case "struct":
		return 23
	default:
		return 13
	}
}

func toCompletionItemKind(kind string) int {
	switch kind {
	case "function":
		return 3
	case "parameter", "variable":
		return 6
	case "struct":
		return 22
	default:
		return 1
	}
}

var slugKeywords = []string{
	"nil", "true", "false",
	"fn", "foreign", "val", "var", "struct", "copy",
	"if", "else", "match", "return", "recur",
	"throw", "defer",
	"nursery", "spawn", "select",
}

func (s *Server) flushDirtyDocs(force bool) error {
	for uri := range s.dirtyDocs {
		if !force {
			if last, ok := s.lastPublishedAt[uri]; ok && time.Since(last) < diagnosticsDebounceWindow {
				continue
			}
		}
		if err := s.publishDiagnosticsFor(uri); err != nil {
			return err
		}
		delete(s.dirtyDocs, uri)
	}
	return nil
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
	if err := s.publishDiagnostics(uri, diags); err != nil {
		return err
	}
	s.lastPublishedAt[uri] = time.Now()
	return nil
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

func collectSymbols(src string) []symbolDef {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	syms := []symbolDef{}
	scopeDepth := 0
	var walkStmt func(ast.Statement)
	var walkExpr func(ast.Expression)
	var addPattern func(ast.MatchPattern, string, string)

	addPattern = func(pat ast.MatchPattern, kind string, detail string) {
		switch p := pat.(type) {
		case *ast.IdentifierPattern:
			if p != nil && p.Value != nil {
				start := p.Value.Token.Position
				end := start + len(p.Value.Token.Literal)
				syms = append(syms, symbolDef{Name: p.Value.Value, Kind: kind, Detail: detail, Start: start, End: end, ScopeDepth: scopeDepth})
			}
		case *ast.BindingPattern:
			if p.Name != nil {
				start := p.Name.Token.Position
				end := start + len(p.Name.Token.Literal)
				syms = append(syms, symbolDef{Name: p.Name.Value, Kind: kind, Detail: detail, Start: start, End: end, ScopeDepth: scopeDepth})
			}
			if p.Pattern != nil {
				addPattern(p.Pattern, kind, detail)
			}
		case *ast.ListPattern:
			for _, el := range p.Elements {
				addPattern(el, kind, detail)
			}
		case *ast.MapPattern:
			for _, pair := range p.Pairs {
				addPattern(pair.Pattern, kind, detail)
			}
			if p.Spread != nil {
				addPattern(p.Spread, kind, detail)
			}
		case *ast.StructPattern:
			for _, f := range p.Fields {
				addPattern(f.Pattern, kind, detail)
			}
		case *ast.SpreadPattern:
			if p.Value != nil {
				start := p.Value.Token.Position
				end := start + len(p.Value.Token.Literal)
				syms = append(syms, symbolDef{Name: p.Value.Value, Kind: kind, Detail: detail, Start: start, End: end, ScopeDepth: scopeDepth})
			}
		case *ast.MultiPattern:
			for _, sub := range p.Patterns {
				addPattern(sub, kind, detail)
			}
		}
	}

	walkStmt = func(st ast.Statement) {
		switch s := st.(type) {
		case *ast.ExpressionStatement:
			walkExpr(s.Expression)
		case *ast.ReturnStatement:
			walkExpr(s.ReturnValue)
		case *ast.ThrowStatement:
			walkExpr(s.Value)
		case *ast.BlockStatement:
			scopeDepth++
			for _, ch := range s.Statements {
				walkStmt(ch)
			}
			scopeDepth--
		case *ast.DeferStatement:
			if s.Call != nil {
				walkStmt(s.Call)
			}
		}
	}

	walkExpr = func(ex ast.Expression) {
		switch e := ex.(type) {
		case *ast.VarExpression:
			addPattern(e.Pattern, "variable", "")
			walkExpr(e.Value)
		case *ast.ValExpression:
			detail := ""
			if _, ok := e.Value.(*ast.FunctionLiteral); ok {
				detail = "function"
			}
			addPattern(e.Pattern, "variable", detail)
			walkExpr(e.Value)
		case *ast.FunctionLiteral:
			scopeDepth++
			for _, p := range e.Parameters {
				if p != nil && p.Name != nil {
					start := p.Name.Token.Position
					end := start + len(p.Name.Token.Literal)
					syms = append(syms, symbolDef{Name: p.Name.Value, Kind: "parameter", Detail: "", Start: start, End: end, ScopeDepth: scopeDepth})
				}
			}
			if e.Body != nil {
				for _, st := range e.Body.Statements {
					walkStmt(st)
				}
			}
			scopeDepth--
		case *ast.IfExpression:
			walkExpr(e.Condition)
			if e.ThenBranch != nil {
				walkStmt(e.ThenBranch)
			}
			if e.ElseBranch != nil {
				walkStmt(e.ElseBranch)
			}
		case *ast.InfixExpression:
			walkExpr(e.Left)
			walkExpr(e.Right)
		case *ast.PrefixExpression:
			walkExpr(e.Right)
		case *ast.CallExpression:
			walkExpr(e.Function)
			for _, a := range e.Arguments {
				walkExpr(a)
			}
		case *ast.MatchExpression:
			walkExpr(e.Value)
			for _, cs := range e.Cases {
				if cs == nil {
					continue
				}
				scopeDepth++
				if cs.Pattern != nil {
					addPattern(cs.Pattern, "variable", "match binding")
				}
				walkExpr(cs.Guard)
				if cs.Body != nil {
					for _, st := range cs.Body.Statements {
						walkStmt(st)
					}
				}
				scopeDepth--
			}
		case *ast.ListLiteral:
			for _, it := range e.Elements {
				walkExpr(it)
			}
		case *ast.MapLiteral:
			for k, v := range e.Pairs {
				walkExpr(k)
				walkExpr(v)
			}
		case *ast.IndexExpression:
			walkExpr(e.Left)
			walkExpr(e.Index)
		case *ast.StructInitExpression:
			walkExpr(e.Schema)
			for _, f := range e.Fields {
				walkExpr(f.Value)
			}
		case *ast.StructCopyExpression:
			walkExpr(e.Source)
			walkExpr(e.Fields)
		case *ast.SpawnExpression:
			walkExpr(e.Body)
		case *ast.BlockStatement:
			walkStmt(e)
		}
	}

	for _, st := range program.Statements {
		walkStmt(st)
	}
	return syms
}

func collectTopLevelSymbols(src string) []symbolDef {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	out := []symbolDef{}
	addPatternTop := func(pat ast.MatchPattern, kind string, detail string) {
		switch p := pat.(type) {
		case *ast.IdentifierPattern:
			if p != nil && p.Value != nil {
				start := p.Value.Token.Position
				end := start + len(p.Value.Token.Literal)
				out = append(out, symbolDef{Name: p.Value.Value, Kind: kind, Detail: detail, Start: start, End: end})
			}
		case *ast.BindingPattern:
			if p.Name != nil {
				start := p.Name.Token.Position
				end := start + len(p.Name.Token.Literal)
				out = append(out, symbolDef{Name: p.Name.Value, Kind: kind, Detail: detail, Start: start, End: end})
			}
		}
	}
	for _, st := range program.Statements {
		es, ok := st.(*ast.ExpressionStatement)
		if !ok || es.Expression == nil {
			continue
		}
		switch e := es.Expression.(type) {
		case *ast.ValExpression:
			detail := ""
			kind := "variable"
			switch e.Value.(type) {
			case *ast.FunctionLiteral:
				kind = "function"
				detail = "fn"
			case *ast.StructSchemaExpression:
				kind = "struct"
				detail = "struct"
			}
			addPatternTop(e.Pattern, kind, detail)
		case *ast.VarExpression:
			addPatternTop(e.Pattern, "variable", "var")
		}
	}
	return out
}

func resolveSymbolAt(name string, useOffset int, syms []symbolDef) (symbolDef, bool) {
	best := symbolDef{}
	found := false
	for _, s := range syms {
		if s.Name != name {
			continue
		}
		if s.Start > useOffset {
			continue
		}
		if !found || s.ScopeDepth > best.ScopeDepth || (s.ScopeDepth == best.ScopeDepth && s.Start > best.Start) {
			best = s
			found = true
		}
	}
	return best, found
}

func identifierAtOffset(src string, off int) (name string, start int, end int) {
	if off < 0 {
		return "", 0, 0
	}
	if off > len(src) {
		off = len(src)
	}
	if name, start, end, ok := identTokenAtByteOffset(src, off); ok {
		return name, start, end
	}
	// LSP positions are between characters; clients often send end-of-word positions.
	if off > 0 {
		if name, start, end, ok := identTokenAtByteOffset(src, off-1); ok {
			return name, start, end
		}
	}
	return "", 0, 0
}

func completionPrefixAtOffset(src string, off int) string {
	name, start, _ := identifierAtOffset(src, off)
	if name == "" {
		return ""
	}
	if off < start {
		return ""
	}
	end := start + len(name)
	if off > end {
		off = end
	}
	if off-start <= 0 {
		return ""
	}
	return name[:off-start]
}

func collectIdentifierHighlights(src string, name string) []lspDocumentHighlight {
	if name == "" {
		return []lspDocumentHighlight{}
	}
	l := lexer.New(src)
	out := make([]lspDocumentHighlight, 0, 16)
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF || tok.Type == token.ILLEGAL {
			break
		}
		if tok.Type != token.IDENT || tok.Literal != name {
			continue
		}
		start := tok.Position
		end := start + len(tok.Literal)
		out = append(out, lspDocumentHighlight{
			Range: offsetRangeToLSP(src, start, end),
			Kind:  1,
		})
	}
	return out
}

func collectScopedReferenceLocations(src string, uri string, name string, target symbolDef, syms []symbolDef, includeDeclaration bool) []lspLocation {
	if name == "" {
		return []lspLocation{}
	}
	l := lexer.New(src)
	out := make([]lspLocation, 0, 16)
	seen := map[string]bool{}
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF || tok.Type == token.ILLEGAL {
			break
		}
		if tok.Type != token.IDENT || tok.Literal != name {
			continue
		}
		resolved, ok := resolveSymbolAt(name, tok.Position, syms)
		if !ok || resolved.Start != target.Start || resolved.End != target.End {
			continue
		}
		if !includeDeclaration && tok.Position == target.Start {
			continue
		}
		rng := offsetRangeToLSP(src, tok.Position, tok.Position+len(tok.Literal))
		key := fmt.Sprintf("%d:%d:%d:%d", rng.Start.Line, rng.Start.Character, rng.End.Line, rng.End.Character)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, lspLocation{
			URI:   uri,
			Range: rng,
		})
	}
	return out
}

func (s *Server) collectReferencesAcrossOpenDocs(originURI string, name string, target symbolDef, includeDeclaration bool) []lspLocation {
	originDoc, ok := s.docs[originURI]
	if !ok {
		return []lspLocation{}
	}
	originSyms := collectSymbols(originDoc.Text)
	out := collectScopedReferenceLocations(originDoc.Text, originURI, name, target, originSyms, includeDeclaration)
	identity, ok := s.resolveModuleSymbolIdentity(originURI, target)
	if !ok {
		return out
	}

	for uri, doc := range s.docs {
		if uri == originURI {
			continue
		}
		syms := collectSymbols(doc.Text)
		docModule := moduleNameFromURI(uri)
		if docModule == identity.Module {
			targets := findTopLevelSymbolsByNameAndKind(syms, identity.Name, identity.Kind)
			for _, t := range targets {
				refs := collectScopedReferenceLocations(doc.Text, uri, identity.Name, t, syms, includeDeclaration)
				out = append(out, refs...)
			}
		}
		bindings := collectImportBindingsForModule(doc.Text, identity.Module)
		for _, b := range bindings {
			if b.SourceName != identity.Name {
				continue
			}
			aliasTargets := findTopLevelSymbolsByName(syms, b.LocalName)
			for _, t := range aliasTargets {
				refs := collectScopedReferenceLocations(doc.Text, uri, b.LocalName, t, syms, includeDeclaration)
				out = append(out, refs...)
			}
		}
	}
	return dedupeLocations(out)
}

func (s *Server) resolveModuleSymbolIdentity(originURI string, target symbolDef) (moduleSymbolIdentity, bool) {
	originDoc, ok := s.docs[originURI]
	if !ok || target.ScopeDepth != 0 {
		return moduleSymbolIdentity{}, false
	}

	module := moduleNameFromURI(originURI)
	if module == "" {
		return moduleSymbolIdentity{}, false
	}

	bindings := collectImportBindingsForModule(originDoc.Text, "")
	for _, b := range bindings {
		if b.LocalName == target.Name {
			kind := target.Kind
			if kind == "variable" {
				kind = s.inferExportKindFromOpenDocs(b.SourceModule, b.SourceName, kind)
			}
			return moduleSymbolIdentity{Module: b.SourceModule, Name: b.SourceName, Kind: kind}, true
		}
	}
	return moduleSymbolIdentity{Module: module, Name: target.Name, Kind: target.Kind}, true
}

func (s *Server) inferExportKindFromOpenDocs(module string, name string, fallback string) string {
	for uri, doc := range s.docs {
		if moduleNameFromURI(uri) != module {
			continue
		}
		for _, exp := range collectExportedTopLevelSymbols(doc.Text) {
			if exp.Name == name {
				return exp.Kind
			}
		}
	}
	return fallback
}

func findTopLevelSymbolsByNameAndKind(syms []symbolDef, name string, kind string) []symbolDef {
	out := make([]symbolDef, 0, 2)
	for _, s := range syms {
		if s.ScopeDepth != 0 || s.Name != name {
			continue
		}
		if kind != "" && s.Kind != kind {
			continue
		}
		out = append(out, s)
	}
	return out
}

func findTopLevelSymbolsByName(syms []symbolDef, name string) []symbolDef {
	out := make([]symbolDef, 0, 2)
	for _, s := range syms {
		if s.ScopeDepth == 0 && s.Name == name {
			out = append(out, s)
		}
	}
	return out
}

func collectExportedTopLevelSymbols(src string) []symbolDef {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	out := []symbolDef{}

	for _, st := range program.Statements {
		if ff, ok := st.(*ast.ForeignFunctionDeclaration); ok {
			if !hasTag(ff.Tags, "@export") || ff.Name == nil {
				continue
			}
			start := ff.Name.Token.Position
			end := start + len(ff.Name.Token.Literal)
			out = append(out, symbolDef{Name: ff.Name.Value, Kind: "function", Start: start, End: end, ScopeDepth: 0})
			continue
		}

		es, ok := st.(*ast.ExpressionStatement)
		if !ok || es.Expression == nil {
			continue
		}
		switch e := es.Expression.(type) {
		case *ast.ValExpression:
			if !hasTag(e.Tags, "@export") {
				continue
			}
			kind := "variable"
			if _, ok := e.Value.(*ast.FunctionLiteral); ok {
				kind = "function"
			}
			for _, n := range topLevelPatternNames(e.Pattern) {
				out = append(out, symbolDef{Name: n.Name, Kind: kind, Start: n.Start, End: n.End, ScopeDepth: 0})
			}
		case *ast.VarExpression:
			if !hasTag(e.Tags, "@export") {
				continue
			}
			for _, n := range topLevelPatternNames(e.Pattern) {
				out = append(out, symbolDef{Name: n.Name, Kind: "variable", Start: n.Start, End: n.End, ScopeDepth: 0})
			}
		}
	}
	return out
}

func collectImportBindingsForModule(src string, onlyModule string) []importBinding {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	out := []importBinding{}

	for _, st := range program.Statements {
		es, ok := st.(*ast.ExpressionStatement)
		if !ok || es.Expression == nil {
			continue
		}
		var pattern ast.MatchPattern
		var value ast.Expression
		switch e := es.Expression.(type) {
		case *ast.ValExpression:
			pattern, value = e.Pattern, e.Value
		case *ast.VarExpression:
			pattern, value = e.Pattern, e.Value
		default:
			continue
		}
		call, ok := value.(*ast.CallExpression)
		if !ok {
			continue
		}
		fn, ok := call.Function.(*ast.Identifier)
		if !ok || fn.Value != "import" {
			continue
		}
		modules := importModuleArgs(call.Arguments)
		if len(modules) != 1 {
			continue
		}
		module := modules[0]
		if onlyModule != "" && module != onlyModule {
			continue
		}
		bindings := importBindingsFromPattern(pattern, module)
		out = append(out, bindings...)
	}
	return out
}

func importModuleArgs(args []ast.Expression) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		s, ok := a.(*ast.StringLiteral)
		if !ok {
			return nil
		}
		out = append(out, s.Value)
	}
	return out
}

func importBindingsFromPattern(pat ast.MatchPattern, module string) []importBinding {
	mp, ok := pat.(*ast.MapPattern)
	if !ok || mp == nil {
		return nil
	}
	out := []importBinding{}
	for _, entry := range mp.Pairs {
		source := patternEntryKeyName(entry.Key)
		local := patternBindingName(entry.Pattern)
		if source == "" && local != "" {
			source = local
		}
		if source == "" || local == "" {
			continue
		}
		out = append(out, importBinding{
			LocalName:    local,
			SourceModule: module,
			SourceName:   source,
		})
	}
	return out
}

func patternEntryKeyName(key ast.Expression) string {
	switch k := key.(type) {
	case *ast.Identifier:
		return k.Value
	case *ast.StringLiteral:
		return k.Value
	case *ast.SymbolLiteral:
		return k.Value
	default:
		return ""
	}
}

func patternBindingName(pat ast.MatchPattern) string {
	switch p := pat.(type) {
	case *ast.IdentifierPattern:
		if p != nil && p.Value != nil {
			return p.Value.Value
		}
	case *ast.BindingPattern:
		if p != nil && p.Name != nil {
			return p.Name.Value
		}
	}
	return ""
}

type patternNameRef struct {
	Name  string
	Start int
	End   int
}

func topLevelPatternNames(pat ast.MatchPattern) []patternNameRef {
	out := []patternNameRef{}
	switch p := pat.(type) {
	case *ast.IdentifierPattern:
		if p != nil && p.Value != nil {
			start := p.Value.Token.Position
			end := start + len(p.Value.Token.Literal)
			out = append(out, patternNameRef{Name: p.Value.Value, Start: start, End: end})
		}
	case *ast.BindingPattern:
		if p != nil && p.Name != nil {
			start := p.Name.Token.Position
			end := start + len(p.Name.Token.Literal)
			out = append(out, patternNameRef{Name: p.Name.Value, Start: start, End: end})
		}
	}
	return out
}

func hasTag(tags []*ast.Tag, name string) bool {
	for _, t := range tags {
		if t != nil && t.Name == name {
			return true
		}
	}
	return false
}

func moduleNameFromURI(uri string) string {
	_, local := normalizeURI(uri)
	if local == "" {
		return ""
	}
	p := filepath.ToSlash(local)
	if strings.HasSuffix(p, ".slug") {
		p = strings.TrimSuffix(p, ".slug")
	}
	if idx := strings.LastIndex(p, "/lib/"); idx >= 0 {
		p = p[idx+len("/lib/"):]
	}
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return ""
	}
	return strings.ReplaceAll(p, "/", ".")
}

func dedupeLocations(in []lspLocation) []lspLocation {
	seen := map[string]bool{}
	out := make([]lspLocation, 0, len(in))
	for _, loc := range in {
		key := fmt.Sprintf("%s:%d:%d:%d:%d", loc.URI, loc.Range.Start.Line, loc.Range.Start.Character, loc.Range.End.Line, loc.Range.End.Character)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, loc)
	}
	return out
}

func isValidIdentifierName(name string) bool {
	if name == "" {
		return false
	}
	r, size := utf8.DecodeRuneInString(name)
	if r == utf8.RuneError && size == 0 {
		return false
	}
	if !(unicode.IsLetter(r) || r == '_') {
		return false
	}
	for i := size; i < len(name); {
		ch, w := utf8.DecodeRuneInString(name[i:])
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_') {
			return false
		}
		i += w
	}
	return true
}

func identTokenAtByteOffset(src string, off int) (name string, start int, end int, ok bool) {
	if off < 0 || off >= len(src) {
		return "", 0, 0, false
	}
	l := lexer.New(src)
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF || tok.Type == token.ILLEGAL {
			return "", 0, 0, false
		}
		if tok.Type != token.IDENT {
			continue
		}
		start = tok.Position
		end = start + len(tok.Literal)
		if off >= start && off < end {
			return tok.Literal, start, end, true
		}
	}
}

func offsetFromPosition(src string, line int, col int) int {
	if line < 0 {
		return 0
	}
	curLine := 0
	i := 0
	for i < len(src) && curLine < line {
		r, size := utf8.DecodeRuneInString(src[i:])
		i += size
		if r == '\n' {
			curLine++
		}
	}
	curCol := 0
	for i < len(src) && curCol < col {
		r, size := utf8.DecodeRuneInString(src[i:])
		if r == '\n' {
			break
		}
		i += size
		curCol++
	}
	return i
}

func positionFromOffset(src string, off int) lspPosition {
	if off < 0 {
		off = 0
	}
	if off > len(src) {
		off = len(src)
	}
	line := 0
	col := 0
	i := 0
	for i < off {
		r, size := utf8.DecodeRuneInString(src[i:])
		i += size
		if r == '\n' {
			line++
			col = 0
			continue
		}
		col++
	}
	return lspPosition{Line: line, Character: col}
}

func offsetRangeToLSP(src string, start int, end int) lspRange {
	if end < start {
		end = start
	}
	return lspRange{
		Start: positionFromOffset(src, start),
		End:   positionFromOffset(src, end),
	}
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
	slog.Debug("lsp.rpc.outbound", "payload", string(body))
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
