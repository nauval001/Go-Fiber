package main

import (
	"fmt"
)

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

// Value Receiver: Hanya membaca data untuk diformat
func (s Student) GetInfo() string {
	return fmt.Sprintf("ID: %d | Name: %s | Grade: %.2f | Active: %v", s.ID, s.Name, s.Grade, s.IsActive)
}

// Pointer Receiver: Memodifikasi data asli
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	// Inisialisasi struct
	std := Student{
		ID:       101,
		Name:     "Muhammad Yazhid",
		Grade:    75.5,
		IsActive: false,
	}

	fmt.Println("Informasi Awal:")
	fmt.Println(std.GetInfo())

	// Memanggil pointer receiver methods
	std.UpdateGrade(92.0)
	std.Activate()

	fmt.Println("\nInformasi Setelah Pembaruan Data:")
	fmt.Println(std.GetInfo())
}