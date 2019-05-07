package main

import (
	"context"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

/**
 * Sample activities used by file processing sample workflow.
 */
const (
	EvaluateFitnessActivityName = "EvaluateFitnessActivity"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		EvaluateFitnessActivity,
		activity.RegisterOptions{Name: EvaluateFitnessActivityName},
	)
}

// func downloadFileActivity(ctx context.Context, fileID string) (*fileInfo, error) {
// 	logger := activity.GetLogger(ctx)
// 	logger.Info("Downloading file...", zap.String("FileID", fileID))
// 	data := downloadFile(fileID)

// 	tmpFile, err := saveToTmpFile(data)
// 	if err != nil {
// 		logger.Error("downloadFileActivity failed to save tmp file.", zap.Error(err))
// 		return nil, err
// 	}

// 	fileInfo := &fileInfo{FileName: tmpFile.Name(), HostID: HostID}
// 	logger.Info("downloadFileActivity succeed.", zap.String("SavedFilePath", fileInfo.FileName))
// 	return fileInfo, nil
// }

func EvaluateFitnessActivity(
	ctx context.Context,
	f func(vec []float64) float64,
	location []float64) (float64, error) {
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("EvaluateFitnessActivity started.")

	value := f(location)
	// processedInfo := &fileInfo{FileName: tmpFile.Name(), HostID: HostID}
	// logger.Info("EvaluateFitnessActivity succeed.", zap.String("SavedFilePath", processedInfo.FileName))
	return value, nil
}

// func uploadFileActivity(ctx context.Context, fInfo fileInfo) error {
// 	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
// 	logger.Info("uploadFileActivity begin.", zap.String("UploadedFileName", fInfo.FileName))

// 	// assert that we are running on the same host as the file was downloaded
// 	// this check is not necessary, just to demo the host specific tasklist is working
// 	if fInfo.HostID != HostID {
// 		logger.Error("uploadFileActivity on wrong host",
// 			zap.String("TargetFile", fInfo.FileName),
// 			zap.String("TargetHostID", fInfo.HostID))
// 		return errors.New("uploadFileActivity running on wrong host")
// 	}

// 	defer os.Remove(fInfo.FileName) // clean up tmp file

// 	err := uploadFile(ctx, fInfo.FileName)
// 	if err != nil {
// 		logger.Error("uploadFileActivity uploading failed.", zap.Error(err))
// 		return err
// 	}
// 	logger.Info("uploadFileActivity succeed.", zap.String("UploadedFileName", fInfo.FileName))
// 	return nil
// }

// func downloadFile(fileID string) []byte {
// 	// dummy downloader
// 	dummyContent := "dummy content for fileID:" + fileID
// 	return []byte(dummyContent)
// }

// func uploadFile(ctx context.Context, filename string) error {
// 	// dummy uploader
// 	_, err := ioutil.ReadFile(filename)
// 	for i := 0; i < 5; i++ {
// 		time.Sleep(1 * time.Second)
// 		// Demonstrates that heartbeat accepts progress data.
// 		// In case of a heartbeat timeout it is included into the error.
// 		activity.RecordHeartbeat(ctx, i)
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func transcodeData(ctx context.Context, data []byte) []byte {
// 	// dummy file processor, just do upper case for the data.
// 	// in real world case, you would want to avoid load entire file content into memory at once.
// 	for i := 0; i < 5; i++ {
// 		time.Sleep(1 * time.Second)
// 		// Demonstrates that heartbeat accepts progress data.
// 		// In case of a heartbeat timeout it is included into the error.
// 		activity.RecordHeartbeat(ctx, i)
// 	}
// 	return []byte(strings.ToUpper(string(data)))
// }

// func saveToTmpFile(data []byte) (f *os.File, err error) {
// 	tmpFile, err := ioutil.TempFile("", "cadence_sample")
// 	if err != nil {
// 		return nil, err
// 	}
// 	_, err = tmpFile.Write(data)
// 	if err != nil {
// 		os.Remove(tmpFile.Name())
// 		return nil, err
// 	}

// 	return tmpFile, nil
// }
