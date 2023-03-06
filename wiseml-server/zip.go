package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
)

func unzipToDir(file io.Reader, outDir string) {
	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Loop through the files in the tar file
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of tar archive
			break
		}
		if err != nil {
			panic(err)
		}

		// Extract the file
		path := header.Name
		targetDir := outDir
		targetPath := targetDir + "/" + path

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it doesn't exist
			if _, err := os.Stat(targetPath); err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(targetPath, 0755)
					if err != nil {
						panic(err)
					}
				}
			}
		case tar.TypeReg:
			// Create the file
			file, err := os.Create(targetPath)
			if err != nil {
				panic(err)
			}

			// Copy the contents of the file from the tar archive
			_, err = io.Copy(file, tarReader)
			if err != nil {
				panic(err)
			}

			// Close the file
			err = file.Close()
			if err != nil {
				panic(err)
			}
		}
	}
}
