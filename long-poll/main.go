package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var jobs = make(map[string]int)

// -----HANDLERS-----

func handleSubmitJob(w http.ResponseWriter, _ *http.Request) {
	currentDate := time.Now()
	jobId := fmt.Sprintf("job:%d", currentDate.UnixMilli())
	jobs[jobId] = 0
	updateJob(jobId, 0)

	w.Write([]byte(fmt.Sprintf("%s", jobId)))
	return
}

func handleCheckStatus(w http.ResponseWriter, r *http.Request) {
	jobId := r.URL.Query().Get("jobId")

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			if jobs[jobId] >= 100 {
				w.Write([]byte(fmt.Sprintf("Job Completed: %d", jobs[jobId])))
				return
			}
		}
	}
}

// -----HTTPSERVER-----

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/submit", handleSubmitJob)
	mux.HandleFunc("/checkstatus", handleCheckStatus)

	port := ":3000"
	log.Printf("Server is listening on port %s", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		panic(err)
	}
}

// -----UTILITY FUNCTIONS-----

func updateJob(jobId string, progress int) {
	jobs[jobId] = progress
	fmt.Printf("updated %s to %d\n", jobId, progress)
	if progress == 100 {
		return
	}
	time.AfterFunc(2*time.Second, func() {
		updateJob(jobId, progress+10)
	})
}

func checkJobComplete(jobId string, ch chan<- bool) {
	// check the job progress every 1 second
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			if jobs[jobId] >= 100 {
				ch <- true
			}
		}
	}
}
