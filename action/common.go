package action

import "golang.org/x/net/context"

type Action interface {
	Key() string
	Columns() string
	Type() string
	Unique() string
	Valid() bool
}

type Actions []Action

type Options struct {
	Secure     bool
	Owner      string
	Group      string
	Restrict   bool
	TargetPath string
}

func (slice Actions) Len() int {
	return len(slice)
}

func (slice Actions) Less(i, j int) bool {
	return slice[i].Key() < slice[j].Key()
}

func (slice Actions) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func NewOptions() *Options {
	return &Options{
		Secure:     false,
		Restrict:   false,
		TargetPath: "",
	}
}

func GetContext(options *Options, manifest *Manifest) context.Context {
	ctx := context.WithValue(context.Background(), "options", options)
	ctx = context.WithValue(ctx, "manifest", manifest)
	return ctx
}
