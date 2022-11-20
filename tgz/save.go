package tgz

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Save(buf *bytes.Buffer, module string) error {

	gz, err := gzip.NewReader(buf)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

tr:
	for {
		head, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break tr
		} else if err != nil {
			return err
		}
		if head.FileInfo().IsDir() {
			continue tr
		}

		parts := strings.Split(head.Name, "/")
		if len(parts) <= 1 {
			continue tr
		}
		parts[0] = module
		crt, err := create(strings.Join(parts, string(os.PathSeparator)))
		if err != nil {
			return err
		}

		err = crt.Chmod(os.FileMode(head.Mode))
		if err != nil {
			return err
		}

		_, err = io.Copy(crt, tr)
		if err != nil {
			return err
		}

		crt.Close()
	}
	return nil
}

func create(file string) (*os.File, error) {
	base := filepath.Dir(file)
	err := os.MkdirAll(base, 0755)
	if err != nil {
		return nil, err
	}
	crt, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	return crt, nil
}
