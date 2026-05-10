package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slug/internal/ast"
	"slug/internal/lexer"
	"slug/internal/parser"
	"slug/internal/token"
	"sort"
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

type rpcSuccessResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result"`
}

type rpcErrorResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Error   *rpcError   `json:"error"`
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
	CodeActionProvider        bool                    `json:"codeActionProvider,omitempty"`
	CompletionProvider        *completionProvider     `json:"completionProvider,omitempty"`
	SignatureHelpProvider     *signatureHelpProvider  `json:"signatureHelpProvider,omitempty"`
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

type codeActionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Range   lspRange `json:"range"`
	Context struct {
		Only []string `json:"only"`
	} `json:"context"`
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

type signatureHelpProvider struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
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

type lspCodeAction struct {
	Title string            `json:"title"`
	Kind  string            `json:"kind,omitempty"`
	Edit  *lspWorkspaceEdit `json:"edit,omitempty"`
	// Internal ranking metadata; not part of LSP payload.
	RankGroup string `json:"-"`
}

type lspSignatureHelp struct {
	Signatures      []lspSignatureInformation `json:"signatures"`
	ActiveSignature int                       `json:"activeSignature"`
	ActiveParameter int                       `json:"activeParameter"`
}

type lspSignatureInformation struct {
	Label         string                    `json:"label"`
	Parameters    []lspParameterInformation `json:"parameters,omitempty"`
	Documentation *lspMarkupContent         `json:"documentation,omitempty"`
}

type lspParameterInformation struct {
	Label         string            `json:"label"`
	Documentation *lspMarkupContent `json:"documentation,omitempty"`
}

