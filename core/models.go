package core

import (
	"encoding/json"
	"fmt"
	"sort" // برای مرتب‌سازی ManageableDepartments
	// "strings" // اگر نیاز به کار با رشته‌ها باشد
	// "time" // اگر نیاز به تاریخ و زمان باشد
)

// User struct ... (بدون تغییر)
type User struct {
	Username   string
	Password   string
	Role       string
	Department string
}

// Employee struct ... (بدون تغییر)
type Employee struct {
	Name      string
	ID        string
	Hours     int
	Locked    bool
	MonthType string
}

// DepartmentData struct ... (بدون تغییر)
type DepartmentData struct {
	DepartmentShiftName string
	TotalHours          int
	ProductionDays      int
	MonthName           string
	Employees           []Employee
}

var AllDepartmentsData = make(map[string]*DepartmentData)

// CloudLinkInfo struct ... (بدون تغییر)
type CloudLinkInfo struct {
	DepartmentShiftName string `json:"department_shift_name"`
	URL                 string `json:"url"`
}

var PersianMonthNames = []string{"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور", "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند"}

const (
	SeranehCell         = "F1"
	ProductionDaysCell  = "F2"
	MonthCell           = "F3"
	EmployeeDataColDept = 0
	EmployeeDataColName = 1
	EmployeeDataColID   = 2
)

var DepartmentShifts = map[string][]string{
	"انبار":             {"شیفتی", "ثابت"},
	"حراست":             {"نگهبانی", "باسکول"},
	"برق":               {"شیفتی", "ثابت"},
	"تولید":             {"شیفتی", "ثابت"},
	"تأسیسات":           {"شیفتی", "ثابت"},
	"مکانیک":            {"شیفتی", "ثابت"},
	"سرمایه های انسانی": {"شیفتی", "ثابت"},
	"کنترل کیفیت":       {"شیفتی", "ثابت"},
	"تراشکاری":          {"شیفتی"},
	"نت":                {"ثابت"},
	"مهندسی سیستم":      {"ثابت"},
	"فناوری اطلاعات":    {"ثابت"},
	"برنامه ریزی":       {"ثابت"},
	"مدیریت":            {"ثابت"},
	"فروش":              {"ثابت"},
	"دفتر فنی":          {"ثابت"},
	"تدارکات":           {"ثابت"},
	"مالی":              {"ثابت"},
	"HSE":               {"شیفتی", "ثابت"},
	"رؤسا و سرپرستان فنی مهندسی": {"ثابت"},
	"مدیران و رؤسا":              {"ثابت"},
}

var ManageableDepartments []string
var defaultEmbeddedCloudLinks map[string]string // متغیر پکیج برای نگهداری لینک‌های پیش‌فرض

func init() {
	for dept, shifts := range DepartmentShifts {
		for _, shift := range shifts {
			ManageableDepartments = append(ManageableDepartments, dept+" - "+shift)
		}
	}
	sort.Strings(ManageableDepartments) // مرتب‌سازی برای نمایش یکسان

	// مقداردهی اولیه defaultEmbeddedCloudLinks در اینجا انجام نمی‌شود،
	// بلکه از طریق InitializeDefaultCloudLinks که از main.go با داده‌های embed شده فراخوانی می‌شود.
	defaultEmbeddedCloudLinks = make(map[string]string)
}

// InitializeDefaultCloudLinks مقادیر پیش‌فرض لینک‌های ابری را از داده‌های embed شده بارگذاری می‌کند.
// این تابع از main.go فراخوانی می‌شود.
func InitializeDefaultCloudLinks(jsonData []byte) {
	if jsonData == nil {
		fmt.Println("هشدار: داده JSON لینک‌های پیش‌فرض ابری برای مقداردهی اولیه خالی است.")
		return
	}
	err := json.Unmarshal(jsonData, &defaultEmbeddedCloudLinks)
	if err != nil {
		fmt.Printf("خطا در پارس کردن JSON لینک‌های پیش‌فرض ابری هنگام مقداردهی اولیه: %v\n", err)
		// در صورت خطا، defaultEmbeddedCloudLinks خالی باقی می‌ماند.
	} else {
		fmt.Println("لینک‌های پیش‌فرض ابری با موفقیت از داده‌های embed شده بارگذاری شدند.")
	}
}

// GetDefaultCloudLinks یک کپی از لینک‌های پیش‌فرض embed شده را برمی‌گرداند.
func GetDefaultCloudLinks() map[string]string {
	// برگرداندن یک کپی برای جلوگیری از تغییرات ناخواسته در مپ اصلی
	linksCopy := make(map[string]string)
	for k, v := range defaultEmbeddedCloudLinks {
		linksCopy[k] = v
	}
	return linksCopy
}
