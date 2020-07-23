package fileprocessing

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
)

/**
 * Sample activities used by file processing sample workflow.
 */

type Activities struct {
	BlobStore *BlobStore
}

func (a *Activities) DownloadFileActivity(ctx context.Context, fileID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading file...", "FileID", fileID)
	data := a.BlobStore.downloadFile(fileID)

	tmpFile, err := saveToTmpFile(data)
	if err != nil {
		logger.Error("downloadFileActivity failed to save tmp file.", "Error", err)
		return "", err
	}
	fileName := tmpFile.Name()
	logger.Info("downloadFileActivity succeed.", "SavedFilePath", fileName)
	return fileName, nil
}

func (a *Activities) ProcessFileActivity(ctx context.Context, fileName string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("processFileActivity started.", "FileName", fileName)

	defer func() { _ = os.Remove(fileName) }() // cleanup temp file

	// read downloaded file
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		logger.Error("processFileActivity failed to read file.", "FileName", fileName, "Error", err)
		return "", err
	}

	// process the file
	transData := transcodeData(ctx, data)
	tmpFile, err := saveToTmpFile(transData)
	if err != nil {
		logger.Error("processFileActivity failed to save tmp file.", "Error", err)
		return "", err
	}

	processedFileName := tmpFile.Name()
	logger.Info("processFileActivity succeed.", "SavedFilePath", processedFileName)
	return processedFileName, nil
}

func (a *Activities) UploadFileActivity(ctx context.Context, fileName string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("uploadFileActivity begin.", "UploadedFileName", fileName)

	defer func() { _ = os.Remove(fileName) }() // cleanup temp file

	err := a.BlobStore.uploadFile(ctx, fileName)
	if err != nil {
		logger.Error("uploadFileActivity uploading failed.", "Error", err)
		return err
	}
	logger.Info("uploadFileActivity succeed.", "UploadedFileName", fileName)
	return nil
}

type BlobStore struct{}

func (b *BlobStore) downloadFile(fileID string) []byte {
	// dummy downloader
	dummyContent := "dummy content for fileID:" + fileID
	return []byte(dummyContent)
}

func (b *BlobStore) uploadFile(ctx context.Context, filename string) error {
	// dummy uploader
	_, err := ioutil.ReadFile(filename)
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		// Demonstrates that heartbeat accepts progress data.
		// In case of a heartbeat timeout it is included into the error.
		activity.RecordHeartbeat(ctx, i)
	}
	if err != nil {
		return err
	}
	return nil
}

func transcodeData(ctx context.Context, data []byte) []byte {
	// dummy file processor, just do upper case for the data.
	// in real world case, you would want to avoid load entire file content into memory at once.
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		// Demonstrates that heartbeat accepts progress data.
		// In case of a heartbeat timeout it is included into the error.
		activity.RecordHeartbeat(ctx, i)
	}
	return []byte(strings.ToUpper(string(data)))
}

func saveToTmpFile(data []byte) (f *os.File, err error) {
	tmpFile, err := ioutil.TempFile("", "temporal_sample")
	if err != nil {
		return nil, err
	}
	_, err = tmpFile.Write(data)
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return nil, err
	}

	return tmpFile, nil
}
