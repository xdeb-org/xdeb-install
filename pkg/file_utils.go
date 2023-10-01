package xdeb

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func writeFile(path string, bytes []byte) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)

	if err != nil {
		return err
	}

	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()
	_, err = file.Write(bytes)

	return err
}

func DownloadFile(path string, url string, followRedirects bool) (string, error) {
	client := &http.Client{}
	resp, err := client.Get(url)

	if err != nil {
		return "", fmt.Errorf("Could not download file %s", url)
	}

	if followRedirects {
		url = resp.Request.URL.String()
		resp, err = client.Get(url)

		if err != nil {
			return "", fmt.Errorf("Could not download file %s", url)
		}
	}

	defer resp.Body.Close()

	fullPath := filepath.Join(path, filepath.Base(url))
	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return fullPath, writeFile(fullPath, bytes)
}
