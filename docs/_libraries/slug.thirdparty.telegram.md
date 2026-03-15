---
title: telegram (slug.thirdparty)
---

## slug.thirdparty.telegram

slug.telegram — Telegram Bot API client

Sends messages, typing indicators, and polls for updates via the
Telegram Bot API. All functions return the decoded JSON response
from Telegram, or `nil` on error.

## Configuration

| Key               | Default                              | Description             |
|-------------------|--------------------------------------|-------------------------|
| `telegram_token`  | required                             | Bot token from @BotFather |
| `telegram_url`    | `https://api.telegram.org/bot`       | API base URL            |

The module will print a warning at load time if `telegram_token` is
not set but will not exit — callers will receive `nil` responses until
a valid token is configured.

## Quick start

```slug
val { sendMessage, getUpdates } = import("slug.telegram")

// send a message
sendMessage(chatId, "Hello from Slug!")

// poll for new messages
val response = getUpdates(0)
val updates  = response['result']
```

## Long polling

Use `getUpdates` with an `offset` equal to the last `update_id + 1`
to avoid receiving the same update twice. The `timeout` parameter
controls how long Telegram holds the connection open waiting for new
updates (default 30 seconds).

## Markdown rendering

`sendMessage` defaults to `parseMode: "Markdown"` which renders bold,
italic, inline code, and code blocks from the message text. Pass
`parseMode: nil` to send plain text.

@effects('net')

### TOC

- [`getUpdates(offset, timeout)`](#getupdatesoffset-timeout)
- [`sendMessage(chatId, text, parseMode)`](#sendmessagechatid-text-parsemode)
- [`sendTyping(chatId)`](#sendtypingchatid)

### Functions

#### `getUpdates(offset, timeout)`
```slug
fn slug.thirdparty.telegram#getUpdates(@num offset, @num timeout = 30) -> ?
```


polls Telegram for new updates starting from `offset`.

Uses long polling — Telegram holds the connection open for up to
`timeout` seconds waiting for new updates before returning an empty
list. Set `offset` to the last received `update_id + 1` to acknowledge
processed updates and avoid receiving them again.

Returns the decoded Telegram response map `{ ok, result: [...] }`,
or `nil` on error.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `offset` | @num  | — |
| `timeout` | @num  | `30` |

**Effects:** `net`

---

#### `sendMessage(chatId, text, parseMode)`
```slug
fn slug.thirdparty.telegram#sendMessage(chatId, text, parseMode = "Markdown") -> ?
```


sends a text message to a Telegram chat.

`parseMode` controls text formatting; defaults to `"Markdown"` which
renders bold (`**text**`), code (`` `text` ``), and code blocks.
Pass `nil` to disable formatting. Returns the Telegram API response
map, or `nil` on error.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `chatId` |  | — |
| `text` |  | — |
| `parseMode` |  | `"Markdown"` |

**Effects:** `net`

---

#### `sendTyping(chatId)`
```slug
fn slug.thirdparty.telegram#sendTyping(chatId) -> ?
```


sends a `typing` chat action to show the bot is composing a reply.

The indicator displays for ~5 seconds or until the next message is
sent. Call just before a slow operation to set user expectations.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `chatId` |  | — |

**Effects:** `net`