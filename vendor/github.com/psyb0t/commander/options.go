package commander

import (
	"io"
)

type Option func(*Options)

type Options struct {
	Stdin io.Reader
	Env   []string
	Dir   string
}

func WithStdin(stdin io.Reader) Option {
	return func(o *Options) {
		o.Stdin = stdin
	}
}

func WithEnv(env []string) Option {
	return func(o *Options) {
		o.Env = env
	}
}

func WithDir(dir string) Option {
	return func(o *Options) {
		o.Dir = dir
	}
}
