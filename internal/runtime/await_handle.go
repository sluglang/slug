package runtime

import "slug/internal/object"

type awaitHandle interface {
	DoneChan() <-chan struct{}
	AwaitResult() object.Object
}
