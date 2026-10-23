package helper

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"api-students/app/model"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))
	if accept == "" { return offered[0], nil }

	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", NotAcceptable("format yang diminta tidak tersedia, pilih salah satu dari: " + strings.Join(offered, ", "))
	}
	return chosen, nil
}

func WriteStudentsCSV(c *fiber.Ctx, students []model.Student) error {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)
	
	header := []string{"id", "nim", "name", "grade", "is_active", "owner_id", "created_at"}
	_ = writer.Write(header)

	for _, s := range students {
		row := []string{
			strconv.Itoa(s.ID), s.NIM, s.Name,
			strconv.FormatFloat(s.Grade, 'f', 2, 64),
			strconv.FormatBool(s.IsActive),
			strconv.Itoa(s.OwnerID),
			s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		_ = writer.Write(row)
	}
	writer.Flush()

	c.Set(fiber.HeaderContentType, FormatCSV)
	c.Set("Content-Disposition", `attachment; filename="students.csv"`)
	return c.SendString(buffer.String())
}