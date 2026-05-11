---
title: channel (slug)
---

## slug.channel

slug.channel — channel-based concurrency primitives

Provides buffered and unbuffered channels for communicating between
concurrent tasks, following Slug's structured concurrency model.

Channels are created with `chan()` and closed with `close()`. Use `send`
and `recv` for blocking operations, or `trySend`/`tryRecv` for
non-blocking variants. Use `await` to block until a spawned task completes.

## Example

```slug
val { chan, send, recv, close } = import("slug.channel")

val ch = chan()

nursery {
  spawn { send(ch, "hello") }
  val msg = recv(ch)
  println(msg)  // => "hello"
}

close(ch)
```

## Timeouts

Both `recv` and `await` accept an optional `timeout` in milliseconds.
A timeout of `0` (default) means wait indefinitely. When a timeout
elapses a `TimeoutError` is thrown.

### TOC

- [`await(handle, timeout)`](#awaithandle-timeout)
- [`chan(capacity)`](#chancapacity)
- [`close(channel)`](#closechannel)
- [`recv(channel, timeout)`](#recvchannel-timeout)
- [`send(channel, payload)`](#sendchannel-payload)
- [`tryRecv(channel)`](#tryrecvchannel)
- [`trySend(channel, payload)`](#trysendchannel-payload)

### Functions

#### `await(handle, timeout)`
```slug
fn slug.channel#await(handle:task, timeout:num = 0):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | task | — |
| `timeout` | num | `0` |

**Throws:** `Error{type:TimeoutError}`

---

#### `chan(capacity)`
```slug
fn slug.channel#chan(capacity:num = 0):chan
```

| Parameter | Type | Default |
| --- | --- | --- |
| `capacity` | num | `0` |

---

#### `close(channel)`
```slug
fn slug.channel#close(channel:chan):nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | chan | — |

---

#### `recv(channel, timeout)`
```slug
fn slug.channel#recv(channel:chan, timeout:num = 0):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | chan | — |
| `timeout` | num | `0` |

**Throws:** `Error{type:TimeoutError}`

---

#### `send(channel, payload)`
```slug
fn slug.channel#send(channel:chan, payload):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | chan | — |
| `payload` |  | — |

---

#### `tryRecv(channel)`
```slug
fn slug.channel#tryRecv(channel:chan):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | chan | — |

---

#### `trySend(channel, payload)`
```slug
fn slug.channel#trySend(channel:chan, payload):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | chan | — |
| `payload` |  | — |