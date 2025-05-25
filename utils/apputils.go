package utils

import (
	"os"
	"path/filepath"
)

// GetExecutableDir مسیر دایرکتوری فایل اجرایی را برمی‌گرداند.
func GetExecutableDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}
