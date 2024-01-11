package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/discord-plays/bigben/cmd/bigben/message"
	"github.com/discord-plays/bigben/tables"
	"github.com/disgoorg/snowflake/v2"
	"github.com/mohae/struct2csv"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
	"xorm.io/xorm"
)

func (b *BigBen) cronNewYears() {
	b.messageNotification("New Year's", message.SendNewYearNotification)()
	if b.uploadToken != "" {
		// auto upload the previous year
		generateAndUploadBackup(b.engine, time.Now().Year()-1, b.uploadToken)
	}
}

func generateAndUploadBackup(engine *xorm.Engine, year int, uploadToken string) {
	exist, err := engine.Exist(&tables.LeaderboardUploads{Year: year})
	if err != nil {
		log.Printf("[generateAndUploadBackup(%d)] Failed to check if year has been added to list: %s\n", year, err)
		return
	}
	if !exist {
		_, err := engine.Insert(&tables.LeaderboardUploads{Year: year})
		if err != nil {
			log.Printf("[generateAndUploadBackup(%d)] Failed to add year to uploads list: %s\n", year, err)
			return
		}
	}
	archive, err := prepareApiUpload(engine, year)
	if err != nil {
		log.Printf("[generateAndUploadBackup(%d)] Failed to generate archive: %s\n", year, err)
		return
	}
	create, err := os.Create(fmt.Sprintf("backup-leaderboard-%d.tar.gz", year))
	if err != nil {
		log.Printf("[generateAndUploadBackup(%d)] Failed to create leaderboard backup file: %s\n", year, err)
		return
	}
	defer create.Close()
	_, err = create.Write(archive.Bytes())
	if err != nil {
		log.Printf("[generateAndUploadBackup(%d)] Failed to write leaderboard backup file: %s\n", year, err)
		return
	}
	resp, err := makeApiUploadRequest(archive, uploadToken)
	if err != nil {
		log.Printf("[generateAndUploadBackup(%d)] Failed to upload archive: %s\n", year, err)
		return
	}
	log.Printf(">>> Archive URL %d: https://cdn.mrmelon54.com/download/auto/%s <<<\n", year, resp["Path"])
	sentBool := true
	_, err = engine.Update(&tables.LeaderboardUploads{Year: year, Sent: &sentBool})
	if err != nil {
		log.Printf("[generateAndUploadBackup(%d)] Failed to update sent column in database: %s\n", year, err)
		return
	}
}

func prepareApiUpload(engine *xorm.Engine, year int) (*bytes.Buffer, error) {
	archive := new(bytes.Buffer)
	gz := gzip.NewWriter(archive)
	tarGz := tar.NewWriter(gz)

	err := writeCsvFile[tables.BongLog](tarGz, year, "bong-log.csv", &tables.BongLog{}, engine)
	if err != nil {
		return nil, fmt.Errorf("writeCsvFile(BongLog): %w", err)
	}
	err = writeCsvFile[tables.GuildSettings](tarGz, year, "guild-settings.csv", &tables.GuildSettings{}, engine)
	if err != nil {
		return nil, fmt.Errorf("writeCsvFile(GuildSettings): %w", err)
	}
	err = writeCsvFile[tables.RoleLog](tarGz, year, "role-log.csv", &tables.RoleLog{}, engine)
	if err != nil {
		return nil, fmt.Errorf("writeCsvFile(RoleLog): %w", err)
	}
	err = writeCsvFile[tables.UserLog](tarGz, year, "user-log.csv", &tables.UserLog{}, engine)
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

func writeCsvFile[T any](tarGz *tar.Writer, year int, name string, t *T, engine *xorm.Engine) error {
	now := time.Now()
	startYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endYear := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	startFlake := snowflake.New(startYear)
	endFlake := snowflake.New(endYear)

	csvBongLog := new(bytes.Buffer)
	csvWriter := struct2csv.NewWriter(csvBongLog)
	err := csvWriter.WriteColNames(*t)
	if err != nil {
		return fmt.Errorf("csvWriter.WriteColNames(): %w", err)
	}
	err = engine.Where("msg_id >= ? and msg_id < ?", startFlake, endFlake).Iterate(t, func(idx int, bean interface{}) error {
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
