package log

import "github.com/rs/zerolog"

func LoggerString(key, value string) LoggerField {
	return func(c zerolog.Context) zerolog.Context {
		return c.Str(key, value)
	}
}
