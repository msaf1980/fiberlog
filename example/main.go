package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/msaf1980/fiberlog"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout)
	app := fiber.New()

	// Default
	// app.Use(fiberlog.New())

	// Custom Config
	app.Use(fiberlog.New(fiberlog.Config{
		Logger: &logger,
		Next: func(ctx *fiber.Ctx) bool {
			// skip if we hit /private
			return ctx.Path() == "/private"
		},
		TagReqHeader:  []string{"host"},
		TagRespHeader: []string{"content-type"},
	}))

	app.Get("/ok", func(c *fiber.Ctx) error {
		c.SendString("ok")
		return nil
	})

	app.Get("/warn", func(c *fiber.Ctx) error {
		return fiber.ErrUnprocessableEntity
	})

	app.Get("/err", func(c *fiber.Ctx) error {
		return fiber.ErrInternalServerError
	})

	app.Listen(":3000")
}
