package core

import (
	"fmt"
	"time"

	"github.com/jalaali/go-jalaali"
)

func GetCurrentPersianMonthName() string {
	now := time.Now()
	_, jm, _, _ := jalaali.ToJalaali(now.Year(), now.Month(), now.Day())
	if int(jm) >= 1 && int(jm) <= len(PersianMonthNames) {
		return PersianMonthNames[int(jm)-1]
	}
	fmt.Println("هشدار: ماه شمسی جاری قابل تشخیص نیست، از اولین ماه استفاده می‌شود.")
	return PersianMonthNames[0]
}
