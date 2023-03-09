package main

import (
	"fmt"
	"log"
	"net/http"
)

func server() {
	fmt.Println("Server Starting")
	manager := JobManager{
		JobLogTracker:    make(LogTracker),
		JobStatusTracker: make(map[string]JobStatus),
	}
	http.HandleFunc("/api/v1/job/launch", manager.jobLaunchHandler)
	http.HandleFunc("/api/v1/job/logs", manager.jobLogHandler)

	err := http.ListenAndServe(":9021", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	server()
}
