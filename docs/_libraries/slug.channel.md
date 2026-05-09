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
fn slug.channel#await(@task handle, @num timeout = 0) -> ?
```


waits for a spawned task to complete and returns its result.

If `timeout` is greater than 0 and the task does not complete within
that many milliseconds, a `TimeoutError` is thrown.

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @task  | — |
| `timeout` | @num  | `0` |

**Throws:** `@struct(Error{type:TimeoutError})`

---

#### `chan(capacity)`
```slug
fn slug.channel#chan(@num capacity = 0) -> @chan
```


creates a new channel with an optional buffer capacity.

An unbuffered channel (capacity 0) blocks the sender until a receiver
is ready. A buffered channel allows up to `capacity` messages to be
queued before blocking.

| Parameter | Type | Default |
| --- | --- | --- |
| `capacity` | @num  | `0` |

---

#### `close(channel)`
```slug
fn slug.channel#close(@chan channel) -> nil
```


closes a channel, signalling that no more values will be sent.

Receivers on a closed channel return nil. Sending to a closed channel
is a runtime error.

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |

---

#### `recv(channel, timeout)`
```slug
fn slug.channel#recv(@chan channel, @num timeout = 0) -> ?
```


receives a value from a channel, blocking until one is available.

Returns a payload value when available, or `nil` when the channel is closed and drained.

If `timeout` is greater than 0 and no value arrives within that many
milliseconds, a `TimeoutError` is thrown.

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |
| `timeout` | @num  | `0` |

**Throws:** `@struct(Error{type:TimeoutError})`

---

#### `send(channel, payload)`
```slug
fn slug.channel#send(@chan channel, payload) -> ?
```


sends a value to a channel, blocking until a receiver is ready.

`payload` must not be `nil`.

Returns the channel, allowing chains: `ch /> send("a") /> send("b")`

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |
| `payload` |  | — |

---

#### `tryRecv(channel)`
```slug
fn slug.channel#tryRecv(@chan channel) -> ?
```


attempts to receive a value from a channel without blocking.

Returns the next value if one is immediately available, `nil` otherwise.

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |

---

#### `trySend(channel, payload)`
```slug
fn slug.channel#trySend(@chan channel, payload) -> ?
```


attempts to send a value to a channel without blocking.

`payload` must not be `nil`.

Returns the channel if the send succeeded, nil if the channel was full.

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |
| `payload` |  | — |
