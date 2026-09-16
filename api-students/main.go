package main

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}

func main() {
	app := fiber.New()
	app.Use(logger.New())

	api := app.Group("/api/v1")
	
	// Terapkan middleware requireJSON khusus untuk endpoint students
	s := api.Group("/students", requireJSON)
	
	s.Get("/", listStudents)
	s.Get("/:id", getStudent)
	s.Post("/", createStudent)
	s.Put("/:id", replaceStudent)
	s.Patch("/:id", patchStudent)
	s.Delete("/:id", deleteStudent)

	// Endpoint tidak dikenal (404 global)
	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	log.Println("Server berjalan di http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}