package fiberlog

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
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

	Username      bool // log from context user parameter username
	UserAgent     bool // log user agent
	ForwardedFor  bool // log X-Forwarded-For (if behing a balancer) or repote IP
	TagReqHeader  []string
	TagRespHeader []string
	Tags          []string // log from context user parameter with keys from tags slice
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
		// X-Request-ID from header
		rid := c.Get(fiber.HeaderXRequestID)
		if rid == "" {
			rid = strconv.FormatUint(c.Context().ID(), 10)
			c.Set(fiber.HeaderXRequestID, rid)
		}

		remoteIP := c.IP()
		dumploggerCtx := sublog.With().
			// Str("pid", pid).
			Str("tag", "request").
			Str("id", rid).
			Int("status", code).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("remote_ip", remoteIP).
			Str("protocol", c.Protocol()).
			Str("host", c.Hostname()).
			Dur("latency", stop.Sub(start))

		if cfg.ForwardedFor {
			forwarded := c.Get(fiber.HeaderXForwardedFor)
			if forwarded == "" {
				forwarded = remoteIP
			}
			dumploggerCtx = dumploggerCtx.Str("forwarded_for", forwarded)
		}
		if cfg.Username {
			i := c.Context().UserValue("username")
			if username, ok := i.(string); ok {
				dumploggerCtx = dumploggerCtx.Str("username", username)
			}
		}
		if cfg.UserAgent {
			dumploggerCtx = dumploggerCtx.Str("user-agent", c.Get(fiber.HeaderUserAgent))
		}
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
		for _, k := range cfg.Tags {
			i := c.Context().UserValue(k)
			if i != nil {
				switch v := i.(type) {
				case string:
					dumploggerCtx = dumploggerCtx.Str(k, v)
				case []string:
					dumploggerCtx = dumploggerCtx.Strs(k, v)
				case int64:
					dumploggerCtx = dumploggerCtx.Int64(k, v)
				case uint64:
					dumploggerCtx = dumploggerCtx.Uint64(k, v)
				case float64:
					dumploggerCtx = dumploggerCtx.Float64(k, v)
				case time.Duration:
					dumploggerCtx = dumploggerCtx.Dur(k, v)
				case bool:
					dumploggerCtx = dumploggerCtx.Bool(k, v)
				case []error:
					dumploggerCtx = dumploggerCtx.Errs(k, v)
				case *zerolog.Event:
					dumploggerCtx = dumploggerCtx.Dict(k, v)
				case zerolog.LogArrayMarshaler:
					dumploggerCtx = dumploggerCtx.Array(k, v)
				case zerolog.LogObjectMarshaler:
					dumploggerCtx = dumploggerCtx.EmbedObject(v)
				case fmt.Stringer:
					dumploggerCtx = dumploggerCtx.Str(k, v.String())
				default:
					dumploggerCtx = dumploggerCtx.Str(k, fmt.Sprintf("<%T>", i))
				}
			}
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