type importEditPlan struct {
	Range   lspRange
	NewText string
	Mode    string
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

type moduleObjectBinding struct {
	LocalName    string
	SourceModule string
}

type moduleSymbolIdentity struct {
	Module string
	Name   string
	Kind   string
}

type functionSignature struct {
	Name       string
	Params     []string
	Detail     string
	ParamDocs  []string
	ScopeDepth int
	Start      int
	End        int
}

type memberIdentityHit struct {
	Name       string
	Range      lspRange
	Candidates []string
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
		"textDocument/codeAction":        s.handleCodeAction,
		"textDocument/signatureHelp":     s.handleSignatureHelp,
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
		CodeActionProvider:        true,
		CompletionProvider:        &completionProvider{ResolveProvider: true},
		SignatureHelpProvider:     &signatureHelpProvider{TriggerCharacters: []string{"(", ","}},
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
	if strings.TrimSpace(sym.Detail) == "" || sym.Kind == "variable" {
		if imported, ok := s.resolveCompletionImportedSymbol(normURI, doc.Text, name); ok {
			if strings.TrimSpace(sym.Detail) == "" {
				sym.Detail = imported.Detail
			}
			if sym.Kind == "variable" && strings.TrimSpace(imported.Kind) != "" {
				sym.Kind = imported.Kind
			}
		}
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
	if identity, _, ok, _ := s.resolveModuleMemberIdentityAtOffset(doc.Text, offset); ok {
		if loc, ok := s.resolveModuleExportLocation(normURI, identity.Module, identity.Name); ok {
			return s.writeResult(req.ID, loc)
		}
	}
	name, _, _ := identifierAtOffset(doc.Text, offset)
	if name == "" {
		return s.writeResult(req.ID, nil)
	}
	syms := collectSymbols(doc.Text)
	sym, found := resolveSymbolAt(name, offset, syms)
	if found {
		if sym.ScopeDepth == 0 {
			allBindings := collectImportBindingsForModule(doc.Text, "")
			for _, b := range allBindings {
				if b.LocalName != name {
					continue
				}
				if loc, ok := s.resolveModuleExportLocation(normURI, b.SourceModule, b.SourceName); ok {
					return s.writeResult(req.ID, loc)
				}
			}
		}
		defRange := offsetRangeToLSP(doc.Text, sym.Start, sym.End)
		loc := lspLocation{
			URI:   normURI,
			Range: defRange,
		}
		return s.writeResult(req.ID, loc)
	}
	for _, module := range collectWildcardImportModules(doc.Text) {
		if loc, ok := s.resolveModuleExportLocation(normURI, module, name); ok {
			return s.writeResult(req.ID, loc)
		}
	}
	return s.writeResult(req.ID, nil)
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
	if identity, _, ok, _ := s.resolveModuleMemberIdentityAtOffset(doc.Text, offset); ok {
		refs := s.collectReferencesForIdentityAcrossOpenDocs(normURI, identity, p.Context.IncludeDeclaration)
		return s.writeResult(req.ID, refs)
	}
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
	if identity, rng, ok, ambiguous := s.resolveModuleMemberIdentityAtOffset(doc.Text, offset); ok {
		return s.writeResult(req.ID, prepareRenameResult{Range: rng, Placeholder: identity.Name})
	} else if ambiguous {
		return s.writeResult(req.ID, nil)
	}
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
	if identity, _, ok, ambiguous := s.resolveModuleMemberIdentityAtOffset(doc.Text, offset); ok {
		refs := s.collectReferencesForIdentityAcrossOpenDocs(normURI, identity, true)
		changes := map[string][]lspTextEdit{}
		for _, ref := range refs {
			changes[ref.URI] = append(changes[ref.URI], lspTextEdit{
				Range:   ref.Range,
				NewText: p.NewName,
			})
		}
		return s.writeResult(req.ID, &lspWorkspaceEdit{Changes: changes})
	} else if ambiguous {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", "ambiguous import member; cannot safely rename")
	}
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

func (s *Server) handleCodeAction(req rpcRequest) error {
	var p codeActionParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.writeError(idOrNil(req.ID), -32602, "Invalid params", err.Error())
	}
	normURI, _ := normalizeURI(p.TextDocument.URI)
	doc, ok := s.docs[normURI]
	if !ok {
		return s.writeResult(req.ID, []lspCodeAction{})
	}
	offset := offsetFromPosition(doc.Text, p.Range.Start.Line, p.Range.Start.Character)
	name, start, end := identifierAtOffset(doc.Text, offset)
	if name == "" {
		return s.writeResult(req.ID, []lspCodeAction{})
	}
	syms := collectSymbols(doc.Text)
	if _, found := resolveSymbolAt(name, offset, syms); found {
		return s.writeResult(req.ID, []lspCodeAction{})
	}
	preferSource := prefersSourceActionsOnly(p.Context.Only)
	quickFixAllowed := len(p.Context.Only) == 0 || containsCodeActionKind(p.Context.Only, "quickfix")
	importKind := "quickfix"
	if preferSource {
		importKind = "source.organizeImports"
	}
	actions := make([]lspCodeAction, 0, 8)

	// Prefer local qualification quick-fixes when a module-object import alias exists.
	if quickFixAllowed {
		for _, b := range collectImportObjectBindings(doc.Text, "") {
			if !s.moduleMayExportName(normURI, b.SourceModule, name) {
				continue
			}
			actions = append(actions, lspCodeAction{
				Title:     fmt.Sprintf("Qualify with '%s.%s'", b.LocalName, name),
				Kind:      "quickfix",
				RankGroup: "qualify",
				Edit: &lspWorkspaceEdit{
					Changes: map[string][]lspTextEdit{
						normURI: {{
							Range:   offsetRangeToLSP(doc.Text, start, end),
							NewText: b.LocalName + "." + name,
						}},
					},
				},
			})
		}
	}

	modules := s.discoverExportingModules(normURI, name)
	if len(modules) == 0 {
		return s.writeResult(req.ID, rankAndDedupeCodeActions(actions))
	}
	for _, module := range modules {
		plan := buildImportEditPlan(doc.Text, module, name)
		title := fmt.Sprintf("Add import for '%s' from '%s'", name, module)
		if plan.Mode == "extend" {
			title = fmt.Sprintf("Extend import from '%s' with '%s'", module, name)
		}
		actions = append(actions, lspCodeAction{
			Title:     title,
			Kind:      importKind,
			RankGroup: plan.Mode,
			Edit: &lspWorkspaceEdit{
				Changes: map[string][]lspTextEdit{
					normURI: {{
						Range:   plan.Range,
						NewText: plan.NewText,
					}},
				},
			},
		})
	}
	return s.writeResult(req.ID, rankAndDedupeCodeActions(actions))
}

func containsCodeActionKind(only []string, kind string) bool {
	for _, k := range only {
		if k == kind || strings.HasPrefix(kind, k+".") || strings.HasPrefix(k, kind+".") {
			return true
		}
	}
	return false
}

func prefersSourceActionsOnly(only []string) bool {
	if len(only) == 0 {
		return false
	}
	hasSource := false
	for _, k := range only {
		if k == "source" || strings.HasPrefix(k, "source.") {
			hasSource = true
			continue
		}
		if k == "quickfix" || strings.HasPrefix(k, "quickfix.") {
			return false
		}
	}
	return hasSource
}

func (s *Server) handleSignatureHelp(req rpcRequest) error {
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
	calleeExpr, activeParam, ok := findCallContext(doc.Text, offset)
	if !ok {
		return s.writeResult(req.ID, nil)
	}
	sig, ok := s.resolveSignatureForCallee(normURI, doc.Text, calleeExpr, offset)
	if !ok {
		return s.writeResult(req.ID, nil)
	}
	label := sig.Name + "(" + strings.Join(sig.Params, ", ") + ")"
	params := make([]lspParameterInformation, 0, len(sig.Params))
	for i, pn := range sig.Params {
		p := lspParameterInformation{Label: pn}
		if i < len(sig.ParamDocs) && strings.TrimSpace(sig.ParamDocs[i]) != "" {
			p.Documentation = &lspMarkupContent{Kind: "markdown", Value: sig.ParamDocs[i]}
		}
		params = append(params, p)
	}
	if activeParam < 0 {
		activeParam = 0
	}
	if activeParam >= len(params) && len(params) > 0 {
		activeParam = len(params) - 1
	}
	help := lspSignatureHelp{
		Signatures: []lspSignatureInformation{{
			Label:      label,
			Parameters: params,
		}},
		ActiveSignature: 0,
		ActiveParameter: activeParam,
	}
	if strings.TrimSpace(sig.Detail) != "" {
		help.Signatures[0].Documentation = &lspMarkupContent{Kind: "markdown", Value: sig.Detail}
	}
	return s.writeResult(req.ID, help)
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
	if (!found || strings.TrimSpace(best.Detail) == "") && label != "" {
		if imported, ok := s.resolveCompletionImportedSymbol(uri, doc.Text, label); ok {
			if !found {
				best = imported
				found = true
			} else {
				if strings.TrimSpace(best.Detail) == "" {
					best.Detail = imported.Detail
				}
				if strings.TrimSpace(best.Kind) == "" || best.Kind == "variable" {
					best.Kind = imported.Kind
				}
			}
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

func (s *Server) resolveCompletionImportedSymbol(originURI string, src string, localName string) (symbolDef, bool) {
	for _, b := range collectImportBindingsForModule(src, "") {
		if b.LocalName != localName {
			continue
		}
		if sym, ok := s.resolveModuleExportSymbolInfo(originURI, b.SourceModule, b.SourceName); ok {
			return sym, true
		}
	}
	for _, module := range collectWildcardImportModules(src) {
		if sym, ok := s.resolveModuleExportSymbolInfo(originURI, module, localName); ok {
			return sym, true
		}
	}
	return symbolDef{}, false
}

func (s *Server) resolveModuleExportSymbolInfo(originURI string, module string, name string) (symbolDef, bool) {
	var src string
	if docURI, ok := s.findOpenDocURIByModule(module); ok {
		src = s.docs[docURI].Text
	} else {
		candidates := modulePathCandidatesFromURI(originURI, module)
		if len(candidates) == 0 {
			return symbolDef{}, false
		}
		loaded := false
		for _, path := range candidates {
			b, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			src = string(b)
			loaded = true
			break
		}
		if !loaded {
			return symbolDef{}, false
		}
	}
	syms := collectSymbols(src)
	sigs := collectFunctionSignatures(src)
	isFn := map[string]bool{}
	for _, sig := range sigs {
		if sig.ScopeDepth == 0 {
			isFn[sig.Name] = true
		}
	}
	best := symbolDef{}
	found := false
	for _, sym := range syms {
		if sym.ScopeDepth != 0 || sym.Name != name {
			continue
		}
		if !found || sym.Start < best.Start {
			best = sym
			found = true
		}
	}
	if found {
		if isFn[best.Name] {
			best.Kind = "function"
		}
		return best, true
	}
	return symbolDef{}, false
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
		case *ast.ForeignFunctionDeclaration:
			if s.Name != nil {
				start := s.Name.Token.Position
				end := start + len(s.Name.Token.Literal)
				detail := buildFunctionDocMarkdown(s.Name.Value, s.Doc, s.HasDoc, s.Tags)
				syms = append(syms, symbolDef{Name: s.Name.Value, Kind: "function", Detail: detail, Start: start, End: end, ScopeDepth: scopeDepth})
			}
		}
	}

	walkExpr = func(ex ast.Expression) {
		switch e := ex.(type) {
		case *ast.VarExpression:
			detail := ""
			if _, ok := e.Value.(*ast.FunctionLiteral); ok {
				for _, n := range topLevelPatternNames(e.Pattern) {
					detail = buildFunctionDocMarkdown(n.Name, e.Doc, e.HasDoc, e.Tags)
					addPattern(e.Pattern, "variable", detail)
					goto walkVarValue
				}
			}
			if e.HasDoc {
				detail = strings.TrimSpace(e.Doc)
			}
			addPattern(e.Pattern, "variable", detail)
		walkVarValue:
			walkExpr(e.Value)
		case *ast.ValExpression:
			detail := ""
			if _, ok := e.Value.(*ast.FunctionLiteral); ok {
				for _, n := range topLevelPatternNames(e.Pattern) {
					detail = buildFunctionDocMarkdown(n.Name, e.Doc, e.HasDoc, e.Tags)
					addPattern(e.Pattern, "variable", detail)
					goto walkValValue
				}
			}
			if e.HasDoc {
				detail = strings.TrimSpace(e.Doc)
			}
			addPattern(e.Pattern, "variable", detail)
		walkValValue:
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

	refs := s.collectReferencesForIdentityAcrossOpenDocs(originURI, identity, includeDeclaration)
	refs = append(refs, out...)
	return dedupeLocations(refs)
}

func (s *Server) collectReferencesForIdentityAcrossOpenDocs(originURI string, identity moduleSymbolIdentity, includeDeclaration bool) []lspLocation {
	out := []lspLocation{}
	for uri, doc := range s.docs {
		if uri == originURI {
			// keep scanning origin too for importer alias/member references
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
		objBindings := collectImportObjectBindings(doc.Text, identity.Module)
		memberRefs := collectMemberReferencesForAliases(doc.Text, uri, objBindings, identity.Name)
		out = append(out, memberRefs...)
		inlineRefs := collectInlineImportMemberReferences(doc.Text, uri, identity.Module, identity.Name)
		out = append(out, inlineRefs...)
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

func collectImportObjectBindings(src string, onlyModule string) []moduleObjectBinding {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	out := []moduleObjectBinding{}

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
		for _, n := range topLevelPatternNames(pattern) {
			out = append(out, moduleObjectBinding{
				LocalName:    n.Name,
				SourceModule: module,
			})
		}
	}
	return out
}

func collectWildcardImportModules(src string) []string {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	out := []string{}
	seen := map[string]bool{}
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
		mp, ok := pattern.(*ast.MapPattern)
		if !ok || mp == nil || !mp.SelectAll {
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
		mods := importModuleArgs(call.Arguments)
		for _, m := range mods {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
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

func modulePathCandidatesFromURI(originURI string, module string) []string {
	if module == "" {
		return nil
	}
	_, local := normalizeURI(originURI)
	if local == "" {
		return nil
	}
	pathParts := strings.Split(module, ".")
	relPath := filepath.Join(pathParts...) + ".slug"

	candidates := []string{
		filepath.Join(filepath.Dir(local), relPath),
	}
	if slugHome := strings.TrimSpace(os.Getenv("SLUG_HOME")); slugHome != "" {
		candidates = append(candidates, filepath.Join(slugHome, "lib", relPath))
	}
	return candidates
}

func (s *Server) resolveModuleExportLocation(originURI string, module string, name string) (lspLocation, bool) {
	var src string
	var uri string
	if docURI, ok := s.findOpenDocURIByModule(module); ok {
		doc := s.docs[docURI]
		src = doc.Text
		uri = doc.URI
	} else {
		candidates := modulePathCandidatesFromURI(originURI, module)
		if len(candidates) == 0 {
			return lspLocation{}, false
		}
		var loaded bool
		for _, path := range candidates {
			b, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			src = string(b)
			uri, _ = normalizeURI(path)
			loaded = true
			break
		}
		if !loaded {
			return lspLocation{}, false
		}
	}
	for _, exp := range collectExportedTopLevelSymbols(src) {
		if exp.Name != name {
			continue
		}
		return lspLocation{
			URI:   uri,
			Range: offsetRangeToLSP(src, exp.Start, exp.End),
		}, true
	}
	return lspLocation{}, false
}

func (s *Server) resolveModuleExportSignature(originURI string, module string, name string) (functionSignature, bool) {
	var src string
	if docURI, ok := s.findOpenDocURIByModule(module); ok {
		src = s.docs[docURI].Text
	} else {
		candidates := modulePathCandidatesFromURI(originURI, module)
		if len(candidates) == 0 {
			return functionSignature{}, false
		}
		var loaded bool
		for _, path := range candidates {
			b, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			src = string(b)
			loaded = true
			break
		}
		if !loaded {
			return functionSignature{}, false
		}
	}
	sigs := collectFunctionSignatures(src)
	for _, sig := range sigs {
		if sig.Name == name && sig.ScopeDepth == 0 {
			return sig, true
		}
	}
	return functionSignature{}, false
}

func (s *Server) findOpenDocURIByModule(module string) (string, bool) {
	for uri := range s.docs {
		if moduleNameFromURI(uri) == module {
			return uri, true
		}
	}
	return "", false
}

func collectFunctionSignatures(src string) []functionSignature {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	out := []functionSignature{}
	scopeDepth := 0

	var walkStmt func(ast.Statement)
	var walkExpr func(ast.Expression)
	add := func(name string, params []*ast.FunctionParameter, detail string, start int, end int) {
		pn := make([]string, 0, len(params))
		paramDocsByName := parseParamDocs(detail)
		pdocs := make([]string, 0, len(params))
		for _, p := range params {
			if p == nil || p.Name == nil {
				continue
			}
			pn = append(pn, p.Name.Value)
			pdocs = append(pdocs, paramDocsByName[p.Name.Value])
		}
		out = append(out, functionSignature{Name: name, Params: pn, Detail: detail, ParamDocs: pdocs, ScopeDepth: scopeDepth, Start: start, End: end})
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
			for _, c := range s.Statements {
				walkStmt(c)
			}
			scopeDepth--
		case *ast.ForeignFunctionDeclaration:
			if s.Name != nil {
				start := s.Name.Token.Position
				end := start + len(s.Name.Token.Literal)
				detail := buildFunctionDocMarkdown(s.Name.Value, s.Doc, s.HasDoc, s.Tags)
				add(s.Name.Value, s.Parameters, detail, start, end)
			}
		}
	}

	walkExpr = func(ex ast.Expression) {
		switch e := ex.(type) {
		case *ast.ValExpression:
			if fn, ok := e.Value.(*ast.FunctionLiteral); ok {
				for _, n := range topLevelPatternNames(e.Pattern) {
					detail := buildFunctionDocMarkdown(n.Name, e.Doc, e.HasDoc, e.Tags)
					add(n.Name, fn.Parameters, detail, n.Start, n.End)
				}
			}
			walkExpr(e.Value)
		case *ast.VarExpression:
			if fn, ok := e.Value.(*ast.FunctionLiteral); ok {
				for _, n := range topLevelPatternNames(e.Pattern) {
					detail := buildFunctionDocMarkdown(n.Name, e.Doc, e.HasDoc, e.Tags)
					add(n.Name, fn.Parameters, detail, n.Start, n.End)
				}
			}
			walkExpr(e.Value)
		case *ast.FunctionLiteral:
			scopeDepth++
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
		case *ast.CallExpression:
			walkExpr(e.Function)
			for _, a := range e.Arguments {
				walkExpr(a)
			}
		case *ast.IndexExpression:
			walkExpr(e.Left)
			walkExpr(e.Index)
		}
	}

	for _, st := range program.Statements {
		walkStmt(st)
	}
	return out
}

func resolveFunctionAt(name string, useOffset int, defs []functionSignature) (functionSignature, bool) {
	best := functionSignature{}
	found := false
	for _, d := range defs {
		if d.Name != name || d.Start > useOffset {
			continue
		}
		if !found || d.ScopeDepth > best.ScopeDepth || (d.ScopeDepth == best.ScopeDepth && d.Start > best.Start) {
			best = d
			found = true
		}
	}
	return best, found
}

func findCallContext(src string, off int) (calleeExpr string, activeParam int, ok bool) {
	if off < 0 {
		return "", 0, false
	}
	if off > len(src) {
		off = len(src)
	}
	depth := 0
	inString := byte(0)
	open := -1
	for i := off - 1; i >= 0; i-- {
		ch := src[i]
		if inString != 0 {
			if ch == inString && !isEscapedAt(src, i) {
				inString = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inString = ch
			continue
		}
		switch ch {
		case ')', ']', '}':
			depth++
		case '(':
			if depth == 0 {
				open = i
				i = -1
				break
			}
			depth--
		case '[', '{':
			if depth > 0 {
				depth--
			}
		}
	}
	if open < 0 {
		return "", 0, false
	}
	j := open - 1
	for j >= 0 && (src[j] == ' ' || src[j] == '\t' || src[j] == '\n' || src[j] == '\r') {
		j--
	}
	if j < 0 {
		return "", 0, false
	}
	start := j
	for start >= 0 {
		b := src[start]
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_' || b == '.' || b == '"' || b == '\'' || b == '(' || b == ')' {
			start--
			continue
		}
		break
	}
	start++
	if start > j {
		return "", 0, false
	}
	callee := strings.TrimSpace(src[start : j+1])
	if callee == "" {
		return "", 0, false
	}
	param := 0
	parenDepth := 0
	bracketDepth := 0
	braceDepth := 0
	inString = 0
	escaped := false
	for i := open + 1; i < off && i < len(src); i++ {
		ch := src[i]
		if inString != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == inString {
				inString = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inString = ch
			continue
		}
		switch ch {
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case '{':
			braceDepth++
		case '}':
			if braceDepth > 0 {
				braceDepth--
			}
		case ',':
			if parenDepth == 0 && bracketDepth == 0 && braceDepth == 0 {
				param++
			}
		}
	}
	return callee, param, true
}

func parseParamDocs(doc string) map[string]string {
	out := map[string]string{}
	if strings.TrimSpace(doc) == "" {
		return out
	}
	lines := strings.Split(doc, "\n")
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if !strings.HasPrefix(s, "@param") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(s, "@param"))
		if rest == "" {
			continue
		}
		parts := strings.Fields(rest)
		if len(parts) == 0 {
			continue
		}
		name := parts[0]
		desc := ""
		if len(parts) > 1 {
			desc = strings.TrimSpace(rest[len(name):])
		}
		if desc != "" {
			out[name] = desc
		}
	}
	return out
}

func buildFunctionDocMarkdown(name string, doc string, hasDoc bool, tags []*ast.Tag) string {
	base := ""
	if hasDoc {
		base = strings.TrimSpace(doc)
	}
	examples := renderTestWithExamplesMarkdown(name, tags)
	switch {
	case base == "":
		return examples
	case examples == "":
		return base
	default:
		return base + examples
	}
}

func renderTestWithExamplesMarkdown(name string, tags []*ast.Tag) string {
	args := tagArgs(tags, "@testWith")
	if len(args) < 2 {
		return ""
	}
	lines := make([]string, 0, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		inputsExpr := args[i]
		expectedExpr := args[i+1]
		inputs := []ast.Expression{inputsExpr}
		if lst, ok := inputsExpr.(*ast.ListLiteral); ok {
			inputs = lst.Elements
		}
		argParts := make([]string, 0, len(inputs))
		for _, in := range inputs {
			argParts = append(argParts, renderTestWithValue(in))
		}
		lines = append(lines, fmt.Sprintf("%s(%s)  // => %s", name, strings.Join(argParts, ", "), renderTestWithValue(expectedExpr)))
	}
	if len(lines) == 0 {
		return ""
	}
	return "\n\n#### Examples\n\n```slug\n" + strings.Join(lines, "\n") + "\n```"
}

func tagArgs(tags []*ast.Tag, name string) []ast.Expression {
	out := make([]ast.Expression, 0, 4)
	for _, t := range tags {
		if t == nil || t.Name != name || len(t.Args) == 0 {
			continue
		}
		out = append(out, t.Args...)
	}
	return out
}

func renderTestWithValue(v ast.Expression) string {
	switch x := v.(type) {
	case *ast.StringLiteral:
		return strconv.Quote(x.Value)
	case *ast.ListLiteral:
		parts := make([]string, 0, len(x.Elements))
		for _, el := range x.Elements {
			parts = append(parts, renderTestWithValue(el))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *ast.Nil, *ast.Boolean:
		return x.String()
	default:
		return x.String()
	}
}

func isEscapedAt(src string, idx int) bool {
	if idx <= 0 || idx >= len(src) {
		return false
	}
	slashes := 0
	for i := idx - 1; i >= 0 && src[i] == '\\'; i-- {
		slashes++
	}
	return slashes%2 == 1
}

func (s *Server) resolveSignatureForCallee(originURI string, src string, callee string, callOffset int) (functionSignature, bool) {
	callee = strings.TrimSpace(callee)
	if callee == "" {
		return functionSignature{}, false
	}
	if strings.Contains(callee, ".") {
		parts := strings.Split(callee, ".")
		member := parts[len(parts)-1]
		prefix := strings.Join(parts[:len(parts)-1], ".")
		if prefix == "import(\"slug.std\")" || strings.HasPrefix(prefix, "import(") {
			mods := parseInlineImportModules(prefix)
			for _, m := range mods {
				if sig, ok := s.resolveModuleExportSignature(originURI, m, member); ok {
					return sig, true
				}
			}
		}
		if mod, ok := moduleForAliasInDoc(src, prefix); ok {
			if sig, ok := s.resolveModuleExportSignature(originURI, mod, member); ok {
				return sig, true
			}
		}
	}
	name := callee
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}
	defs := collectFunctionSignatures(src)
	if sig, ok := resolveFunctionAt(name, callOffset, defs); ok {
		if sig.ScopeDepth == 0 {
			for _, b := range collectImportBindingsForModule(src, "") {
				if b.LocalName == name {
					if imp, ok := s.resolveModuleExportSignature(originURI, b.SourceModule, b.SourceName); ok {
						return imp, true
					}
				}
			}
			for _, mod := range collectWildcardImportModules(src) {
				if imp, ok := s.resolveModuleExportSignature(originURI, mod, name); ok {
					return imp, true
				}
			}
		}
		return sig, true
	}
	for _, b := range collectImportBindingsForModule(src, "") {
		if b.LocalName == name {
			if imp, ok := s.resolveModuleExportSignature(originURI, b.SourceModule, b.SourceName); ok {
				return imp, true
			}
		}
	}
	for _, mod := range collectWildcardImportModules(src) {
		if imp, ok := s.resolveModuleExportSignature(originURI, mod, name); ok {
			return imp, true
		}
	}
	return functionSignature{}, false
}

func parseInlineImportModules(expr string) []string {
	expr = strings.TrimSpace(expr)
	if !strings.HasPrefix(expr, "import(") || !strings.HasSuffix(expr, ")") {
		return nil
	}
	inside := strings.TrimSpace(expr[len("import(") : len(expr)-1])
	if inside == "" {
		return nil
	}
	parts := strings.Split(inside, ",")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"'")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func moduleForAliasInDoc(src string, alias string) (string, bool) {
	for _, b := range collectImportObjectBindings(src, "") {
		if b.LocalName == alias {
			return b.SourceModule, true
		}
	}
	return "", false
}

func (s *Server) resolveModuleMemberIdentityAtOffset(src string, off int) (moduleSymbolIdentity, lspRange, bool, bool) {
	hits := collectModuleMemberHitsAtOffset(src, off)
	if len(hits) == 0 {
		return moduleSymbolIdentity{}, lspRange{}, false, false
	}
	hit := hits[0]
	if len(hit.Candidates) == 0 {
		return moduleSymbolIdentity{}, lspRange{}, false, false
	}
	if len(hit.Candidates) == 1 {
		return moduleSymbolIdentity{Module: hit.Candidates[0], Name: hit.Name, Kind: "variable"}, hit.Range, true, false
	}
	modules := s.modulesExportingName(hit.Candidates, hit.Name)
	if len(modules) == 1 {
		return moduleSymbolIdentity{Module: modules[0], Name: hit.Name, Kind: "variable"}, hit.Range, true, false
	}
	return moduleSymbolIdentity{}, hit.Range, false, true
}

func collectModuleMemberHitsAtOffset(src string, off int) []memberIdentityHit {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	objBindings := collectImportObjectBindings(src, "")
	aliasToModule := map[string]string{}
	for _, b := range objBindings {
		aliasToModule[b.LocalName] = b.SourceModule
	}
	hits := []memberIdentityHit{}

	var walkStmt func(ast.Statement)
	var walkExpr func(ast.Expression)

	walkStmt = func(st ast.Statement) {
		switch s := st.(type) {
		case *ast.ExpressionStatement:
			walkExpr(s.Expression)
		case *ast.ReturnStatement:
			walkExpr(s.ReturnValue)
		case *ast.ThrowStatement:
			walkExpr(s.Value)
		case *ast.BlockStatement:
			for _, c := range s.Statements {
				walkStmt(c)
			}
		}
	}

	walkExpr = func(ex ast.Expression) {
		switch e := ex.(type) {
		case *ast.IndexExpression:
			if e.IsDotLookup {
				if modules, ok := modulesForDotLookupLeft(e.Left, aliasToModule); ok {
					switch k := e.Index.(type) {
					case *ast.SymbolLiteral:
						start := k.Token.Position
						end := start + len(k.Token.Literal)
						if off >= start && off < end {
							hits = append(hits, memberIdentityHit{
								Name:       k.Value,
								Range:      offsetRangeToLSP(src, start, end),
								Candidates: modules,
							})
						}
					case *ast.Identifier:
						start := k.Token.Position
						end := start + len(k.Token.Literal)
						if off >= start && off < end {
							hits = append(hits, memberIdentityHit{
								Name:       k.Value,
								Range:      offsetRangeToLSP(src, start, end),
								Candidates: modules,
							})
						}
					}
				}
			}
			walkExpr(e.Left)
			walkExpr(e.Index)
		case *ast.CallExpression:
			walkExpr(e.Function)
			for _, a := range e.Arguments {
				walkExpr(a)
			}
		case *ast.ValExpression:
			walkExpr(e.Value)
		case *ast.VarExpression:
			walkExpr(e.Value)
		case *ast.IfExpression:
			walkExpr(e.Condition)
			if e.ThenBranch != nil {
				walkStmt(e.ThenBranch)
			}
			if e.ElseBranch != nil {
				walkStmt(e.ElseBranch)
			}
		}
	}

	for _, st := range program.Statements {
		walkStmt(st)
	}
	return hits
}

func modulesForDotLookupLeft(left ast.Expression, aliasToModule map[string]string) ([]string, bool) {
	if ident, ok := left.(*ast.Identifier); ok {
		module, ok := aliasToModule[ident.Value]
		if !ok {
			return nil, false
		}
		return []string{module}, true
	}
	call, ok := left.(*ast.CallExpression)
	if !ok {
		return nil, false
	}
	fn, ok := call.Function.(*ast.Identifier)
	if !ok || fn.Value != "import" {
		return nil, false
	}
	modules := importModuleArgs(call.Arguments)
	if len(modules) < 1 {
		return nil, false
	}
	return modules, true
}

func collectMemberReferencesForAliases(src string, uri string, aliases []moduleObjectBinding, member string) []lspLocation {
	if len(aliases) == 0 || member == "" {
		return nil
	}
	aliasSet := map[string]bool{}
	for _, a := range aliases {
		aliasSet[a.LocalName] = true
	}
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	out := []lspLocation{}
	seen := map[string]bool{}

	var walkStmt func(ast.Statement)
	var walkExpr func(ast.Expression)

	walkStmt = func(st ast.Statement) {
		switch s := st.(type) {
		case *ast.ExpressionStatement:
			walkExpr(s.Expression)
		case *ast.ReturnStatement:
			walkExpr(s.ReturnValue)
		case *ast.ThrowStatement:
			walkExpr(s.Value)
		case *ast.BlockStatement:
			for _, c := range s.Statements {
				walkStmt(c)
			}
		}
	}
	walkExpr = func(ex ast.Expression) {
		switch e := ex.(type) {
		case *ast.IndexExpression:
			if left, ok := e.Left.(*ast.Identifier); ok && e.IsDotLookup && aliasSet[left.Value] {
				switch k := e.Index.(type) {
				case *ast.SymbolLiteral:
					if k.Value == member {
						start := k.Token.Position
						end := start + len(k.Token.Literal)
						rng := offsetRangeToLSP(src, start, end)
						key := fmt.Sprintf("%d:%d:%d:%d", rng.Start.Line, rng.Start.Character, rng.End.Line, rng.End.Character)
						if !seen[key] {
							seen[key] = true
							out = append(out, lspLocation{URI: uri, Range: rng})
						}
					}
				case *ast.Identifier:
					if k.Value == member {
						start := k.Token.Position
						end := start + len(k.Token.Literal)
						rng := offsetRangeToLSP(src, start, end)
						key := fmt.Sprintf("%d:%d:%d:%d", rng.Start.Line, rng.Start.Character, rng.End.Line, rng.End.Character)
						if !seen[key] {
							seen[key] = true
							out = append(out, lspLocation{URI: uri, Range: rng})
						}
					}
				}
			}
			walkExpr(e.Left)
			walkExpr(e.Index)
		case *ast.CallExpression:
			walkExpr(e.Function)
			for _, a := range e.Arguments {
				walkExpr(a)
			}
		case *ast.ValExpression:
			walkExpr(e.Value)
		case *ast.VarExpression:
			walkExpr(e.Value)
		case *ast.IfExpression:
			walkExpr(e.Condition)
			if e.ThenBranch != nil {
				walkStmt(e.ThenBranch)
			}
			if e.ElseBranch != nil {
				walkStmt(e.ElseBranch)
			}
		}
	}
	for _, st := range program.Statements {
		walkStmt(st)
	}
	return out
}

func collectInlineImportMemberReferences(src string, uri string, module string, member string) []lspLocation {
	if module == "" || member == "" {
		return nil
	}
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()
	out := []lspLocation{}
	seen := map[string]bool{}

	var walkStmt func(ast.Statement)
	var walkExpr func(ast.Expression)

	walkStmt = func(st ast.Statement) {
		switch s := st.(type) {
		case *ast.ExpressionStatement:
			walkExpr(s.Expression)
		case *ast.ReturnStatement:
			walkExpr(s.ReturnValue)
		case *ast.ThrowStatement:
			walkExpr(s.Value)
		case *ast.BlockStatement:
			for _, c := range s.Statements {
				walkStmt(c)
			}
		}
	}
	walkExpr = func(ex ast.Expression) {
		switch e := ex.(type) {
		case *ast.IndexExpression:
			if e.IsDotLookup {
				if mods, ok := modulesForDotLookupLeft(e.Left, nil); ok && containsString(mods, module) {
					switch k := e.Index.(type) {
					case *ast.SymbolLiteral:
						if k.Value == member {
							start := k.Token.Position
							end := start + len(k.Token.Literal)
							rng := offsetRangeToLSP(src, start, end)
							key := fmt.Sprintf("%d:%d:%d:%d", rng.Start.Line, rng.Start.Character, rng.End.Line, rng.End.Character)
							if !seen[key] {
								seen[key] = true
								out = append(out, lspLocation{URI: uri, Range: rng})
							}
						}
					case *ast.Identifier:
						if k.Value == member {
							start := k.Token.Position
							end := start + len(k.Token.Literal)
							rng := offsetRangeToLSP(src, start, end)
							key := fmt.Sprintf("%d:%d:%d:%d", rng.Start.Line, rng.Start.Character, rng.End.Line, rng.End.Character)
							if !seen[key] {
								seen[key] = true
								out = append(out, lspLocation{URI: uri, Range: rng})
							}
						}
					}
				}
			}
			walkExpr(e.Left)
			walkExpr(e.Index)
		case *ast.CallExpression:
			walkExpr(e.Function)
			for _, a := range e.Arguments {
				walkExpr(a)
			}
		case *ast.ValExpression:
			walkExpr(e.Value)
		case *ast.VarExpression:
			walkExpr(e.Value)
		case *ast.IfExpression:
			walkExpr(e.Condition)
			if e.ThenBranch != nil {
				walkStmt(e.ThenBranch)
			}
			if e.ElseBranch != nil {
				walkStmt(e.ElseBranch)
			}
		}
	}

	for _, st := range program.Statements {
		walkStmt(st)
	}
	return out
}

func (s *Server) modulesExportingName(candidates []string, name string) []string {
	out := []string{}
	for _, candidate := range candidates {
		if s.moduleExportsName(candidate, name) {
			out = append(out, candidate)
		}
	}
	return out
}

func (s *Server) moduleExportsName(module string, name string) bool {
	for uri, doc := range s.docs {
		if moduleNameFromURI(uri) != module {
			continue
		}
		for _, exp := range collectExportedTopLevelSymbols(doc.Text) {
			if exp.Name == name {
				return true
			}
		}
	}
	return false
}

func containsString(vs []string, target string) bool {
	for _, v := range vs {
		if v == target {
			return true
		}
	}
	return false
}

func importInsertionPosition(src string) lspPosition {
	lines := strings.Split(src, "\n")
	line := 0
	for i, ln := range lines {
		trim := strings.TrimSpace(ln)
		if trim == "" {
			if i == line {
				continue
			}
			break
		}
		if strings.Contains(trim, "import(") && (strings.HasPrefix(trim, "val ") || strings.HasPrefix(trim, "var ")) {
			line = i + 1
			continue
		}
		break
	}
	return lspPosition{Line: line, Character: 0}
}

func buildImportEditPlan(src string, module string, symbol string) importEditPlan {
	if plan, ok := extendExistingImportBinding(src, module, symbol); ok {
		return plan
	}
	insertPos := importInsertionPosition(src)
	return importEditPlan{
		Range:   lspRange{Start: insertPos, End: insertPos},
		NewText: fmt.Sprintf("val { %s } = import(\"%s\")\n", symbol, module),
		Mode:    "insert",
	}
}

func extendExistingImportBinding(src string, module string, symbol string) (importEditPlan, bool) {
	l := lexer.New(src)
	p := parser.New(l, "<lsp>", src)
	program := p.ParseProgram()

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
		mods := importModuleArgs(call.Arguments)
		if len(mods) != 1 || mods[0] != module {
			continue
		}
		mp, ok := pattern.(*ast.MapPattern)
		if !ok || mp == nil {
			continue
		}
		for _, entry := range mp.Pairs {
			key := patternEntryKeyName(entry.Key)
			b := patternBindingName(entry.Pattern)
			if key == symbol || b == symbol {
				return importEditPlan{}, false
			}
		}
		if mp.SelectAll {
			return importEditPlan{}, false
		}
		if len(mp.Pairs) == 0 {
			return importEditPlan{}, false
		}
		last := mp.Pairs[len(mp.Pairs)-1]
		ins := lspPosition{Line: positionFromOffset(src, offsetAfterNode(src, last.Pattern)).Line, Character: positionFromOffset(src, offsetAfterNode(src, last.Pattern)).Character}
		return importEditPlan{
			Range:   lspRange{Start: ins, End: ins},
			NewText: ", " + symbol,
			Mode:    "extend",
		}, true
	}
	return importEditPlan{}, false
}

func rankAndDedupeCodeActions(in []lspCodeAction) []lspCodeAction {
	if len(in) <= 1 {
		return in
	}
	type ranked struct {
		a lspCodeAction
		r int
		k string
	}
	seen := map[string]bool{}
	items := make([]ranked, 0, len(in))
	for _, a := range in {
		k := actionDedupKey(a)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		items = append(items, ranked{a: a, r: actionRank(a), k: k})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].r != items[j].r {
			return items[i].r < items[j].r
		}
		return items[i].a.Title < items[j].a.Title
	})
	out := make([]lspCodeAction, 0, len(items))
	for _, it := range items {
		out = append(out, it.a)
	}
	return out
}

func actionRank(a lspCodeAction) int {
	switch a.RankGroup {
	case "extend":
		return 0
	case "qualify":
		return 1
	case "insert":
		return 2
	}
	if a.Edit == nil {
		return 100
	}
	for _, edits := range a.Edit.Changes {
		for _, e := range edits {
			if strings.Contains(e.NewText, ".") && !strings.HasPrefix(e.NewText, "val {") && !strings.HasPrefix(e.NewText, "var {") {
				return 0
			}
			if strings.HasPrefix(e.NewText, ", ") {
				return 1
			}
			if strings.HasPrefix(e.NewText, "val {") || strings.HasPrefix(e.NewText, "var {") {
				return 2
			}
		}
	}
	return 50
}

func actionDedupKey(a lspCodeAction) string {
	if a.Edit == nil {
		return a.Title
	}
	parts := []string{a.Title}
	for uri, edits := range a.Edit.Changes {
		for _, e := range edits {
			parts = append(parts, fmt.Sprintf("%s:%d:%d:%d:%d:%s", uri, e.Range.Start.Line, e.Range.Start.Character, e.Range.End.Line, e.Range.End.Character, e.NewText))
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

func offsetAfterNode(src string, n ast.Node) int {
	if n == nil {
		return 0
	}
	start := -1
	length := 0
	switch v := n.(type) {
	case *ast.IdentifierPattern:
		if v.Value != nil {
			start = v.Value.Token.Position
			length = len(v.Value.Token.Literal)
		}
	case *ast.BindingPattern:
		if v.Name != nil {
			start = v.Name.Token.Position
			length = len(v.Name.Token.Literal)
		}
	case *ast.Identifier:
		start = v.Token.Position
		length = len(v.Token.Literal)
	case *ast.SymbolLiteral:
		start = v.Token.Position
		length = len(v.Token.Literal)
	default:
		s := n.String()
		if s == "" {
			return 0
		}
		idx := strings.Index(src, s)
		if idx >= 0 {
			start = idx
			length = len(s)
		}
	}
	if start < 0 {
		return 0
	}
	end := start + length
	if end > len(src) {
		end = len(src)
	}
	return end
}

func (s *Server) discoverExportingModules(originURI string, name string) []string {
	seen := map[string]bool{}
	out := []string{}
	for uri, doc := range s.docs {
		mod := moduleNameFromURI(uri)
		if mod == "" {
			continue
		}
		for _, exp := range collectExportedTopLevelSymbols(doc.Text) {
			if exp.Name == name && !seen[mod] {
				seen[mod] = true
				out = append(out, mod)
			}
		}
	}
	slugHome := strings.TrimSpace(os.Getenv("SLUG_HOME"))
	if slugHome == "" {
		return out
	}
	libRoot := filepath.Join(slugHome, "lib")
	_ = filepath.WalkDir(libRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() || !strings.HasSuffix(path, ".slug") {
			return nil
		}
		rel, err := filepath.Rel(libRoot, path)
		if err != nil {
			return nil
		}
		mod := strings.TrimSuffix(filepath.ToSlash(rel), ".slug")
		mod = strings.ReplaceAll(mod, "/", ".")
		if seen[mod] {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, exp := range collectExportedTopLevelSymbols(string(b)) {
			if exp.Name == name {
				seen[mod] = true
				out = append(out, mod)
				break
			}
		}
		return nil
	})
	return out
}

func (s *Server) moduleMayExportName(originURI string, module string, name string) bool {
	if s.moduleExportsName(module, name) {
		return true
	}
	candidates := modulePathCandidatesFromURI(originURI, module)
	for _, path := range candidates {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, exp := range collectExportedTopLevelSymbols(string(b)) {
			if exp.Name == name {
				return true
			}
		}
	}
	return false
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
	if s.isCanceledID(id) {
		return nil
	}
	return writeFramedMessage(s.out, rpcSuccessResponse{JSONRPC: "2.0", ID: idOrNil(id), Result: result})
}

func (s *Server) writeError(id interface{}, code int, message string, data interface{}) error {
	if s.isCanceledAnyID(id) {
		return nil
	}
	return writeFramedMessage(s.out, rpcErrorResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message, Data: data}})
}

func (s *Server) isCanceledID(id json.RawMessage) bool {
	key := cancelIDKey(id)
	if key == "" {
		return false
	}
	if !s.canceledReqs[key] {
		return false
	}
	delete(s.canceledReqs, key)
	return true
}

func (s *Server) isCanceledAnyID(id interface{}) bool {
	key := cancelAnyIDKey(id)
	if key == "" {
		return false
	}
	if !s.canceledReqs[key] {
		return false
	}
	delete(s.canceledReqs, key)
	return true
}

func cancelIDKey(id json.RawMessage) string {
	if len(id) == 0 {
		return ""
	}
	return string(bytes.TrimSpace(id))
}

func cancelAnyIDKey(id interface{}) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case json.RawMessage:
		return cancelIDKey(v)
	case string:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
