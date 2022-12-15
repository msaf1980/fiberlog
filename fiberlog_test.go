package fiberlog

import (
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func testGetReq(t *testing.T, address, uri string, statusCode int) {
	req, err := http.NewRequest("GET", address+uri, nil)
	if err != nil {
		t.Errorf("http.NewRequest() error = %v", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Do(/ok) error = %v", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != statusCode {
		t.Errorf("Do(/ok) = %d (%s)", resp.StatusCode, string(body))
	}
}

func TestNew(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	app := fiber.New()

	app.Use(New(Config{
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		app.Listen(":3000")
	}()
	wg.Wait()
	time.Sleep(10 * time.Millisecond)

	testGetReq(t, "http://127.0.0.1:3000", "/ok", http.StatusOK)
	testGetReq(t, "http://127.0.0.1:3000", "/warn", 422)
	testGetReq(t, "http://127.0.0.1:3000", "/err", http.StatusInternalServerError)
	testGetReq(t, "http://127.0.0.1:3000", "/", http.StatusNotFound)
}
