# Module 8: Concurrency

Slug concurrency is structured: tasks belong to a logical scope and are awaited explicitly.

## Lesson 8.1: Core ideas

- `spawn` starts child work.
- `await(handle)` waits for result.
- `nursery` defines ownership boundary.
- errors/cancellation flow through the structure.

## Lesson 8.2: Start with one spawned task

```slug
val work = nursery fn() {
  val t = spawn { 20 + 22 }
  await(t)
}

println(work())
```

Expected output:

```text
42
```

## Lesson 8.3: Add timeout-aware await

```slug
val getSlow = nursery fn() {
  val t = spawn { slowTask() }
  await(t, 500)
}
```

### Mental model
- `await(handle)` blocks until completion.
- `await(handle, timeoutMs)` throws on timeout.

## Lesson 8.4: Fan-out and fan-in

```slug
val fetchBoth = nursery fn(id) {
  val userT = spawn { fetchUser(id) }
  val postsT = spawn { fetchPosts(id) }

  val user = await(userT, 500)
  val posts = await(postsT, 1000)

  { user: user, posts: posts }
}
```

## Lesson 8.5: Limits with `nursery limit`

```slug
val handler = nursery limit 10 fn(ids) {
  ids /> map(fn(id) { spawn { fetchUser(id) } })
}
```

Use limits when fan-out can grow large.

## Lesson 8.6: Error handling around async flows

```slug
val run = nursery fn() {
  defer onerror(err) { println("task flow failed:", err) }
  val t = spawn { riskyWork() }
  await(t)
}
```

Common mistakes:
- Forgetting to await handles you care about.
- Treating cancellation as immediate at every line (it is observed at suspension points).

### Try it

Implement `awaitAll(handles)` that returns results in order and rethrows first failure.
