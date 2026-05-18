package main

import (
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

			body := `{"payload":"test job ` + string(rune(i)) + `"}`
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
