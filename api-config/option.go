package config

type Options struct {
	FileName  string
	FilePaths []string
	AppEnv    string
}

func DefaultConfigOptions() Options {
	opts := Options{
		FileName: "config",
		FilePaths: []string{
			".",
			"./configs",
		},
	}

	return opts
}

type ConfigOption func(o *Options)

func WithFileName(n string) ConfigOption {
	return func(o *Options) {
		o.FileName = n
	}
}

func WithFilePaths(p ...string) ConfigOption {
	return func(o *Options) {
		o.FilePaths = p
	}
}

func WithAppEnv(e string) ConfigOption {
	return func(o *Options) {
		o.AppEnv = e
	}
}
