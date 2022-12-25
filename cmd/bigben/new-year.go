package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/MrMelon54/bigben/cmd/bigben/message"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

func (b *BigBen) cronNewYears() {
	b.messageNotification("New Year's", message.SendNewYearNotification)()
	archive, err := prepareApiUpload()
	if err != nil {
		log.Printf("[cronNewYears()] Failed to generate archive: %s\n", err)
		return
	}
	resp, err := makeApiUploadRequest(archive)
	if err != nil {
		log.Printf("[cronNewYears()] Failed to upload archive: %s\n", err)
		return
	}
	log.Printf(">>> Archive URL: https://cdn.mrmelon54.com/download/auto/%s <<<\n", resp["Path"])
}

func prepareApiUpload() (*bytes.Buffer, error) {
	archive := new(bytes.Buffer)
	gz := gzip.NewWriter(archive)
	tarGz := tar.NewWriter(gz)

	now := time.Now()

	csvBongLog := new(bytes.Buffer)

	err := tarGz.WriteHeader(genTarFileHeader("bong-log.csv", int64(csvBongLog.Len()), now))
	if err != nil {
		return nil, fmt.Errorf("tarGz.WriteHeader(): %w", err)
	}
	_, err = io.Copy(tarGz, csvBongLog)
	if err != nil {
		return nil, fmt.Errorf("io.Copy(): %w", err)
	}

	err = tarGz.Close()
	if err != nil {
		return nil, fmt.Errorf("tarGz.Close(): %w", err)
	}
	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("gz.Close(): %w", err)
	}
	return archive, nil
}

func genTarFileHeader(name string, size int64, now time.Time) *tar.Header {
	return &tar.Header{
		Typeflag:   tar.TypeReg,
		Name:       name,
		Size:       size,
		Mode:       0,
		Uid:        1000,
		Gid:        1000,
		Uname:      "bigben",
		Gname:      "bigben",
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}
}

func makeApiUploadRequest(archive *bytes.Buffer) (map[string]any, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "bigben.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("writer.CreateFormFile(): %w", err)
	}

	// copy data
	_, err = io.Copy(part, archive)
	if err != nil {
		return nil, fmt.Errorf("io.Copy(): %w", err)
	}

	// close multipart writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("writer.Close(): %w", err)
	}

	// generate request
	req, err := http.NewRequest(http.MethodPost, "https://api.mrmelon54.com/v1/upload", body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest(): %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do(): %w", err)
	}

	if resp.StatusCode == http.StatusCreated {
		a := map[string]any{}
		err = json.NewDecoder(resp.Body).Decode(&a)
		return a, err
	}
	respBody := new(bytes.Buffer)
	_, _ = respBody.ReadFrom(resp.Body)
	return nil, fmt.Errorf("invalid status code %d: %s", resp.StatusCode, strings.TrimSpace(respBody.String()))
}
