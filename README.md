# fiberlog

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/msaf1980/fiberlog)

HTTP request/response logger for [Fiber](https://github.com/gofiber/fiber) using [zerolog](https://github.com/rs/zerolog).

### Install

```sh
go get -u github.com/gofiber/fiber/v2
go get -u github.com/msaf1980/fiberlog
```

### Usage

```go
package main

import (
  "github.com/gofiber/fiber/v2"
  "github.com/msaf1980/fiberlog"
)

func main() {
  app := fiber.New()

  // Default
  app.Use(fiberlog.New())

  // Custom Config
  app.Use(fiberlog.New(fiberlog.Config{
    Logger: &zerolog.New(os.Stdout),
    Next: func(ctx *fiber.Ctx) bool {
      // skip if we hit /private
      return ctx.Path() == "/private"
    },
  }))

  app.Listen(3000)
}
```

If we need log addional fields, it's possible

```go
package main

import (
  "github.com/gofiber/fiber/v2"
  "github.com/msaf1980/fiberlog"
)

func main() {
  app := fiber.New()

  // Default
  app.Use(fiberlog.New())

  // basic auth
  app.Use(basicauth.New(basicauth.Config{
    Users: map[string]string{
      "test":  "password",
      "admin": "123456",
    },
  }))

  // Custom Config
  app.Use(fiberlog.New(fiberlog.Config{
    Logger: &zerolog.New(os.Stdout),
    Next: func(ctx *fiber.Ctx) bool {
      // skip if we hit /private
      return ctx.Path() == "/private"
    },
    // TagReqHeader:  []string{},
    TagRespHeader:   []string{"content-type"},
    LogUsername:     "username",
    LogUserAgent:    true,
    LogForwardedFor: true,
    // store it in handler like c.Context().SetUserValue("test", "test_value")
    Tags:            []string{"test"},
  }))

  app.Listen(3000)
}
```

### Example

Run app server:

```sh
$ go run example/main.go
```

Test request:

```sh
$ curl http://localhost:3000/ok
$ curl -u test:password http://localhost:3000/basic
$ curl http://localhost:3000/warn
$ curl http://localhost:3000/err
```

![screen](./example/screen.png)
