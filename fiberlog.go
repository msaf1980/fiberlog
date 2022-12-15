package fiberlog

import (
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config defines the config for logger middleware.
type Config struct {
	// Next defines a function to skip this middleware.
	Next func(ctx *fiber.Ctx) bool

	// Logger is a *zerolog.Logger that writes the logs.
	//
	// Default: log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	Logger *zerolog.Logger

	UserAgent     bool
	TagReqHeader  []string
	TagRespHeader []string
}

// New is a zerolog middleware that allows you to pass a Config.
//
//	app := fiber.New()
//
//	// Without config
//	app.Use(New())
//
//	// With config
//	app.Use(New(Config{Logger: &zerolog.New(os.Stdout)}))
func New(config ...Config) fiber.Handler {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set PID once
	// pid := strconv.Itoa(os.Getpid())

	var sublog zerolog.Logger
	if cfg.Logger == nil {
		sublog = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		sublog = *cfg.Logger
	}

	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		start := time.Now()

		// Handle request, store err for logging
		chainErr := c.Next()

		if chainErr != nil {
			if err := c.App().ErrorHandler(c, chainErr); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		// Set latency stop time
		stop := time.Now()

		code := c.Response().StatusCode()

		dumploggerCtx := sublog.With().
			// Str("pid", pid).
			Uint64("id", c.Context().ID()).
			Int("status", code).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Dur("latency", stop.Sub(start))

		for _, k := range cfg.TagReqHeader {
			if v := c.Get(k); v != "" {
				dumploggerCtx = dumploggerCtx.Str(k, v)
			}
		}
		for _, k := range cfg.TagRespHeader {
			if v := c.GetRespHeader(k); v != "" {
				dumploggerCtx = dumploggerCtx.Str(k, v)
			}
		}

		if cfg.UserAgent {
			dumploggerCtx = dumploggerCtx.Str("user-agent", c.Get(fiber.HeaderUserAgent))
		}

		msg := ""
		if chainErr != nil {
			msg = chainErr.Error()
		}

		dumplogger := dumploggerCtx.Logger()
		switch {
		case code >= fiber.StatusBadRequest && code < fiber.StatusInternalServerError:
			dumplogger.Warn().Msg(msg)
		case code >= http.StatusInternalServerError:
			dumplogger.Error().Msg(msg)
		default:
			dumplogger.Info().Msg(msg)
		}

		return nil
	}
}
