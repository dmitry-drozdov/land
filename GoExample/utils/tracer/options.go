package tracer

import "crypto/tls"

type tracerOptions struct {
	endpoint    string
	serviceName string
	insecure    bool
	tlsConfig   *tls.Config
}

type Option func(options *tracerOptions) error

func WithEndpoint(e string) Option {
	return func(opts *tracerOptions) error {
		opts.endpoint = e
		return nil
	}
}

func WithServiceName(s string) Option {
	return func(opts *tracerOptions) error {
		opts.serviceName = s
		return nil
	}
}

func WithInsecure(c bool) Option {
	return func(opts *tracerOptions) error {
		opts.insecure = c
		return nil
	}
}

func WithTLS(cfg *tls.Config) Option {
	return func(opts *tracerOptions) error {
		opts.tlsConfig = cfg
		return nil
	}
}
