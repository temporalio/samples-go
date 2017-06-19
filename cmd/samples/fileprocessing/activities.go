package main

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

/**
 * Sample activities used by file processing sample workflow.
 */

// This is registration process where you register all your activity handlers.
func init() {
	cadence.RegisterActivity(downloadFileActivity)
	cadence.RegisterActivity(processFileActivity)
	cadence.RegisterActivity(uploadFileActivity)
}

func downloadFileActivity(ctx context.Context, fileID string) (*fileInfo, error) {
	logger := cadence.GetActivityLogger(ctx)
	logger.Info("Downloading file...", zap.String("FileID", fileID))
	data := downloadFile(fileID)

	tmpFile, err := saveToTmpFile(data)
	if err != nil {
		logger.Error("downloadFileActivity failed to save tmp file.", zap.Error(err))
		return nil, err
	}

	fileInfo := &fileInfo{FileName: tmpFile.Name(), HostID: HostID}
	logger.Info("downloadFileActivity succeed.", zap.String("SavedFilePath", fileInfo.FileName))
	return fileInfo, nil
}

func processFileActivity(ctx context.Context, fInfo fileInfo) (*fileInfo, error) {
	logger := cadence.GetActivityLogger(ctx).With(zap.String("HostID", HostID))
	// assert that we are running on the same host as the file was downloaded
	// this check is not necessary, just to demo the host specific tasklist is working
	if fInfo.HostID != HostID {
		logger.Error("processFileActivity on wrong host",
			zap.String("TargetFile", fInfo.FileName),
			zap.String("TargetHostID", fInfo.HostID))
		return nil, errors.New("processFileActivity running on wrong host")
	}

	defer os.Remove(fInfo.FileName) // cleanup temp file

	// read downloaded file
	data, err := ioutil.ReadFile(fInfo.FileName)
	if err != nil {
		logger.Error("processFileActivity failed to read file.", zap.String("FileName", fInfo.FileName), zap.Error(err))
		return nil, err
	}

	// process the file
	transData := transcodeData(data)
	tmpFile, err := saveToTmpFile(transData)
	if err != nil {
		logger.Error("processFileActivity failed to save tmp file.", zap.Error(err))
		return nil, err
	}

	processedInfo := &fileInfo{FileName: tmpFile.Name(), HostID: HostID}
	logger.Info("processFileActivity succeed.", zap.String("SavedFilePath", processedInfo.FileName))
	return processedInfo, nil
}

func uploadFileActivity(ctx context.Context, fInfo fileInfo) error {
	logger := cadence.GetActivityLogger(ctx).With(zap.String("HostID", HostID))
	// assert that we are running on the same host as the file was downloaded
	// this check is not necessary, just to demo the host specific tasklist is working
	if fInfo.HostID != HostID {
		logger.Error("uploadFileActivity on wrong host",
			zap.String("TargetFile", fInfo.FileName),
			zap.String("TargetHostID", fInfo.HostID))
		return errors.New("uploadFileActivity running on wrong host")
	}

	defer os.Remove(fInfo.FileName) // clean up tmp file

	err := uploadFile(fInfo.FileName)
	if err != nil {
		logger.Error("uploadFileActivity uploading failed.", zap.Error(err))
		return err
	}
	logger.Info("uploadFileActivity succeed.", zap.String("UploadedFileName", fInfo.FileName))
	return nil
}

func downloadFile(fileID string) []byte {
	// dummy downloader
	dummyContent := "dummy content for fileID:" + fileID
	return []byte(dummyContent)
}

func uploadFile(filename string) error {
	// dummy uploader
	_, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return nil
}

func transcodeData(data []byte) []byte {
	// dummy file processor, just do upper case for the data.
	// in real world case, you would want to avoid load entire file content into memory at once.
	return []byte(strings.ToUpper(string(data)))
}

func saveToTmpFile(data []byte) (f *os.File, err error) {
	tmpFile, err := ioutil.TempFile("", "cadence_sample")
	if err != nil {
		return nil, err
	}
	_, err = tmpFile.Write(data)
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, err
	}

	return tmpFile, nil
}
