package main

import "fmt"

func main() {
	// 1. Deklarasi lima variabel dengan tipe berbeda
	var nama string = "Gibran"
	var umur int = 21
	var ipk float64 = 3.85
	var isAktif bool = true
	hobi := []string{"Membaca", "Coding"}

	fmt.Println("Data Variabel:", nama, umur, ipk, isAktif, hobi)

	// 2. Deklarasi map untuk menyimpan data mahasiswa
	nilaiMahasiswa := make(map[string]int)

	// 3. Operasi menambah data ke map
	nilaiMahasiswa["Gibran"] = 85
	nilaiMahasiswa["Joko"] = 90
	nilaiMahasiswa["Anto"] = 75

	// Operasi membaca dengan pengecekan keberadaan
	if nilai, exists := nilaiMahasiswa["Gibran"]; exists {
		fmt.Println("Nilai Gibran ditemukan:", nilai)
	} else {
		fmt.Println("Data Gibran tidak ada.")
	}

	// Operasi menghapus data dari map
	delete(nilaiMahasiswa, "Anto")

	// Operasi menelusuri seluruh isi map
	fmt.Println("Daftar Nilai Mahasiswa:")
	for key, value := range nilaiMahasiswa {
		fmt.Printf("- %s: %d\n", key, value)
	}
}