package config

import (
	"errors"
	"log/slog"

	"api-students/app/model"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewApp merakit aplikasi: membuat instance Fiber, memasang middleware,
// lalu mendaftarkan route. File ini adalah tempat seluruh bagian bertemu.
func NewApp(
	logger *slog.Logger,
	pool *pgxpool.Pool,
	deps route.Dependencies,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Praktikum Backend Lanjut"),
		ErrorHandler: newErrorHandler(logger),

		// Membatasi ukuran body mencegah satu request besar menghabiskan
		// memori server (denial of service paling murah dilakukan).
		BodyLimit: 1 * 1024 * 1024, // 1 MB
	})

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", ""))
	route.Register(app, deps)

	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler adalah jaring pengaman terakhir: error yang tidak
// tertangani di service berakhir di sini dengan format yang tetap konsisten.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID, _ := c.Locals("requestid").(string)
		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):
			// sudah AppError

		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{Status: fiber.StatusRequestEntityTooLarge, Code: "PAYLOAD_TOO_LARGE", Message: "ukuran body melebihi batas"}

		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{Status: fiberErr.Code, Code: "HTTP_ERROR", Message: fiberErr.Message}
			} else {
				appErr = helper.Internal(err) // Perbaikan Bug 3: Interna1 -> Internal
			}
		}

		// Perbaikan Bug 4 & 5: Logika terbalik & nil pointer
		if appErr.Status >= fiber.StatusInternalServerError {
			causeMsg := "unknown"
			if appErr.Cause() != nil {
				causeMsg = appErr.Cause().Error()
			}
			logger.Error("request_failed",
				slog.String("request_id", requestID), slog.String("path", c.Path()),
				slog.Int("status", appErr.Status), slog.String("error", causeMsg))
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID), slog.String("path", c.Path()),
				slog.String("code", appErr.Code), slog.Int("status", appErr.Status))
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success: false, Code: appErr.Code, Message: appErr.Message,
			Fields: appErr.Fields, RequestID: requestID,
		})
	}
}
