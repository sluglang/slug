---
title: ollama (slug.thirdparty)
---

## slug.thirdparty.ollama

slug.ollama — Ollama LLM client

Sends chat requests to a locally running Ollama instance via the
`/api/chat` endpoint and returns the model's reply as a string.

## Configuration

| Key             | Default                    | Description              |
|-----------------|----------------------------|--------------------------|
| `ollama_url`    | `http://localhost:11434`   | Ollama base URL          |
| `ollama_model`  | `llama3`                   | Model name to use        |

## Quick start

```slug
val { ollamaChat } = import("slug.ollama")

val history = [
  { role: "user", content: "What is Slug?" }
]

val reply = ollamaChat(history)
println(reply)
```

## Conversation history

`history` is a list of `{ role, content }` maps in Ollama's message
format. Roles are `"user"`, `"assistant"`, and `"system"`. Build up
history by appending each turn:

```slug
val history1 = history :+ { role: "assistant", content: reply }
val history2 = history1 :+ { role: "user", content: "Tell me more" }
val reply2   = ollamaChat(history2)
```

## System prompt

Pass `systemPrompt` to prepend a system message before every request.
The system message is injected fresh on each call and never stored in
`history`, so it doesn't grow with the conversation.

```slug
ollamaChat(history, systemPrompt: "You are a helpful Slug expert.")
```

@effects('net')

### TOC

- [`ollamaChat(history, systemPrompt, model)`](#ollamachathistory-systemprompt-model)

### Functions

#### `ollamaChat(history, systemPrompt, model)`
```slug
fn slug.thirdparty.ollama#ollamaChat(@list history, @str systemPrompt = nil, @str model = cfg(ollama_model, llama3)) -> ?
```


sends a chat request to Ollama and returns the model's reply text.

`history` is a list of `{ role, content }` maps representing the
conversation so far. Roles are `"user"` and `"assistant"`. If
`systemPrompt` is provided it is prepended as a `"system"` message
on every call without being stored in `history`.

Returns `nil` on error (details printed to stdout).

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `history` | @list  | — |
| `systemPrompt` | @str  | `nil` |
| `model` | @str  | `cfg(ollama_model, llama3)` |

**Effects:** `net`