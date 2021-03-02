package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CreateTarballFrom(src string, dest io.WriteCloser) error {
	defer dest.Close()
	gzipWriter := gzip.NewWriter(dest)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		header := &tar.Header{
			Name:    p,
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not write header for file %s", p))
		}

		file, err := os.Open(p)
		if err != nil {
			return err
		}

		if _, err := io.Copy(tarWriter, file); err != nil {
			return err
		}

		return nil
	})
}
