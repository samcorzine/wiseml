package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
)
import "net/http"

func SimpleIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %s!", r.URL.Path[1:])
}

func HttpFileHandler(response http.ResponseWriter, request *http.Request) {
	http.ServeFile(response, request, "Index.html")
}

func server() {
	fmt.Println("Server Starting")
	http.HandleFunc("/", SimpleIndexHandler)
	http.HandleFunc("/index", HttpFileHandler)

	http.ListenAndServe(":8080", nil)
}

func launchContainer(image string, sourceDir string, artifactDir string) {
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
		//Env:        []string{"TF_CPP_MIN_LOG_LEVEL=0"},
		Tty: true,
		//AttachStdin:  true,
		//AttachStdout: true,
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

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
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

func ExtractTarGz(gzipStream io.Reader) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}
		//println(header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(header.Name, 0755)
			if err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %s in %s",
				header.Typeflag,
				header.Name)
		}

	}
}

func main() {
	r, err := os.Open("/Users/samcorzine/playground/wiseml/src.tgz")
	if err != nil {
		fmt.Println("error")
	}
	ExtractTarGz(r)
	launchContainer(
		"tensorflow/tensorflow",
		"/Users/samcorzine/GolandProjects/wiseml-server/src",
		"/Users/samcorzine/GolandProjects/wiseml-server/model")
}
