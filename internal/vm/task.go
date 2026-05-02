package vm

import (
	"slug/internal/object"
	"sync"
)

type VMTaskHandle struct {
	done   chan struct{}
	result object.Object
	mu     sync.Mutex
	closed bool
}

func NewVMTaskHandle() *VMTaskHandle {
	return &VMTaskHandle{
		done: make(chan struct{}),
	}
}

func (h *VMTaskHandle) Type() object.ObjectType { return object.TASK_HANDLE_OBJ }

func (h *VMTaskHandle) Inspect() string { return "<task>" }

func (h *VMTaskHandle) DoneChan() <-chan struct{} {
	return h.done
}

func (h *VMTaskHandle) AwaitResult() object.Object {
	if h.result == nil {
		return object.NIL
	}
	return h.result
}

func (h *VMTaskHandle) Complete(result object.Object) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.result = result
	h.closed = true
	close(h.done)
}

func (h *VMTaskHandle) Cancel(reason string) {
	payload := &object.Map{Pairs: map[object.MapKey]object.MapPair{}}
	payload.Put(&object.String{Value: "type"}, &object.String{Value: "cancelled"})
	payload.Put(&object.String{Value: "reason"}, &object.String{Value: reason})
	h.Complete(&object.RuntimeError{
		Payload: payload,
	})
}
