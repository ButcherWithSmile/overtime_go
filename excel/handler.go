package excel

import (
	"fmt"
	"io"
	"overtime_go/core" // اطمینان از صحت نام ماژول
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ReadBasicDataFromExcel(filePath string) (totalHours, prodDays int, monthName string, err error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return 0, 0, "", fmt.Errorf("خطا در باز کردن فایل اکسل %s: %w", filePath, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("خطا در بستن فایل اکسل %s: %v\n", filePath, err)
		}
	}()

	// از اولین شیت استفاده می‌کنیم
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return 0, 0, "", fmt.Errorf("فایل اکسل هیچ شیتی ندارد: %s", filePath)
	}
	sheetName := sheetList[0]

	seranehValStr, errGetSeraneh := f.GetCellValue(sheetName, core.SeranehCell)
	if errGetSeraneh != nil {
		// fmt.Printf("هشدار: خطا در خواندن سلول سرانه %s از شیت '%s': %v. مقدار صفر در نظر گرفته شد.\n", core.SeranehCell, sheetName, errGetSeraneh)
		totalHours = 0 // اگر خطا بود یا خالی بود، صفر در نظر بگیر
	} else {
		// تلاش برای تبدیل به عدد، حتی اگر رشته باشد (مثلا "۱۲۳")
		// excelize ممکن است اعداد را به صورت رشته برگرداند اگر فرمت سلول Text باشد.
		parsedVal, errParse := strconv.ParseFloat(strings.TrimSpace(seranehValStr), 64)
		if errParse == nil && parsedVal >= 0 {
			totalHours = int(parsedVal)
		} else {
			// fmt.Printf("هشدار: مقدار سرانه '%s' در سلول %s شیت '%s' نامعتبر است. مقدار صفر در نظر گرفته شد.\n", seranehValStr, core.SeranehCell, sheetName)
			totalHours = 0
		}
	}

	prodDaysValStr, errGetProd := f.GetCellValue(sheetName, core.ProductionDaysCell)
	if errGetProd != nil {
		// fmt.Printf("هشدار: خطا در خواندن سلول روزهای تولید %s از شیت '%s': %v. مقدار صفر در نظر گرفته شد.\n", core.ProductionDaysCell, sheetName, errGetProd)
		prodDays = 0
	} else {
		parsedVal, errParse := strconv.ParseFloat(strings.TrimSpace(prodDaysValStr), 64)
		if errParse == nil && parsedVal >= 0 {
			prodDays = int(parsedVal)
		} else {
			// fmt.Printf("هشدار: مقدار روزهای تولید '%s' در سلول %s شیت '%s' نامعتبر است. مقدار صفر در نظر گرفته شد.\n", prodDaysValStr, core.ProductionDaysCell, sheetName)
			prodDays = 0
		}
	}

	monthNameVal, errGetMonth := f.GetCellValue(sheetName, core.MonthCell)
	if errGetMonth != nil {
		// fmt.Printf("هشدار: خطا در خواندن سلول ماه %s از شیت '%s': %v.\n", core.MonthCell, sheetName, errGetMonth)
		monthName = ""
	} else {
		monthName = strings.TrimSpace(monthNameVal)
		isValidMonth := false
		for _, m := range core.PersianMonthNames {
			if m == monthName {
				isValidMonth = true
				break
			}
		}
		if !isValidMonth {
			// fmt.Printf("هشدار: نام ماه '%s' در سلول %s شیت '%s' نامعتبر است.\n", monthName, core.MonthCell, sheetName)
			monthName = ""
		}
	}
	// برگرداندن nil برای خطا اگر هیچ خطای بحرانی رخ نداده باشد
	// و فقط مقادیر پیش‌فرض استفاده شده باشند.
	// اگر می‌خواهید برای مقادیر نامعتبر هم خطا برگردانید، باید اینجا خطا را تنظیم کنید.
	// فعلا، اگر فایل باز شود، خطا nil است مگر اینکه خود excelize خطا بدهد.
	return totalHours, prodDays, monthName, nil
}

