package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var jobs = make(map[string]int)

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

	w.Write([]byte(fmt.Sprintf("JobStatus: %d", jobs[jobId])))
	return
}

func updateJob(jobId string, progress int) {
	jobs[jobId] = progress
	fmt.Printf("updated %s to %d\n", jobId, progress)
	if progress == 100 {
		return
	}
	time.AfterFunc(5*time.Second, func() {
		updateJob(jobId, progress+10)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/submit", handleSubmitJob)
	mux.HandleFunc("/checkstatus", handleCheckStatus)

	// submit
	// check status

	port := ":3000"
	log.Printf("Server is listening on port %s", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		panic(err)
	}
}
