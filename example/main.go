package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/msaf1980/fiberlog"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout)
	app := fiber.New()

	// Default
	// app.Use(fiberlog.New())

	if os.Getenv("FIBERLOG_BASICAUTH") != "" {
		app.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				"test":  "password",
				"admin": "123456",
			},
		}))
	}

	// Custom Config
	app.Use(fiberlog.New(fiberlog.Config{
		Logger: &logger,
		Next: func(ctx *fiber.Ctx) bool {
			// skip if we hit /private
			return ctx.Path() == "/private"
		},
		LogHost:     true,
		LogUsername: "username",
		// TagReqHeader:  []string{"host"},
		TagRespHeader: []string{"content-type"},
	}))

	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Get("/warn", func(c *fiber.Ctx) error {
		return fiber.ErrUnprocessableEntity
	})

	app.Get("/err", func(c *fiber.Ctx) error {
		return fiber.ErrInternalServerError
	})

	logger.Fatal().Err(app.Listen(":3000"))
}
