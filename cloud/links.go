package cloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"overtime_go/core" // این import صحیح است
	// اطمینان حاصل کنید که هیچ import دیگری به overtime_go/utils در اینجا وجود ندارد
)

// ... بقیه کد ...

const cloudLinksFilename = "cloud_links.json"

var (
	loadedCloudLinks map[string]string
)

func LoadCloudLinks() map[string]string {
	if loadedCloudLinks != nil {
		return loadedCloudLinks
	}

	appDir, err := GetExecutableDir() // فراخوانی مستقیم تابع از همین پکیج cloud
	var linksFilePath string
	// ... (بقیه کد LoadCloudLinks بدون تغییر عمده، فقط فراخوانی GetExecutableDir اصلاح می‌شود) ...
	if err == nil {
		linksFilePath = filepath.Join(appDir, cloudLinksFilename)
	} else {
		fmt.Printf("هشدار: خطا در گرفتن مسیر فایل اجرایی برای cloud_links.json: %v. تلاش برای مسیر فعلی.\n", err)
		currentDir, _ := os.Getwd()
		linksFilePath = filepath.Join(currentDir, cloudLinksFilename)
	}

	fileData, err := ioutil.ReadFile(linksFilePath)
	currentLinksFromFile := make(map[string]string)
	useEmbeddedDefaults := false

	if err != nil {
		fmt.Printf("فایل cloud_links.json در مسیر '%s' یافت نشد یا خطا در خواندن (%v). استفاده از لینک‌های پیش‌فرض جاسازی شده.\n", linksFilePath, err)
		useEmbeddedDefaults = true
	} else {
		errJson := json.Unmarshal(fileData, &currentLinksFromFile)
		if errJson != nil {
			fmt.Printf("خطا در پارس کردن cloud_links.json از مسیر '%s' (%v). استفاده از لینک‌های پیش‌فرض جاسازی شده.\n", linksFilePath, errJson)
			useEmbeddedDefaults = true
		}
	}

	embeddedDefaults := core.GetDefaultCloudLinks()
	finalResolvedLinks := make(map[string]string)

	for _, deptShift := range core.ManageableDepartments {
		linkFromFile, foundInFile := currentLinksFromFile[deptShift]
		linkFromEmbedded, foundInEmbedded := embeddedDefaults[deptShift]

		if !useEmbeddedDefaults && foundInFile && strings.TrimSpace(linkFromFile) != "" {
			finalResolvedLinks[deptShift] = linkFromFile
		} else if foundInEmbedded && strings.TrimSpace(linkFromEmbedded) != "" {
			finalResolvedLinks[deptShift] = linkFromEmbedded
		} else {
			finalResolvedLinks[deptShift] = ""
			if deptShift != "" {
				// fmt.Printf("هشدار: هیچ لینکی (نه در فایل، نه پیش‌فرض) برای واحد '%s' یافت نشد.\n", deptShift)
			}
		}
	}
	loadedCloudLinks = finalResolvedLinks
	return loadedCloudLinks
}

func SaveCloudLinks(linksToSave map[string]string) error {
	appDir, err := GetExecutableDir() // فراخوانی مستقیم تابع از همین پکیج cloud
	var linksFilePath string
	// ... (بقیه کد SaveCloudLinks بدون تغییر عمده، فقط فراخوانی GetExecutableDir اصلاح می‌شود) ...
	if err == nil {
		linksFilePath = filepath.Join(appDir, cloudLinksFilename)
	} else {
		fmt.Printf("هشدار: خطا در گرفتن مسیر فایل اجرایی برای ذخیره cloud_links.json: %v. تلاش برای مسیر فعلی.\n", err)
		currentDir, _ := os.Getwd()
		linksFilePath = filepath.Join(currentDir, cloudLinksFilename)
	}

	fileData, err := json.MarshalIndent(linksToSave, "", "  ")
	if err != nil {
		return fmt.Errorf("خطا در تبدیل لینک‌ها به JSON: %w", err)
	}

	err = ioutil.WriteFile(linksFilePath, fileData, 0644)
	if err != nil {
		return fmt.Errorf("خطا در نوشتن فایل cloud_links.json در مسیر '%s': %w", linksFilePath, err)
	}

	newLoadedLinks := make(map[string]string)
	for k, v := range linksToSave {
		newLoadedLinks[k] = v
	}
	loadedCloudLinks = newLoadedLinks

	fmt.Printf("فایل cloud_links.json با موفقیت در مسیر '%s' ذخیره شد.\n", linksFilePath)
	return nil
}

// ConvertToDownloadLink ... (بدون تغییر)
func ConvertToDownloadLink(link string) string {
	dl := strings.TrimSpace(link)
	if strings.Contains(dl, "dropbox.com/s/") {
		dl = strings.Replace(dl, "www.dropbox.com", "dl.dropboxusercontent.com", 1)
		parsedURL, err := url.Parse(dl)
		if err == nil {
			query := parsedURL.Query()
			if query.Get("dl") != "1" {
				query.Set("dl", "1")
				parsedURL.RawQuery = query.Encode()
				return parsedURL.String()
			}
			return dl
		}
		if strings.Contains(dl, "?") && !strings.Contains(dl, "dl=") {
			dl += "&dl=1"
		} else if !strings.Contains(dl, "?") {
			dl += "?dl=1"
		} else {
			dl = strings.Replace(dl, "dl=0", "dl=1", 1)
		}
	} else if strings.Contains(dl, "dropbox.com/scl/") {
		parsedURL, err := url.Parse(dl)
		if err == nil {
			query := parsedURL.Query()
			query.Set("dl", "1")
			parsedURL.RawQuery = query.Encode()
			return parsedURL.String()
		}
	}
	return dl
}
