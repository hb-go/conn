package conn

type Options struct {
	ReaderPoolSize int64

	WriterPoolSize int64
}

func newOptions(opt ...Option) Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}

	return opts
}

func ReaderPoolSize(size int64) Option {
	return func(o *Options) {
		o.ReaderPoolSize = size
	}
}

func WriterPoolSize(size int64) Option {
	return func(o *Options) {
		o.WriterPoolSize = size
	}
}
