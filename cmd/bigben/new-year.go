package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/MrMelon54/bigben/cmd/bigben/message"
	"github.com/MrMelon54/bigben/tables"
	"github.com/mohae/struct2csv"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
	"xorm.io/xorm"
)

func (b *BigBen) cronNewYears() {
	b.messageNotification("New Year's", message.SendNewYearNotification)()
	if b.uploadToken != "" {
		archive, err := prepareApiUpload(b.engine)
		if err != nil {
			log.Printf("[cronNewYears()] Failed to generate archive: %s\n", err)
			return
		}
		resp, err := makeApiUploadRequest(archive, b.uploadToken)
		if err != nil {
			log.Printf("[cronNewYears()] Failed to upload archive: %s\n", err)
			return
		}
		log.Printf(">>> Archive URL: https://cdn.mrmelon54.com/download/auto/%s <<<\n", resp["Path"])
	}
}

func prepareApiUpload(engine *xorm.Engine) (*bytes.Buffer, error) {
	archive := new(bytes.Buffer)
	gz := gzip.NewWriter(archive)
	tarGz := tar.NewWriter(gz)

	err := writeCsvFile[tables.BongLog](tarGz, "bong-log.csv", &tables.BongLog{}, engine)
	if err != nil {
		return nil, fmt.Errorf("writeCsvFile(BongLog): %w", err)
	}
	err = writeCsvFile[tables.GuildSettings](tarGz, "guild-settings.csv", &tables.GuildSettings{}, engine)
	if err != nil {
		return nil, fmt.Errorf("writeCsvFile(GuildSettings): %w", err)
	}
	err = writeCsvFile[tables.RoleLog](tarGz, "role-log.csv", &tables.RoleLog{}, engine)
	if err != nil {
		return nil, fmt.Errorf("writeCsvFile(RoleLog): %w", err)
	}
	err = writeCsvFile[tables.UserLog](tarGz, "user-log.csv", &tables.UserLog{}, engine)
	if err != nil {
		return nil, fmt.Errorf("writeCsvFile(UserLog): %w", err)
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

func writeCsvFile[T any](tarGz *tar.Writer, name string, t *T, engine *xorm.Engine) error {
	now := time.Now()

	csvBongLog := new(bytes.Buffer)
	csvWriter := struct2csv.NewWriter(csvBongLog)
	err := csvWriter.WriteColNames(*t)
	if err != nil {
		return fmt.Errorf("csvWriter.WriteColNames(): %w", err)
	}
	err = engine.Iterate(t, func(idx int, bean interface{}) error {
		if t2, ok := bean.(*T); ok {
			return csvWriter.WriteStruct(*t2)
		}
		return fmt.Errorf("failed to convert to iterating type")
	})
	if err != nil {
		return fmt.Errorf("engine.Iterate(): %w", err)
	}
	csvWriter.Flush()

	err = tarGz.WriteHeader(genTarFileHeader(name, int64(csvBongLog.Len()), now))
	if err != nil {
		return fmt.Errorf("tarGz.WriteHeader(): %w", err)
	}
	_, err = io.Copy(tarGz, csvBongLog)
	if err != nil {
		return fmt.Errorf("io.Copy(): %w", err)
	}
	return nil
}

func genTarFileHeader(name string, size int64, now time.Time) *tar.Header {
	return &tar.Header{
		Typeflag:   tar.TypeReg,
		Name:       name,
		Size:       size,
		Mode:       0o660,
		Uid:        1000,
		Gid:        1000,
		Uname:      "bigben",
		Gname:      "bigben",
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}
}

func makeApiUploadRequest(archive *bytes.Buffer, token string) (map[string]any, error) {
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
	req.Header.Set("Authorization", "Bearer "+token)

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
