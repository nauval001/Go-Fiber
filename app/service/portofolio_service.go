package service

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	
	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository // Untuk cek apakah student ID ada di PostgreSQL
}

func NewPortfolioService(portfolios repository.PortfolioRepository, students repository.StudentRepository) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students}
}

// GET /students/:id/portfolio
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid { return helper.BadRequest("id tidak valid") }

	portfolio, err := s.portfolios.FindByStudentID(ctx, studentID)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portofolio belum dibuat")
		}
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio ditemukan", portfolio)
}

// PUT /students/:id/portfolio
func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid { return helper.BadRequest("id tidak valid") }

	// Validasi apakah student ID benar-benar ada di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
	}

	var req model.UpsertPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	saved, err := s.portfolios.Upsert(ctx, model.Portfolio{
		StudentID: studentID,
		Headline:  req.Headline,
		Bio:       req.Bio,
	})
	if err != nil { return helper.Internal(err) }
	
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", saved)
}

// POST /students/:id/portfolio/projects
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid { return helper.BadRequest("id tidak valid") }

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	err := s.portfolios.AddProject(ctx, studentID, model.Project{
		Title: req.Title, Techs: req.Techs, Link: req.Link, Year: req.Year,
	})
	if err != nil {
		if errors.Is(err, repository.ErrProjectLimit) {
			return helper.Conflict("maksimal 20 proyek tercapai")
		}
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("buat portofolio terlebih dahulu")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", nil)
}

// DELETE /students/:id/portfolio/projects/:projectId
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid { return helper.BadRequest("id tidak valid") }
	
	projectID := c.Params("projectId")
	if projectID == "" { return helper.BadRequest("id proyek harus diisi") }

	if err := s.portfolios.DeleteProject(ctx, studentID, projectID); err != nil {
		return helper.Internal(err)
	}
	return helper.NoContent(c)
}

// GET /portfolios/stats/technologies
func (s *PortfolioService) StatsTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.GetTopTechnologies(ctx)
	if err != nil { return helper.Internal(err) }
	
	return helper.Success(c, fiber.StatusOK, "statistik teknologi teratas", stats)
}