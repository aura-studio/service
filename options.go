package service

import "time"

type Option func(*Service)

type Options struct {
	Timeout time.Duration
}

var defaultOptions = Options{
	Timeout: 10 * time.Second,
}

func WithTimeout(timeout time.Duration) Option {
	return func(s *Service) {
		s.Timeout = timeout
	}
}
