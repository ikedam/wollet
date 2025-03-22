package log

import "github.com/rs/zerolog"

func String(key, value string) LogField {
	return func(e *zerolog.Event) *zerolog.Event {
		return e.Str(key, value)
	}
}

func Int(key string, value int) LogField {
	return func(e *zerolog.Event) *zerolog.Event {
		return e.Int(key, value)
	}
}

func WithError(err error) LogField {
	return func(e *zerolog.Event) *zerolog.Event {
		return e.Err(err)
	}
}
