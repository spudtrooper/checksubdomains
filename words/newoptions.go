// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package words

import "time"

type NewOption func(*newOptionImpl)

type NewOptions interface {
	Timeout() time.Duration
	Threads() int
	Start() string
}

func NewTimeout(timeout time.Duration) NewOption {
	return func(opts *newOptionImpl) {
		opts.timeout = timeout
	}
}
func NewTimeoutFlag(timeout *time.Duration) NewOption {
	return func(opts *newOptionImpl) {
		opts.timeout = *timeout
	}
}

func NewThreads(threads int) NewOption {
	return func(opts *newOptionImpl) {
		opts.threads = threads
	}
}
func NewThreadsFlag(threads *int) NewOption {
	return func(opts *newOptionImpl) {
		opts.threads = *threads
	}
}

func NewStart(start string) NewOption {
	return func(opts *newOptionImpl) {
		opts.start = start
	}
}
func NewStartFlag(start *string) NewOption {
	return func(opts *newOptionImpl) {
		opts.start = *start
	}
}

type newOptionImpl struct {
	timeout time.Duration
	threads int
	start   string
}

func (n *newOptionImpl) Timeout() time.Duration { return n.timeout }
func (n *newOptionImpl) Threads() int           { return n.threads }
func (n *newOptionImpl) Start() string          { return n.start }

func makeNewOptionImpl(opts ...NewOption) *newOptionImpl {
	res := &newOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeNewOptions(opts ...NewOption) NewOptions {
	return makeNewOptionImpl(opts...)
}
