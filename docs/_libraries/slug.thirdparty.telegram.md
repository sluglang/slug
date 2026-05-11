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
| `telegram-token`  | required                             | Bot token from @BotFather |
| `telegram-url`    | `https://api.telegram.org/bot`       | API base URL            |

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
fn slug.thirdparty.telegram#getUpdates(offset:num, timeout:num = 30):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `offset` | num | — |
| `timeout` | num | `30` |

**Effects:** `net`

---

#### `sendMessage(chatId, text, parseMode)`
```slug
fn slug.thirdparty.telegram#sendMessage(chatId, text, parseMode = "Markdown"):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `chatId` |  | — |
| `text` |  | — |
| `parseMode` |  | `"Markdown"` |

**Effects:** `net`

---

#### `sendTyping(chatId)`
```slug
fn slug.thirdparty.telegram#sendTyping(chatId):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `chatId` |  | — |

**Effects:** `net`