package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			b := make([]byte, 16)
			_, err := rand.Read(b)
			if err != nil {
				log.Fatal(err)
			}
			// Format as xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
			uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

			body := `{"payload":"test job ` + uuid + `"}`
			r, err := http.Post(
				"http://localhost:8080/jobs",
				"application/json",
				strings.NewReader(body),
			)

			if err != nil {
				log.Println("failed to create job", err, "error", err)
				return
			}
			log.Println("job created with status", r.StatusCode, "for job", i)
		}(i)
	}

	wg.Wait()
}