func ReadEmployeesFromExcel(filePath string, departmentShiftFilter string) ([]core.Employee, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("خطا در باز کردن فایل اکسل %s: %w", filePath, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("خطا در بستن فایل اکسل %s: %v\n", filePath, err)
		}
	}()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("فایل اکسل هیچ شیتی ندارد: %s", filePath)
	}
	sheetName := sheetList[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("خطا در خواندن ردیف‌ها از شیت '%s' در فایل %s: %w", sheetName, filePath, err)
	}

	var employees []core.Employee
	if len(rows) <= 1 {
		return employees, nil
	}

	for rIdx, row := range rows {
		if rIdx == 0 { // نادیده گرفتن ردیف هدر
			continue
		}
		if len(row) <= core.EmployeeDataColID { // حداقل باید ستون‌های مورد نیاز وجود داشته باشند
			// fmt.Printf("ردیف %d در شیت '%s' تعداد ستون کافی ندارد (نیاز به %d، موجود %d).\n", rIdx+1, sheetName, core.EmployeeDataColID+1, len(row))
			continue
		}

		deptFromFile := ""
		if len(row) > core.EmployeeDataColDept { // بررسی وجود ستون
			deptFromFile = strings.TrimSpace(row[core.EmployeeDataColDept])
		}

		if !strings.EqualFold(deptFromFile, departmentShiftFilter) {
			continue
		}

		name := ""
		if len(row) > core.EmployeeDataColName {
			name = strings.TrimSpace(row[core.EmployeeDataColName])
		}
		id := ""
		if len(row) > core.EmployeeDataColID {
			id = strings.TrimSpace(row[core.EmployeeDataColID])
		}

		if name == "" || id == "" || name == "نامشخص" || id == "0000" {
			// fmt.Printf("رد شدن از ردیف با اطلاعات ناقص در شیت '%s': Name='%s', ID='%s'\n", sheetName, name, id)
			continue
		}

		employees = append(employees, core.Employee{
			Name:      name,
			ID:        id,
			Hours:     0,
			Locked:    false,
			MonthType: "",
		})
	}
	return employees, nil
}

func WriteDataToExcel(writer io.Writer, data [][]interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("داده‌ای برای نوشتن وجود ندارد")
	}

	f := excelize.NewFile()
	// اولین شیت به طور خودکار "Sheet1" نام دارد.
	sheetName := "Sheet1"

	for r, rowData := range data {
		for c, cellData := range rowData {
			cellName, err := excelize.CoordinatesToCellName(c+1, r+1)
			if err != nil {
				return fmt.Errorf("خطا در تبدیل مختصات به نام سلول: %w", err)
			}
			// برای اطمینان از اینکه اعداد به صورت عدد ذخیره می‌شوند (اگرچه interface{} می‌تواند هر چیزی باشد)
			// می‌توان نوع داده را بررسی کرد، اما SetCellValue معمولا به درستی عمل می‌کند.
			err = f.SetCellValue(sheetName, cellName, cellData)
			if err != nil {
				return fmt.Errorf("خطا در نوشتن مقدار در سلول %s: %w", cellName, err)
			}
		}
	}

	err := f.Write(writer)
	if err != nil {
		return fmt.Errorf("خطا در نوشتن فایل اکسل در writer: %w", err)
	}
	return nil
}

func GetUniqueDepartmentsFromFile(filePath string) ([]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("خطا در باز کردن فایل اکسل %s: %w", filePath, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("خطا در بستن فایل اکسل %s: %v\n", filePath, err)
		}
	}()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("فایل اکسل هیچ شیتی ندارد: %s", filePath)
	}
	sheetName := sheetList[0]

	cols, err := f.GetCols(sheetName)
	if err != nil {
		return nil, fmt.Errorf("خطا در خواندن ستون‌ها از شیت '%s' فایل %s: %w", sheetName, filePath, err)
	}
	if len(cols) == 0 || len(cols) <= core.EmployeeDataColDept { // باید حداقل ستون دپارتمان وجود داشته باشد
		return nil, fmt.Errorf("ستون نام واحد (اندیس %d) در شیت '%s' فایل %s یافت نشد یا خالی است.", core.EmployeeDataColDept, sheetName, filePath)
	}

	firstCol := cols[core.EmployeeDataColDept]
	uniqueDeptsMap := make(map[string]bool)
	var uniqueDeptsList []string

	for i := 1; i < len(firstCol); i++ { // از ردیف دوم (اندیس 1) شروع می‌کنیم
		deptName := strings.TrimSpace(firstCol[i])
		if deptName != "" && !uniqueDeptsMap[deptName] {
			uniqueDeptsMap[deptName] = true
			uniqueDeptsList = append(uniqueDeptsList, deptName)
		}
	}
	return uniqueDeptsList, nil
}
