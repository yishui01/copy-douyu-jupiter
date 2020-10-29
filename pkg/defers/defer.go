package defers

import "copy/pkg/util/xdefer"

var (
	globalDefers = xdefer.NewStack()
)

func Register(fns ...func() error) {
	globalDefers.Push(fns...)
}

func Clean() {
	globalDefers.Clean()
}
