package di

import (
	"log/slog"
	"os"
)

// GetLogger returns the process-wide structured logger (text handler on stderr).
func (c *Container) GetLogger() *slog.Logger {
	if c.logger == nil {
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
		c.logger = slog.New(handler)
	}
	return c.logger
}
