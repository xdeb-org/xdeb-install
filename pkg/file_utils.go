package xdeb

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

func compressData(in io.Reader, out io.Writer) error {
	// Copied from the library's usage example:
	// https://github.com/klauspost/compress/tree/master/zstd#usage
	enc, err := zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedBestCompression))

	if err != nil {
		return err
	}

	_, err = io.Copy(enc, in)

	if err != nil {
		enc.Close()
		return err
	}

	return enc.Close()
}

func decompressData(in io.Reader, out io.Writer) error {
	// Copied from the library's usage example:
	// https://github.com/klauspost/compress/tree/master/zstd#decompressor
	d, err := zstd.NewReader(in)

	if err != nil {
		return err
	}

	defer d.Close()

	// Copy content...
	_, err = io.Copy(out, d)
	return err
}

func decompressFile(path string) ([]byte, error) {
	file, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(file)

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	if err = decompressData(reader, writer); err != nil {
		return nil, err
	}

	if err = writer.Flush(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func writeFile(path string, data []byte) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return "", err
	}

	file, err := os.Create(path)

	if err != nil {
		return "", err
	}

	defer file.Close()
	_, err = file.Write(data)

	return path, err
}

func writeFileCompressed(path string, data []byte) (string, error) {
	reader := bytes.NewReader(data)

	var compressedData bytes.Buffer
	writer := bufio.NewWriter(&compressedData)

	if err := compressData(reader, writer); err != nil {
		return "", err
	}

	if err := writer.Flush(); err != nil {
		return "", err
	}

	return writeFile(fmt.Sprintf("%s.zst", path), compressedData.Bytes())
}

func DownloadFile(path string, baseRequestUrl string, followRedirects bool, compress bool) (string, error) {
	requestUrl := baseRequestUrl

	client := NewHttpClient()
	resp, err := client.Get(requestUrl)

	if err != nil {
		return "", fmt.Errorf("could not download file '%s'", requestUrl)
	}

	if followRedirects {
		requestUrl = resp.Request.URL.String()
		resp, err = client.Get(requestUrl)

		if err != nil {
			return "", fmt.Errorf("could not download file '%s'", requestUrl)
		}
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("could not download file '%s'", requestUrl)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if compress {
		return writeFileCompressed(path, body)
	}

	return writeFile(path, body)
}
