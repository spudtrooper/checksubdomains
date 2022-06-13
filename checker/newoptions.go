// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package checker

import "time"

type NewOption func(*newOptionImpl)

type NewOptions interface {
	Sublist3r() string
	Timeout() time.Duration
	Threads() int
	SubdomainsFile() string
	HtmlOutputFile() string
	Verbose() bool
}

func NewSublist3r(sublist3r string) NewOption {
	return func(opts *newOptionImpl) {
		opts.sublist3r = sublist3r
	}
}
func NewSublist3rFlag(sublist3r *string) NewOption {
	return func(opts *newOptionImpl) {
		opts.sublist3r = *sublist3r
	}
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

func NewSubdomainsFile(subdomainsFile string) NewOption {
	return func(opts *newOptionImpl) {
		opts.subdomainsFile = subdomainsFile
	}
}
func NewSubdomainsFileFlag(subdomainsFile *string) NewOption {
	return func(opts *newOptionImpl) {
		opts.subdomainsFile = *subdomainsFile
	}
}

func NewHtmlOutputFile(htmlOutputFile string) NewOption {
	return func(opts *newOptionImpl) {
		opts.htmlOutputFile = htmlOutputFile
	}
}
func NewHtmlOutputFileFlag(htmlOutputFile *string) NewOption {
	return func(opts *newOptionImpl) {
		opts.htmlOutputFile = *htmlOutputFile
	}
}

func NewVerbose(verbose bool) NewOption {
	return func(opts *newOptionImpl) {
		opts.verbose = verbose
	}
}
func NewVerboseFlag(verbose *bool) NewOption {
	return func(opts *newOptionImpl) {
		opts.verbose = *verbose
	}
}

type newOptionImpl struct {
	sublist3r      string
	timeout        time.Duration
	threads        int
	subdomainsFile string
	htmlOutputFile string
	verbose        bool
}

func (n *newOptionImpl) Sublist3r() string      { return n.sublist3r }
func (n *newOptionImpl) Timeout() time.Duration { return n.timeout }
func (n *newOptionImpl) Threads() int           { return n.threads }
func (n *newOptionImpl) SubdomainsFile() string { return n.subdomainsFile }
func (n *newOptionImpl) HtmlOutputFile() string { return n.htmlOutputFile }
func (n *newOptionImpl) Verbose() bool          { return n.verbose }

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
