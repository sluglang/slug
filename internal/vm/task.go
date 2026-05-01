package vm

import "slug/internal/object"

type VMTaskHandle struct {
	done   chan struct{}
	result object.Object
}

func NewVMTaskHandle() *VMTaskHandle {
	return &VMTaskHandle{
		done: make(chan struct{}),
	}
}

func (h *VMTaskHandle) Type() object.ObjectType { return object.TASK_HANDLE_OBJ }

func (h *VMTaskHandle) Inspect() string { return "<task>" }

func (h *VMTaskHandle) Complete(result object.Object) {
	h.result = result
	close(h.done)
}
