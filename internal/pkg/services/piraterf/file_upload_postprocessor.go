package piraterf

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
)

// fileConversionPostprocessor handles file processing based on module name
// from form data.
func (s *PIrateRF) fileConversionPostprocessor(
	response map[string]any,
	request *http.Request,
) (map[string]any, error) {
	// Get module name from form data
	moduleName := request.FormValue("module")

	// Debug logging
	logrus.WithFields(logrus.Fields{
		"response":   response,
		"moduleName": moduleName,
	}).Info("fileConversionPostprocessor called")

	switch moduleName {
	case gorpitx.ModuleNameFSK:
		return s.dataFilePostprocessor(response)
	case gorpitx.ModuleNamePIFMRDS:
		return s.audioConversionPostprocessor(response)
	case gorpitx.ModuleNameSPECTRUMPAINT, gorpitx.ModuleNamePISSSTV:
		return s.imageConversionPostprocessor(response)
	case gorpitx.ModuleNameSENDIQ:
		return s.iqFilePostprocessor(response)
	default:
		return response, nil
	}
}

// dataFilePostprocessor moves uploaded files to the data directory
// for FSK module.
func (s *PIrateRF) dataFilePostprocessor(
	response map[string]any,
) (map[string]any, error) {
	// Get the file path from the response
	filePath, ok := response["path"].(string)
	if !ok {
		return response, nil // Not a string path, return unchanged
	}

	// Ensure data directory exists
	if err := s.ensureFilesDirsExist(); err != nil {
		return response, ctxerrors.Wrap(err, "failed to ensure directories exist")
	}

	// Generate destination path in data uploads directory
	filename := filepath.Base(filePath)
	destPath := filepath.Join(s.config.FilesDir, dataUploadsPath, filename)

	// Move file to data directory
	if err := moveFile(filePath, destPath); err != nil {
		return response, ctxerrors.Wrapf(
			err,
			"failed to move file to data directory: %s",
			destPath,
		)
	}

	// Update response with new path and mark as moved
	response["path"] = destPath
	response["moved"] = true
	response["destination_directory"] = dataUploadsPath

	return response, nil
}

// iqFilePostprocessor moves uploaded IQ files to the IQ directory
// for SENDIQ module.
func (s *PIrateRF) iqFilePostprocessor(
	response map[string]any,
) (map[string]any, error) {
	// Get the file path from the response
	filePath, ok := response["path"].(string)
	if !ok {
		return response, nil // Not a string path, return unchanged
	}

	// Ensure IQ directory exists
	if err := s.ensureFilesDirsExist(); err != nil {
		return response, ctxerrors.Wrap(err, "failed to ensure directories exist")
	}

	// Generate destination path in IQ uploads directory
	filename := filepath.Base(filePath)
	destPath := filepath.Join(s.config.FilesDir, iqsUploadsPath, filename)

	// Move file to IQ directory
	if err := moveFile(filePath, destPath); err != nil {
		return response, ctxerrors.Wrapf(
			err,
			"failed to move file to IQ directory: %s",
			destPath,
		)
	}

	// Update response with new path and mark as moved
	response["path"] = destPath
	response["moved"] = true
	response["destination_directory"] = iqsUploadsPath

	return response, nil
}

// moveFile moves a file from src to dst, handling cross-device link errors.
func moveFile(src, dst string) error {
	// Try rename first (fastest if on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename failed, try copy + delete
	if err := copyFileStream(src, dst); err != nil {
		return ctxerrors.Wrapf(err, "failed to copy file")
	}

	// Remove source file after successful copy
	if err := os.Remove(src); err != nil {
		return ctxerrors.Wrapf(err, "failed to remove source file after copy")
	}

	return nil
}

// copyFileStream copies a file from src to dst using io.Copy.
func copyFileStream(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to open source file")
	}

	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			logrus.WithError(closeErr).Warn("Failed to close source file")
		}
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to create destination file")
	}

	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			logrus.WithError(closeErr).Warn("Failed to close destination file")
		}
	}()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to copy file content")
	}

	return nil
}
