package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"overtime_go/core" // برای دسترسی به core.User
)

// اطلاعات کاربران پیش‌فرض (رمزها باید در زمان اجرا هش شوند)
var (
	// نقش‌ها: "admin", "department_head"
	// دپارتمان: "all" برای ادمین، نام فارسی واحد برای رؤسا
	defaultRawUsers = map[string]struct {
		Password   string
		Role       string
		Department string
	}{
		"admin":           {"Admin@1371", "admin", "all"},
		"technicaloffice": {"Technicaloffice@3668", "department_head", "دفتر فنی"},
		"hr":              {"Hr@9610", "department_head", "سرمایه های انسانی"},
		"production":      {"Production@1427", "department_head", "تولید"},
		// سایر کاربران پیش‌فرض ...
		// "warehouse": {"Warehouse@8902", "department_head", "انبار"},
		// "security": {"Security@6285", "department_head", "حراست"},
		// "electrical": {"Electrical@3937", "department_head", "برق"},
		// "facilities": {"Facilities@1121", "department_head", "تأسیسات"},
		// "mechanical": {"Mechanical@4451", "department_head", "مکانیک"},
		// "qc": {"Qc@4673", "department_head", "کنترل کیفیت"},
		// "machining": {"Machining@5110", "department_head", "تراشکاری"},
		// "maintenance": {"Maintenance@1064", "department_head", "نت"},
		// "se": {"Se@9743", "department_head", "مهندسی سیستم"},
		// "it": {"It@1277", "department_head", "فناوری اطلاعات"},
		// "planning": {"Planning@4595", "department_head", "برنامه ریزی"},
		// "management": {"Management@6643", "department_head", "مدیریت"},
		// "sales": {"Sales@2817", "department_head", "فروش"},
		// "logistics": {"Logistics@7677", "department_head", "تدارکات"},
		// "finance": {"Finance@3115", "department_head", "مالی"},
		// "hse": {"Hse@4011", "department_head", "HSE"},
		// "technicalengineering": {"Technicalengineering@5432", "department_head", "فنی مهندسی"}, // این نقش خاص است

	}
	// کاربران با رمزهای هش شده در حافظه
	ProcessedUsers = make(map[string]core.User)
)

// HashPassword رمز عبور را با SHA256 هش می‌کند.
func HashPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

// InitializeDefaultUsers رمزهای پیش‌فرض را هش و در ProcessedUsers ذخیره می‌کند.
func InitializeDefaultUsers() {
	for username, u := range defaultRawUsers {
		ProcessedUsers[username] = core.User{
			Username:   username,
			Password:   HashPassword(u.Password),
			Role:       u.Role,
			Department: u.Department,
		}
	}
}

// AuthenticateUser بررسی می‌کند که آیا نام کاربری و رمز عبور معتبر هستند.
func AuthenticateUser(username, password string) (*core.User, bool) {
	user, exists := ProcessedUsers[username]
	if !exists {
		return nil, false
	}

	if user.Password == HashPassword(password) {
		return &user, true
	}
	return nil, false
}
