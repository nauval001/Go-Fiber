package main

import "fmt"

// Menukar nilai dua integer
func swap(a, b *int) {
	*a, *b = *b, *a
}

// Menambahkan item baru ke slice melalui pointer
func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

// Menerima salinan nilai
func passByValue(x int) {
	x = 100
}

// Menerima alamat memori
func passByPointer(x *int) {
	*x = 100
}

func main() {
	// Uji coba swap
	x, y := 10, 20
	fmt.Printf("Sebelum swap: x=%d, y=%d\n", x, y)
	swap(&x, &y)
	fmt.Printf("Setelah swap: x=%d, y=%d\n", x, y)

	// Uji coba updateSlice
	listBuah := []string{"Pisang", "Jeruk"}
	updateSlice(&listBuah, "Mangga")
	fmt.Println("Slice setelah diupdate:", listBuah)

	// Perbandingan Value vs Pointer
	num := 42
	passByValue(num)
	fmt.Println("Hasil Pass by Value:", num) // Tetap 42

	passByPointer(&num)
	fmt.Println("Hasil Pass by Pointer:", num) // Menjadi 100
}