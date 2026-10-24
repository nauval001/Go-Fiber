package model

import "time"

type Project struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Techs      []string  `json:"techs"`
	Link       string    `json:"link"`
	Year       int       `json:"year"`
}

type Certificate struct {
	Name   string    `json:"name"`
	Issuer string    `json:"issuer"`
	Date   time.Time `json:"date"`
}

type Portfolio struct {
	ID          string        `json:"id"`
	StudentID   int           `json:"student_id"`
	Headline    string        `json:"headline"`
	Bio         string        `json:"bio"`
	Projects    []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Request untuk PUT (ganti seluruh isi)
type UpsertPortfolioRequest struct {
	Headline string `json:"headline" validate:"required,min=5,max=100"`
	Bio      string `json:"bio" validate:"required,min=10,max=1000"`
}

// Request untuk menambah proyek
type AddProjectRequest struct {
	Title string   `json:"title" validate:"required,min=3,max=100"`
	Techs []string `json:"techs" validate:"required,min=1,dive,min=2,max=30"`
	Link  string   `json:"link" validate:"omitempty,url"`
	Year  int      `json:"year" validate:"required,min=2000,max=2100"`
}