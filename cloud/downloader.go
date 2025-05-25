package cloud

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath" // اضافه شد برای GetExecutableDir
	"time"
)

// GetExecutableDir مسیر دایرکتوری فایل اجرایی را برمی‌گرداند.
// این تابع از پکیج utils به اینجا منتقل شد.
func GetExecutableDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}

const DefaultDownloadTimeout = 30 * time.Second

// DownloadFile محتوای یک URL را دانلود و در مسیر مقصد ذخیره می‌کند.
func DownloadFile(urlStr, destPath string) error {
	// ... (بقیه کد بدون تغییر) ...
	client := http.Client{
		Timeout: DefaultDownloadTimeout,
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد درخواست HTTP: %w", err)
	}
	req.Header.Set("User-Agent", "OvertimeAppGoClient/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("خطا در ارسال درخواست HTTP به %s: %w", urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("خطا در دانلود فایل: وضعیت سرور %d برای URL %s", resp.StatusCode, urlStr)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد فایل مقصد %s: %w", destPath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("خطا در نوشتن داده‌های دانلود شده در فایل %s: %w", destPath, err)
	}

	return nil
}

// DownloadToTempFile فایل را از URL دانلود کرده و در یک فایل موقت ذخیره می‌کند.
func DownloadToTempFile(urlStr, tempFilePattern string) (string, error) {
	// ... (بقیه کد بدون تغییر) ...
	tempFile, err := os.CreateTemp("", tempFilePattern)
	if err != nil {
		return "", fmt.Errorf("خطا در ایجاد فایل موقت (%s): %w", tempFilePattern, err)
	}
	tempFilePath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		os.Remove(tempFilePath)
		return "", fmt.Errorf("خطا در بستن اولیه فایل موقت %s: %w", tempFilePath, err)
	}

	err = DownloadFile(urlStr, tempFilePath)
	if err != nil {
		os.Remove(tempFilePath)
		return "", fmt.Errorf("خطا در دانلود به فایل موقت %s: %w", tempFilePath, err)
	}

	return tempFilePath, nil
}
