package main

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
    "sync"
    "time"
)

func main() {
    // PASTE TOKEN ADMIN ANDA DI SINI
    adminToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluMSIsInJvbGUiOiJhZG1pbiIsImlzcyI6InByYWt0aWt1bS1iYWNrZW5kIiwic3ViIjoiMiIsImV4cCI6MTc5MjcyMTk5NiwiaWF0IjoxNzkyNzIxMDk2fQ.WT8hlszESn-vpJ3y-Ffcz2xq9bCW7qRqJpGBODt3eFE" 
    url := "http://localhost:3000/api/v1/supervisors/1/assign"

    totalRequests := 30
    fmt.Printf("Menembak %d request secara serentak ke Supervisor 1 (Kuota 3)...\n", totalRequests)

    var wg sync.WaitGroup
    start := make(chan struct{})
    results := make(chan int, totalRequests)

    for i := 1; i <= totalRequests; i++ {
        wg.Add(1)
        go func(studentID int) {
            defer wg.Done()
            <-start // Tunggu gerbang dibuka

            body := fmt.Sprintf(`{"student_id": %d}`, studentID)
            req, _ := http.NewRequest("POST", url, bytes.NewBufferString(body))
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("Authorization", "Bearer "+adminToken)

            client := &http.Client{Timeout: 5 * time.Second}
            resp, err := client.Do(req)
            if err != nil {
                results <- 0
                return
            }
            defer resp.Body.Close()
            io.Copy(io.Discard, resp.Body)
            results <- resp.StatusCode
        }(i)
    }

    close(start) // Tembak bersamaan!
    wg.Wait()
    close(results)

    tally := map[int]int{}
    for code := range results {
        tally[code]++
    }

    fmt.Println("\n=== HASIL UJI BEBAN ===")
    fmt.Printf("Sukses (200 OK)      : %d\n", tally[200])
    fmt.Printf("Gagal (409 Conflict) : %d\n", tally[409])
    fmt.Printf("Error Lain/Timeout   : %d\n", tally[0])
	for code, count := range tally {
		fmt.Printf("Status %d : %d\n", code, count)
	}
}