package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"io"
	"log"
	"net/http"
	"os"
)

func (jm *JobManager) launchContainer(image string, sourceDir string, artifactDir string, jobID string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	workdir := "/tmp"
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      image,
		Cmd:        []string{"python", "-u", "train.py"},
		Volumes:    map[string]struct{}{},
		WorkingDir: workdir,
		Tty:        true,
	},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: sourceDir,
					Target: workdir,
				},
				{
					Type:   mount.TypeBind,
					Source: artifactDir,
					Target: "/opt/ml/model",
				},
			},
		}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	jm.JobStatusTracker[jobID] = InProgress

	go func() {
		reader, err := cli.ContainerLogs(context.Background(), resp.ID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Timestamps: false,
		})
		if err != nil {
			panic(err)
		}
		defer reader.Close()
		defer func() {
			jm.JobStatusTracker[jobID] = Completed
			log.Println("Job completed")
		}()

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			//fmt.Println(scanner.Text())
			logTracker := jm.JobLogTracker[jobID]
			_, err := logTracker.Write([]byte(scanner.Text() + "\n"))
			if err != nil {
				log.Println("Error writing to log buffer")
				return
			}
		}
	}()

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
}

type JobLaunchResponse struct {
	JobID string
}

type LogTracker map[string]io.ReadWriter

type JobStatus int

const (
	InProgress JobStatus = iota
	Failed
	Completed
	Creating
)

func (s JobStatus) MarshalJSON() ([]byte, error) {
	var status string
	switch s {
	case InProgress:
		status = "InProgress"
	case Failed:
		status = "Failed"
	case Completed:
		status = "Completed"
	case Creating:
		status = "Creating"
	}
	return json.Marshal(status)
}

type JobManager struct {
	JobLogTracker    LogTracker
	JobStatusTracker map[string]JobStatus
}

func (jm *JobManager) jobLaunchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form in the request
	err := r.ParseMultipartForm(10 << 20) // 10 MB maximum file size
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the file from the multipart form
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	id := jm.jobRunner(file)
	response := JobLaunchResponse{
		JobID: id,
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

type JobLogsResponse struct {
	Logs      string    `json:"Logs"`
	JobStatus JobStatus `json:"JobStatus"`
	NewOffset int64     `json:"NewOffset"`
}

type LogRequest struct {
	JobID  string `json:"jobId"`
	Offset int64  `json:"offset"`
}

func (jm *JobManager) jobLogHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var logRequest LogRequest
	err := json.NewDecoder(r.Body).Decode(&logRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jobId := logRequest.JobID
	offset := logRequest.Offset
	if jobId == "" {
		http.Error(w, "Job ID not provided", http.StatusBadRequest)
		return
	}

	logBuf, ok := jm.JobLogTracker[jobId]
	if !ok {
		http.Error(w, "Job ID not found", http.StatusNotFound)
		return
	}

	// Read from the offset
	logs, err := io.ReadAll(logBuf)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JobLogsResponse{
		Logs:      string(logs),
		JobStatus: jm.JobStatusTracker[jobId],
		NewOffset: offset + int64(len(logs)),
	})
}

// Launch a container with the given image and source directory
// Return a unique ID for the job
func (jm *JobManager) jobRunner(source io.Reader) string {
	// TODO: Clean up the temp directories
	sourceDir, err := os.MkdirTemp("", "wiseml-source")
	if err != nil {
		log.Fatal(err)
	}
	artifactDir, err := os.MkdirTemp("", "wiseml-artifact")
	if err != nil {
		log.Fatal(err)
	}

	// Unzip the file
	unzipToDir(source, sourceDir)

	jobId := uuid.Generate().String()
	// Create a buffer to write the logs to
	buf := new(bytes.Buffer)
	jm.JobLogTracker[jobId] = buf

	go jm.launchContainer(
		"tensorflow/tensorflow",
		sourceDir,
		artifactDir,
		jobId,
	)

	fmt.Println("Job Launched with ID: ", jobId)
	return jobId
}
