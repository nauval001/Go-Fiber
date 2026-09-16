package main

type Student struct {
	ID       int     `json:"id"`
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// POST: ID digenerate otomatis, field lainnya wajib (misal IsActive kita set default true nanti)
type CreateStudentRequest struct {
	NIM   string  `json:"nim"`
	Name  string  `json:"name"`
	Grade float64 `json:"grade"`
}

// PUT: Seluruh field wajib dikirim karena ini mengganti seluruh data
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PATCH: Field menggunakan pointer agar bisa dibedakan antara "tidak dikirim" (nil) dan "dikirim kosong"
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// Amplop baku untuk semua respons
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}