package gui

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"overtime_go/cloud"
	"overtime_go/core"
	"overtime_go/excel"
	"overtime_go/resources"
)

type MainUI struct {
	Window fyne.Window
	App    fyne.App
	User   *core.User

	currentDepartmentData *core.DepartmentData
	currentEmployees      binding.UntypedList

	deptComboBox        *widget.Select
	totalHoursInput     *widget.Entry
	productionDaysInput *widget.Entry
	employeesTable      *widget.Table
	summaryLabel        *widget.Label
	exportButton        *widget.Button
	updateCloudButton   *widget.Button
	manageLinksButton   *widget.Button
	resetButton         *widget.Button

	adminManualEmployeesInput *widget.Entry
	adminCreateTableButton    *widget.Button
	adminImportExcelButton    *widget.Button

	logoutHandler func()
}

func CreateMainUI(win fyne.Window, appInst fyne.App, user *core.User, logoutCallback func()) fyne.CanvasObject {
	ui := &MainUI{
		Window:           win,
		App:              appInst,
		User:             user,
		currentEmployees: binding.NewUntypedList(),
		logoutHandler:    logoutCallback,
	}

	topControls := ui.createTopControls()

	ui.employeesTable = widget.NewTable(
		ui.tableDataLength,
		ui.tableCreateCell,
		ui.tableUpdateCell,
	)
	ui.employeesTable.SetColumnWidth(0, 220)
	ui.employeesTable.SetColumnWidth(1, 100)
	ui.employeesTable.SetColumnWidth(2, 160)
	ui.employeesTable.SetColumnWidth(3, 130)
	ui.employeesTable.SetColumnWidth(4, 70)

	tableContainer := container.NewScroll(ui.employeesTable)

	ui.summaryLabel = widget.NewLabel("مجموع ساعات: 0 / سرانه: 0")
	ui.summaryLabel.Alignment = fyne.TextAlignCenter

	tableGroupCard := widget.NewCard("تخصیص ساعات اضافه کاری پرسنل", "",
		container.NewBorder(nil, ui.summaryLabel, nil, nil, tableContainer),
	)

	bottomButtons := ui.createBottomButtons()

	mainContent := container.NewBorder(
		topControls,
		bottomButtons,
		nil,
		nil,
		tableGroupCard,
	)

	accessibleDepts := ui.getAccessibleDepartmentShifts()
	if len(accessibleDepts) > 0 && ui.deptComboBox != nil {
		ui.loadDepartmentDataByName(accessibleDepts[0])
		ui.deptComboBox.Selected = accessibleDepts[0]
		ui.refreshUIForCurrentDepartment()
	} else if ui.deptComboBox != nil {
		ui.deptComboBox.SetSelected("")
		ui.clearUIForNoDepartment()
	}
	ui.updateSummaryLabel()

	return mainContent
}

func (ui *MainUI) refreshUIForCurrentDepartment() {
	if ui.currentDepartmentData != nil {
		ui.totalHoursInput.SetText(strconv.Itoa(ui.currentDepartmentData.TotalHours))
		ui.productionDaysInput.SetText(strconv.Itoa(ui.currentDepartmentData.ProductionDays))

		var empInterfaces []interface{}
		for i := range ui.currentDepartmentData.Employees {
			empInterfaces = append(empInterfaces, &ui.currentDepartmentData.Employees[i])
		}
		ui.currentEmployees.Set(empInterfaces)

		if ui.User.Role != "admin" {
			ui.totalHoursInput.Disable()
		} else {
			ui.totalHoursInput.Enable()
		}

		enableActions := len(ui.currentDepartmentData.Employees) > 0
		if enableActions {
			ui.exportButton.Enable()
		} else {
			ui.exportButton.Disable()
		}
		ui.resetButton.Enable()

		if ui.updateCloudButton != nil {
			if ui.currentDepartmentData != nil || (ui.deptComboBox != nil && ui.deptComboBox.Selected != "" && ui.deptComboBox.Selected != ui.deptComboBox.PlaceHolder) {
				ui.updateCloudButton.Enable()
			} else {
				ui.updateCloudButton.Disable()
			}
		}

		if ui.User.Role == "admin" && ui.adminManualEmployeesInput != nil {
			enableAdminControls := ui.currentDepartmentData != nil || (ui.deptComboBox != nil && ui.deptComboBox.Selected != "" && ui.deptComboBox.Selected != ui.deptComboBox.PlaceHolder)
			if enableAdminControls {
				ui.adminManualEmployeesInput.Enable()
				ui.adminCreateTableButton.Enable()
				ui.adminImportExcelButton.Enable()
			} else {
				ui.adminManualEmployeesInput.Disable()
				ui.adminCreateTableButton.Disable()
				ui.adminImportExcelButton.Disable()
			}
		}
	} else {
		ui.clearUIForNoDepartment()
	}
	ui.employeesTable.Refresh()
	ui.updateSummaryLabel()
}

func (ui *MainUI) createTopControls() fyne.CanvasObject {
	deptLabel := widget.NewLabel("واحد سازمانی:")
	accessibleDepts := ui.getAccessibleDepartmentShifts()
	ui.deptComboBox = widget.NewSelect(accessibleDepts, func(selected string) {
		ui.onDepartmentChanged(selected)
	})
	if len(accessibleDepts) == 0 {
		ui.deptComboBox.PlaceHolder = "-- هیچ واحدی قابل دسترسی نیست --"
		ui.deptComboBox.Disable()
	} else {
		ui.deptComboBox.PlaceHolder = "یک واحد را انتخاب کنید"
	}
	seranehLabel := widget.NewLabel("سرانه کل (ساعت):")
	ui.totalHoursInput = widget.NewEntry()
	ui.totalHoursInput.SetText("0")
	// ui.totalHoursInput.Alignment حذف شد
	ui.totalHoursInput.MultiLine = false
	ui.totalHoursInput.Wrapping = fyne.TextTruncate

	if ui.User.Role != "admin" {
		ui.totalHoursInput.Disable()
	} else {
		ui.totalHoursInput.OnChanged = func(s string) {
			if ui.currentDepartmentData == nil && s != "0" && s != "" {
				ui.totalHoursInput.SetText("0")
				dialog.ShowInformation("راهنما", "لطفا ابتدا یک واحد را انتخاب کنید تا سرانه آن را تغییر دهید.", ui.Window)
				return
			}
			if ui.currentDepartmentData == nil {
				return
			}
			newTotal, err := strconv.Atoi(s)
			originalText := strconv.Itoa(ui.currentDepartmentData.TotalHours)
			if err == nil && newTotal >= 0 {
				if ui.currentDepartmentData.TotalHours != newTotal {
					ui.currentDepartmentData.TotalHours = newTotal
					ui.reallocateHours()
				}
			} else if s != "" {
				ui.totalHoursInput.SetText(originalText)
				dialog.ShowError(fmt.Errorf("مقدار سرانه باید عدد صحیح غیرمنفی باشد"), ui.Window)
			} else if s == "" {
				if ui.currentDepartmentData.TotalHours != 0 {
					ui.currentDepartmentData.TotalHours = 0
					ui.reallocateHours()
				}
			}
		}
	}
	prodDaysLabel := widget.NewLabel("روزهای تولید:")
	ui.productionDaysInput = widget.NewEntry()
	ui.productionDaysInput.SetText("0")
	// ui.productionDaysInput.Alignment حذف شد
	ui.productionDaysInput.Disable()
	baseControls := container.New(layout.NewFormLayout(),
		deptLabel, ui.deptComboBox,
		seranehLabel, ui.totalHoursInput,
		prodDaysLabel, ui.productionDaysInput,
	)
	if ui.User.Role == "admin" {
		adminTitle := widget.NewLabelWithStyle("کنترل‌های ادمین:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

		adminEmpCountLabel := widget.NewLabel("تعداد پرسنل (جدول دستی):")
		ui.adminManualEmployeesInput = widget.NewEntry()
		ui.adminManualEmployeesInput.SetPlaceHolder("مثلا: 5")
		ui.adminManualEmployeesInput.Validator = func(s string) error {
			if s == "" {
				return nil
			}
			val, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("باید عدد باشد")
			}
			if val <= 0 {
				return fmt.Errorf("باید مثبت باشد")
			}
			return nil
		}
		ui.adminCreateTableButton = widget.NewButtonWithIcon("ایجاد جدول دستی", theme.ContentAddIcon(), ui.onAdminCreateTable)
		ui.adminImportExcelButton = widget.NewButtonWithIcon("وارد کردن دستی اکسل", theme.DocumentIcon(), ui.onAdminImportExcel)

		adminSpecificControls := container.NewVBox(
			adminTitle,
			container.New(layout.NewFormLayout(),
				adminEmpCountLabel, ui.adminManualEmployeesInput,
			),
			container.NewGridWithColumns(2, ui.adminCreateTableButton, ui.adminImportExcelButton),
		)

		enableAdminControls := len(accessibleDepts) > 0 && ui.deptComboBox.Selected != "" && ui.deptComboBox.Selected != ui.deptComboBox.PlaceHolder
		if enableAdminControls {
			ui.adminManualEmployeesInput.Enable()
			ui.adminCreateTableButton.Enable()
			ui.adminImportExcelButton.Enable()
		} else {
			ui.adminManualEmployeesInput.Disable()
			ui.adminCreateTableButton.Disable()
			ui.adminImportExcelButton.Disable()
		}
		return container.NewVBox(baseControls, widget.NewSeparator(), adminSpecificControls)
	}
	return baseControls
}
func (ui *MainUI) createBottomButtons() fyne.CanvasObject {
	ui.exportButton = widget.NewButtonWithIcon("خروجی اکسل", theme.DocumentSaveIcon(), ui.onExportToExcel) // آیکن اصلاح شد
	ui.resetButton = widget.NewButtonWithIcon("پاک کردن جدول", theme.ContentClearIcon(), ui.onResetTable)
	helpButton := widget.NewButtonWithIcon("راهنما", theme.HelpIcon(), ui.onShowHelp)
	aboutButton := widget.NewButtonWithIcon("درباره", theme.InfoIcon(), ui.onShowAbout)
	logoutButton := widget.NewButtonWithIcon("خروج از حساب", theme.LogoutIcon(), ui.onLogout)
	var leftButtonWidgets []fyne.CanvasObject
	if ui.User.Role == "admin" {
		ui.manageLinksButton = widget.NewButtonWithIcon("مدیریت لینک‌ها", theme.SettingsIcon(), ui.onManageCloudLinks)
		leftButtonWidgets = append(leftButtonWidgets, ui.manageLinksButton)
	} else {
		ui.updateCloudButton = widget.NewButtonWithIcon("به‌روزرسانی از سرور", theme.DownloadIcon(), ui.onUpdateFromCloud)
		leftButtonWidgets = append(leftButtonWidgets, ui.updateCloudButton)
	}
	leftButtonWidgets = append(leftButtonWidgets, ui.exportButton)
	rightButtonWidgetsElements := []fyne.CanvasObject{ui.resetButton, helpButton, aboutButton, logoutButton}

	ui.exportButton.Disable()
	ui.resetButton.Disable()
	if ui.updateCloudButton != nil {
		ui.updateCloudButton.Disable()
	}
	if ui.manageLinksButton != nil {
		ui.manageLinksButton.Enable()
	}

	// اصلاح خطای too many arguments
	var combinedRightWidgets []fyne.CanvasObject
	combinedRightWidgets = append(combinedRightWidgets, layout.NewSpacer(), widget.NewSeparator())
	combinedRightWidgets = append(combinedRightWidgets, rightButtonWidgetsElements...)

	return container.NewBorder(
		nil, nil,
		container.NewHBox(leftButtonWidgets...),
		container.NewHBox(combinedRightWidgets...),
	)
}
func (ui *MainUI) getAccessibleDepartmentShifts() []string {
	var accessible []string
	if ui.User.Role == "admin" {
		return core.ManageableDepartments
	}
	baseDept := ui.User.Department
	if baseDept == "فنی مهندسی" {
		accessible = []string{
			"تراشکاری - شیفتی", "دفتر فنی - ثابت", "برق - ثابت", "برق - شیفتی",
			"مکانیک - ثابت", "مکانیک - شیفتی", "نت - ثابت", "تأسیسات - ثابت",
			"تأسیسات - شیفتی", "رؤسا و سرپرستان فنی مهندسی - ثابت",
		}
	} else if baseDept == "سرمایه های انسانی" {
		accessible = []string{
			"سرمایه های انسانی - ثابت", "سرمایه های انسانی - شیفتی", "مدیران و رؤسا - ثابت",
		}
	} else {
		if shifts, ok := core.DepartmentShifts[baseDept]; ok {
			for _, shift := range shifts {
				accessible = append(accessible, baseDept+" - "+shift)
			}
		} else {
			fmt.Printf("خطا: واحد سازمانی '%s' برای کاربر تعریف نشده است.\n", baseDept)
		}
	}
	sort.Strings(accessible)
	return accessible
}
func (ui *MainUI) onDepartmentChanged(selectedDeptShift string) {
	if selectedDeptShift == "" || selectedDeptShift == ui.deptComboBox.PlaceHolder || selectedDeptShift == "-- هیچ واحدی قابل دسترسی نیست --" {
		ui.clearUIForNoDepartment()
		if ui.User.Role == "admin" && ui.adminManualEmployeesInput != nil {
			ui.adminManualEmployeesInput.Disable()
			ui.adminCreateTableButton.Disable()
			ui.adminImportExcelButton.Disable()
		}
		return
	}

	fmt.Println("واحد انتخاب شده توسط کاربر:", selectedDeptShift)
	ui.loadDepartmentDataByName(selectedDeptShift)
	ui.refreshUIForCurrentDepartment()
	if ui.User.Role == "admin" && ui.adminManualEmployeesInput != nil {
		ui.adminManualEmployeesInput.Enable()
		ui.adminCreateTableButton.Enable()
		ui.adminImportExcelButton.Enable()
	}
}
func (ui *MainUI) loadDepartmentDataByName(deptShiftName string) {
	data, exists := core.AllDepartmentsData[deptShiftName]
	if !exists {
		fmt.Printf("داده‌ای برای واحد '%s' در حافظه یافت نشد. یک ورودی جدید ایجاد می‌شود.\n", deptShiftName)
		newData := &core.DepartmentData{
			DepartmentShiftName: deptShiftName,
			TotalHours:          0,
			ProductionDays:      0,
			MonthName:           "",
			Employees:           []core.Employee{},
		}
		core.AllDepartmentsData[deptShiftName] = newData
		ui.currentDepartmentData = newData
	} else {
		ui.currentDepartmentData = data
	}
}
func (ui *MainUI) clearUIForNoDepartment() {
	ui.currentDepartmentData = nil
	ui.currentEmployees.Set(nil)
	ui.totalHoursInput.SetText("0")
	ui.productionDaysInput.SetText("0")
	if ui.User.Role != "admin" {
		ui.totalHoursInput.Disable()
	} else {
		ui.totalHoursInput.SetText("0")
		ui.totalHoursInput.Disable()
	}
	ui.productionDaysInput.Disable()
	ui.employeesTable.Refresh()
	ui.updateSummaryLabel()
	ui.exportButton.Disable()
	ui.resetButton.Disable()
	if ui.updateCloudButton != nil {
		ui.updateCloudButton.Disable()
	}
	if ui.adminManualEmployeesInput != nil {
		ui.adminManualEmployeesInput.Disable()
	}
	if ui.adminCreateTableButton != nil {
		ui.adminCreateTableButton.Disable()
	}
	if ui.adminImportExcelButton != nil {
		ui.adminImportExcelButton.Disable()
	}
}
func (ui *MainUI) tableDataLength() (rows int, cols int) {
	length := ui.currentEmployees.Length() // اصلاح شد
	return length, 5
}
func (ui *MainUI) tableCreateCell() fyne.CanvasObject {
	return container.NewMax()
}
func (ui *MainUI) tableUpdateCell(id widget.TableCellID, template fyne.CanvasObject) {
	rawItem, err := ui.currentEmployees.GetValue(id.Row)
	if err != nil {
		template.(*fyne.Container).Objects = []fyne.CanvasObject{widget.NewLabel("")}
		template.(*fyne.Container).Refresh()
		return
	}

	emp, ok := rawItem.(*core.Employee)
	if !ok || emp == nil {
		fyne.LogError("Item in list is not a valid *core.Employee or is nil", nil)
		template.(*fyne.Container).Objects = []fyne.CanvasObject{widget.NewLabel("خطا")}
		template.(*fyne.Container).Refresh()
		return
	}
	cellContainer := template.(*fyne.Container)
	switch id.Col {
	case 0:
		nameLabel := widget.NewLabel(emp.Name)
		cellContainer.Objects = []fyne.CanvasObject{nameLabel}
	case 1:
		idLabel := widget.NewLabel(emp.ID)
		cellContainer.Objects = []fyne.CanvasObject{idLabel}
	case 2:
		entry := widget.NewEntry()
		entry.SetText(strconv.Itoa(emp.Hours))
		entry.Validator = func(s string) error {
			if s == "" {
				return fmt.Errorf("ساعت نمی‌تواند خالی باشد")
			}
			val, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("باید عدد باشد")
			}
			if val < 0 || val > 999 {
				return fmt.Errorf("ساعت باید بین 0 و 999 باشد")
			}
			return nil
		}
		var isUpdatingHoursFromEntry bool
		entry.OnChanged = func(s string) {
			if isUpdatingHoursFromEntry {
				return
			}

			if entry.Validate() == nil {
				newHours, _ := strconv.Atoi(s)
				if emp.Hours != newHours {
					isUpdatingHoursFromEntry = true
					emp.Hours = newHours
					emp.Locked = true
					ui.reallocateHours()
					isUpdatingHoursFromEntry = false
				}
			}
		}
		// OnFocusLost حذف شد
		cellContainer.Objects = []fyne.CanvasObject{entry}
	case 3:
		monthToDisplay := emp.MonthType
		if monthToDisplay == "" && ui.currentDepartmentData != nil {
			monthToDisplay = ui.currentDepartmentData.MonthName
		}
		if monthToDisplay == "" {
			monthToDisplay = core.GetCurrentPersianMonthName()
		}
		monthLabel := widget.NewLabel(monthToDisplay)
		monthLabel.Alignment = fyne.TextAlignCenter
		cellContainer.Objects = []fyne.CanvasObject{monthLabel}
	case 4:
		check := widget.NewCheck("", func(locked bool) {
			if emp.Locked != locked {
				emp.Locked = locked
				ui.reallocateHours()
			}
		})
		check.SetChecked(emp.Locked)
		centeredCheck := container.NewCenter(check)
		cellContainer.Objects = []fyne.CanvasObject{centeredCheck}
	}
	cellContainer.Refresh()
}
func (ui *MainUI) reallocateHours() {
	if ui.currentDepartmentData == nil || len(ui.currentDepartmentData.Employees) == 0 {
		ui.updateSummaryLabel()
		if ui.currentDepartmentData == nil || len(ui.currentDepartmentData.Employees) == 0 {
			ui.currentEmployees.Set(nil)
		}
		ui.employeesTable.Refresh()
		return
	}
	lockedSum := 0
	var unlockedIndices []int
	for i := range ui.currentDepartmentData.Employees {
		if ui.currentDepartmentData.Employees[i].Locked {
			lockedSum += ui.currentDepartmentData.Employees[i].Hours
		} else {
			unlockedIndices = append(unlockedIndices, i)
		}
	}
	remainingTotalHours := ui.currentDepartmentData.TotalHours - lockedSum
	if remainingTotalHours < 0 {
		remainingTotalHours = 0
	}
	numUnlocked := len(unlockedIndices)
	baseHoursPerUnlocked := 0
	extraHoursCount := 0
	if numUnlocked > 0 {
		baseHoursPerUnlocked = remainingTotalHours / numUnlocked
		extraHoursCount = remainingTotalHours % numUnlocked
	} else if remainingTotalHours > 0 {
		fmt.Println("هشدار: تمام پرسنل قفل هستند اما هنوز ساعت برای توزیع باقی مانده است.")
	}
	extraDistributed := 0
	for _, idx := range unlockedIndices {
		allocatedHours := baseHoursPerUnlocked
		if extraDistributed < extraHoursCount {
			allocatedHours++
		}
		extraDistributed++
		ui.currentDepartmentData.Employees[idx].Hours = allocatedHours
	}
	var empInterfaces []interface{}
	for i := range ui.currentDepartmentData.Employees {
		empInterfaces = append(empInterfaces, &ui.currentDepartmentData.Employees[i])
	}
	err := ui.currentEmployees.Set(empInterfaces)
	if err != nil {
		fyne.LogError("Failed to set new data to currentEmployees list", err)
	}
	ui.employeesTable.Refresh()
	ui.updateSummaryLabel()
}
func (ui *MainUI) updateSummaryLabel() {
	if ui.currentDepartmentData == nil {
		ui.summaryLabel.SetText("مجموع ساعات: 0 / سرانه: 0")
		ui.summaryLabel.Refresh()
		return
	}
	currentAllocatedHours := 0
	for _, emp := range ui.currentDepartmentData.Employees {
		currentAllocatedHours += emp.Hours
	}
	targetHours := ui.currentDepartmentData.TotalHours
	text := fmt.Sprintf("مجموع ساعات: %d / سرانه: %d", currentAllocatedHours, targetHours)
	ui.summaryLabel.SetText(text)
	ui.summaryLabel.Refresh()
}
func (ui *MainUI) onUpdateFromCloud() {
	if ui.currentDepartmentData == nil || ui.currentDepartmentData.DepartmentShiftName == "" {
		dialog.ShowInformation("راهنما", "ابتدا یک واحد سازمانی را انتخاب کنید.", ui.Window)
		return
	}
	deptShiftName := ui.currentDepartmentData.DepartmentShiftName
	dialog.ShowConfirm("تأیید به‌روزرسانی", fmt.Sprintf("آیا می‌خواهید اطلاعات واحد '%s' را از سرور به‌روزرسانی کنید؟", deptShiftName), func(confirm bool) {
		if !confirm {
			return
		}
		progress := dialog.NewProgressInfinite("در حال دریافت اطلاعات", "لطفاً منتظر بمانید...", ui.Window)
		progress.Show()
		go func() {
			defer progress.Hide()
			allLinks := cloud.LoadCloudLinks()
			link, ok := allLinks[deptShiftName]
			if !ok || link == "" {
				dialog.ShowError(fmt.Errorf("لینک دانلود برای واحد '%s' یافت نشد", deptShiftName), ui.Window)
				return
			}
			downloadURL := cloud.ConvertToDownloadLink(link)

			tempFilePath, err := cloud.DownloadToTempFile(downloadURL, "cloud_dl_*.xlsx")
			if err != nil {
				dialog.ShowError(fmt.Errorf("خطا در دانلود فایل از %s: %w", downloadURL, err), ui.Window)
				return
			}
			defer os.Remove(tempFilePath)
			seraneh, prodDays, monthF3, err := excel.ReadBasicDataFromExcel(tempFilePath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("خطا در خواندن اطلاعات پایه از اکسل (%s): %w", filepath.Base(tempFilePath), err), ui.Window)
				return
			}
			employees, err := excel.ReadEmployeesFromExcel(tempFilePath, deptShiftName)
			if err != nil {
				dialog.ShowError(fmt.Errorf("خطا در خواندن لیست پرسنل از اکسل (%s) برای واحد '%s': %w", filepath.Base(tempFilePath), deptShiftName, err), ui.Window)
				return
			}

			finalMonthName := monthF3
			if finalMonthName == "" {
				finalMonthName = core.GetCurrentPersianMonthName()
				dialog.ShowInformation("هشدار ماه", fmt.Sprintf("مقدار ماه در فایل اکسل (سلول %s) نامعتبر یا خالی است.\n از ماه جاری سیستم (%s) استفاده خواهد شد.", core.MonthCell, finalMonthName), ui.Window)
			}
			// Update UI directly
			ui.currentDepartmentData.Employees = employees
			ui.totalHoursInput.SetText(strconv.Itoa(seraneh))
			ui.productionDaysInput.SetText(strconv.Itoa(prodDays))
			ui.reallocateHours()
			ui.exportButton.Enable()
			dialog.ShowInformation("موفقیت", fmt.Sprintf("اطلاعات واحد '%s' با موفقیت به‌روز شد.", deptShiftName), ui.Window)
		}()
	}, ui.Window)
}
func (ui *MainUI) onExportToExcel() {
	if ui.currentDepartmentData == nil || len(ui.currentDepartmentData.Employees) == 0 {
		dialog.ShowInformation("خطا", "داده‌ای برای خروجی گرفتن وجود ندارد.", ui.Window)
		return
	}
	currentAllocated := 0
	for _, emp := range ui.currentDepartmentData.Employees {
		currentAllocated += emp.Hours
	}
	if currentAllocated != ui.currentDepartmentData.TotalHours {
		dialog.ShowInformation("خطا در تخصیص", fmt.Sprintf("مجموع ساعات تخصیص یافته (%d) با سرانه کل (%d) برابر نیست. لطفاً ساعات را بررسی کنید.", currentAllocated, ui.currentDepartmentData.TotalHours), ui.Window)
		return
	}
	monthForFile := ui.currentDepartmentData.MonthName
	if monthForFile == "" {
		monthForFile = core.GetCurrentPersianMonthName()
	}
	dateStr := time.Now().Format("2006-01-02")
	defaultFileName := fmt.Sprintf("%s - %s - %s.xlsx", ui.currentDepartmentData.DepartmentShiftName, monthForFile, dateStr)
	defaultFileName = strings.ReplaceAll(strings.ReplaceAll(defaultFileName, "/", "_"), "\\", "_")
	fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, errDialog error) {
		if errDialog != nil {
			dialog.ShowError(errDialog, ui.Window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		dataForExcel := [][]interface{}{
			{"نام واحد", "نام پرسنل", "کد پرسنلی", "ساعت اضافه کاری", "ماه"},
		}
		deptName := ui.currentDepartmentData.DepartmentShiftName
		monthAssigned := ui.currentDepartmentData.MonthName
		if monthAssigned == "" {
			monthAssigned = core.GetCurrentPersianMonthName()
		}
		for _, emp := range ui.currentDepartmentData.Employees {
			dataForExcel = append(dataForExcel, []interface{}{
				deptName, emp.Name, emp.ID, emp.Hours, monthAssigned,
			})
		}
		errWrite := excel.WriteDataToExcel(writer, dataForExcel)
		if errWrite != nil {
			dialog.ShowError(fmt.Errorf("خطا در ذخیره فایل اکسل: %w", errWrite), ui.Window)
			return
		}
		dialog.ShowInformation("موفقیت", "فایل اکسل با موفقیت ذخیره شد:\n"+writer.URI().Path(), ui.Window)
	}, ui.Window)
	fileSaveDialog.SetFileName(defaultFileName)
	fileSaveDialog.Show()
}
func (ui *MainUI) onResetTable() {
	if ui.currentDepartmentData == nil || ui.currentDepartmentData.DepartmentShiftName == "" {
		dialog.ShowInformation("راهنما", "ابتدا یک واحد سازمانی را انتخاب کنید.", ui.Window)
		return
	}
	deptName := ui.currentDepartmentData.DepartmentShiftName
	dialog.ShowConfirm("تأیید پاک کردن", fmt.Sprintf("آیا مطمئن هستید که می‌خواهید جدول واحد '%s' را پاک کنید؟\n(سرانه، روزهای تولید و لیست پرسنل صفر خواهند شد)", deptName), func(confirm bool) {
		if !confirm {
			return
		}
		originalData, ok := core.AllDepartmentsData[deptName]
		if !ok {
			dialog.ShowError(fmt.Errorf("خطای داخلی: داده‌های واحد '%s' برای ریست یافت نشد.", deptName), ui.Window)
			return
		}
		originalData.TotalHours = 0
		originalData.ProductionDays = 0
		originalData.MonthName = ""
		originalData.Employees = []core.Employee{}
		ui.totalHoursInput.SetText("0")
		ui.productionDaysInput.SetText("0")
		ui.currentEmployees.Set(nil)
		ui.reallocateHours()
		ui.exportButton.Disable()
		dialog.ShowInformation("پاک شد", fmt.Sprintf("جدول واحد '%s' با موفقیت پاک شد.", deptName), ui.Window)
	}, ui.Window)
}
func (ui *MainUI) onShowHelp() {
	seranehInfo := fmt.Sprintf("(از سلول %s)", core.SeranehCell)
	prodDaysInfo := fmt.Sprintf("(از سلول %s)", core.ProductionDaysCell)
	monthInfo := fmt.Sprintf("(از سلول %s)", core.MonthCell)

	var helpText string
	if ui.User.Role == "admin" {
		helpText = fmt.Sprintf(`راهنمای مدیر:
1. لینک‌ها: تنظیم لینک دانلود اکسل واحدها (از طریق دکمه "مدیریت لینک‌ها").
2. فایل‌های اکسل ورودی: باید شامل ستون A برای "نام واحد"، B برای "نام پرسنل" و C برای "کد پرسنلی" باشند. سرانه کل در %s، روزهای تولید در %s و نام ماه تخصیص در %s فایل اکسل قرار گیرد.
3. ورود دستی/اکسل: برای وارد کردن اطلاعات به صورت دستی یا از طریق فایل اکسل.
4. ویرایش سرانه: سرانه کل برای واحد انتخاب شده توسط ادمین قابل ویرایش است.
5. بررسی و خروجی: مشاهده و بررسی تخصیص‌ها. خروجی اکسل (ماه بر اساس %s).
6. پاک کردن جدول: حذف کامل اطلاعات برای واحد انتخاب شده.`,
			core.SeranehCell, core.ProductionDaysCell, core.MonthCell, core.MonthCell)
	} else {
		helpText = fmt.Sprintf(`راهنمای بالاترین مقام واحد:
1. دریافت اطلاعات: با کلیک بر "به‌روزرسانی از سرور"، لیست پرسنل، سرانه، روزهای تولید و ماه تخصیص از سرور خوانده می‌شود. (سرانه از %s، روزهای تولید از %s، ماه از %s).
2. تخصیص ساعات: فقط ستون "ساعت اضافه کاری" قابل ویرایش است. مجموع باید با سرانه برابر بماند.
3. ماه تخصیص: ماه تخصیص یافته (از سرور) در ستون "ماه تخصیص" نمایش داده می‌شود و قابل ویرایش نیست.
4. قفل کردن ساعت: با تیک ستون "قفل"، ساعت پرسنل ثابت می‌ماند.
5. بررسی نهایی: برچسب "مجموع ساعات" باید نشان دهد که مجموع با سرانه برابر است.
6. خروجی اکسل: پس از تخصیص صحیح، خروجی بگیرید (ماه فایل بر اساس ماه سرور خواهد بود).
7. پاک کردن جدول: حذف اطلاعات جدول فعلی برای بارگذاری مجدد.
توجه: سرانه، روز تولید، ماه و لیست پرسنل قابل ویرایش نیستند.`, seranehInfo, prodDaysInfo, monthInfo)
	}
	displayHelpText := strings.ReplaceAll(helpText, "<b>", "")
	displayHelpText = strings.ReplaceAll(displayHelpText, "</b>", "")
	displayHelpText = strings.ReplaceAll(displayHelpText, "<ol>", "")
	displayHelpText = strings.ReplaceAll(displayHelpText, "</ol>", "")
	displayHelpText = strings.ReplaceAll(displayHelpText, "<li>", "\n- ")
	displayHelpText = strings.ReplaceAll(displayHelpText, "<i>", "")
	displayHelpText = strings.ReplaceAll(displayHelpText, "</i>", "")

	helpDialogContent := widget.NewLabel(displayHelpText)
	helpDialogContent.Wrapping = fyne.TextWrapWord

	dialog.ShowCustom("راهنما", "بستن", container.NewScroll(helpDialogContent), ui.Window)
}
func (ui *MainUI) onShowAbout() {
	iconResource := fyne.NewStaticResource("enterprise.ico", resources.AppIconData)
	title := widget.NewLabelWithStyle("سیستم تخصیص ساعت اضافه کاری نورد میلگرد کاسپین", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	version := widget.NewLabel("نسخه 1.0.0 (GoLang - Fyne Edition)")
	developer := widget.NewLabel("توسعه دهنده: هاتف پوررجبی")
	emailLink := widget.NewHyperlink("hatef.pr@outlook.com", parseURL("mailto:hatef.pr@outlook.com"))

	content := container.NewVBox(
		container.NewCenter(widget.NewIcon(iconResource)),
		title,
		version,
		developer,
		container.NewCenter(emailLink),
	)
	dialog.ShowCustom("درباره برنامه", "بستن", content, ui.Window)
}
func (ui *MainUI) onLogout() {
	dialog.ShowConfirm("خروج از حساب", "آیا مطمئن هستید که می‌خواهید از حساب کاربری خود خارج شوید؟", func(confirm bool) {
		if confirm {
			if ui.logoutHandler != nil {
				ui.logoutHandler()
			} else {
				fyne.LogError("Logout handler is nil in MainUI", nil)
				ui.Window.Close()
			}
		}
	}, ui.Window)
}
func (ui *MainUI) onAdminCreateTable() {
	if ui.currentDepartmentData == nil || ui.currentDepartmentData.DepartmentShiftName == "" {
		dialog.ShowError(fmt.Errorf("ابتدا یک واحد سازمانی را برای ایجاد جدول انتخاب کنید"), ui.Window)
		return
	}
	if err := ui.adminManualEmployeesInput.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("تعداد پرسنل نامعتبر است: %w", err), ui.Window)
		return
	}
	numEmployeesStr := ui.adminManualEmployeesInput.Text
	numEmployees, _ := strconv.Atoi(numEmployeesStr)
	totalHoursStr := ui.totalHoursInput.Text
	totalHours, err := strconv.Atoi(totalHoursStr)
	if err != nil || totalHours < 0 {
		dialog.ShowError(fmt.Errorf("مقدار سرانه کل (%s) برای جدول دستی نامعتبر است. لطفاً یک عدد صحیح غیرمنفی وارد کنید.", totalHoursStr), ui.Window)
		return
	}
	deptName := ui.currentDepartmentData.DepartmentShiftName
	dialog.ShowConfirm("ایجاد جدول دستی", fmt.Sprintf("ایجاد جدول جدید برای واحد '%s' با %d پرسنل و سرانه %d ساعت؟\nاطلاعات قبلی این واحد پاک خواهد شد.", deptName, numEmployees, totalHours), func(confirm bool) {
		if !confirm {
			return
		}
		dataToUpdate, ok := core.AllDepartmentsData[deptName]
		if !ok {
			dialog.ShowError(fmt.Errorf("خطای داخلی: داده‌های واحد '%s' یافت نشد.", deptName), ui.Window)
			return
		}
		dataToUpdate.TotalHours = totalHours
		dataToUpdate.ProductionDays = 0
		dataToUpdate.MonthName = core.GetCurrentPersianMonthName()
		dataToUpdate.Employees = make([]core.Employee, numEmployees)
		for i := 0; i < numEmployees; i++ {
			dataToUpdate.Employees[i] = core.Employee{
				Name:      fmt.Sprintf("پرسنل جدید %d", i+1),
				ID:        fmt.Sprintf("%04d", 1000+i+(time.Now().Second()%100)),
				Hours:     0,
				Locked:    false,
				MonthType: dataToUpdate.MonthName,
			}
		}

		ui.totalHoursInput.SetText(strconv.Itoa(dataToUpdate.TotalHours))
		ui.productionDaysInput.SetText("0")

		var empInterfaces []interface{}
		for i := range dataToUpdate.Employees {
			empInterfaces = append(empInterfaces, &dataToUpdate.Employees[i])
		}
		ui.currentEmployees.Set(empInterfaces)
		ui.reallocateHours()
		ui.exportButton.Enable()
		dialog.ShowInformation("موفقیت", fmt.Sprintf("جدول دستی برای واحد '%s' با موفقیت ایجاد و سرانه توزیع شد.", deptName), ui.Window)
	}, ui.Window)
}
func (ui *MainUI) onAdminImportExcel() {
	fileOpenDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, errDialog error) {
		if errDialog != nil {
			dialog.ShowError(errDialog, ui.Window)
			return
		}
		if reader == nil {
			return
		}
		filePath := reader.URI().Path()
		errCloseReader := reader.Close()
		if errCloseReader != nil {
			fyne.LogError("Failed to close file reader in onAdminImportExcel", errCloseReader)
		}

		progress := dialog.NewProgressInfinite("در حال پردازش فایل اکسل", "لطفاً منتظر بمانید...", ui.Window)
		progress.Show()
		go func() {
			defer progress.Hide()

			fileTotalHours, fileProdDays, fileMonthF3, errReadBase := excel.ReadBasicDataFromExcel(filePath)
			if errReadBase != nil {
				dialog.ShowInformation("هشدار خواندن فایل", fmt.Sprintf("خطا در خواندن اطلاعات پایه (F1,F2,F3) از فایل اکسل: %v\nبا مقادیر پیش‌فرض برای اولین واحد ادامه داده می‌شود.", errReadBase), ui.Window)
				fileTotalHours = 0
				fileProdDays = 0
				fileMonthF3 = ""
			}
			finalFileMainMonth := fileMonthF3
			if finalFileMainMonth == "" {
				finalFileMainMonth = core.GetCurrentPersianMonthName()
				if fileMonthF3 != "" {
					dialog.ShowInformation("هشدار ماه فایل", fmt.Sprintf("مقدار ماه در سلول %s فایل اکسل ('%s') نامعتبر است. از ماه جاری سیستم (%s) برای اولین واحد استفاده خواهد شد.", core.MonthCell, fileMonthF3, finalFileMainMonth), ui.Window)
				}
			}
			importedDeptShiftsSuccess := []string{}
			skippedDeptsMessages := []string{}

			uniqueDeptsInFile, err := excel.GetUniqueDepartmentsFromFile(filePath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("خطا در خواندن لیست واحدها از فایل اکسل: %w", err), ui.Window)
				return
			}
			if len(uniqueDeptsInFile) == 0 {
				dialog.ShowInformation("بدون داده", "هیچ نام واحدی در ستون اول فایل اکسل یافت نشد.", ui.Window)
				return
			}
			processedFirstDeptInFile := false
			for _, deptShiftToImport := range uniqueDeptsInFile {
				isManageable := false
				for _, manageable := range core.ManageableDepartments {
					if deptShiftToImport == manageable {
						isManageable = true
						break
					}
				}
				if !isManageable {
					skippedDeptsMessages = append(skippedDeptsMessages, fmt.Sprintf("%s (واحد تعریف نشده در برنامه)", deptShiftToImport))
					continue
				}
				employees, err := excel.ReadEmployeesFromExcel(filePath, deptShiftToImport)
				if err != nil {
					skippedDeptsMessages = append(skippedDeptsMessages, fmt.Sprintf("%s (خطا در خواندن پرسنل: %v)", deptShiftToImport, err))
					continue
				}
				if len(employees) == 0 {
					skippedDeptsMessages = append(skippedDeptsMessages, fmt.Sprintf("%s (بدون پرسنل در فایل)", deptShiftToImport))
					continue
				}
				currentDeptTotalHours := 0
				currentDeptProdDays := 0
				currentDeptMonth := core.GetCurrentPersianMonthName()
				if !processedFirstDeptInFile {
					currentDeptTotalHours = fileTotalHours
					currentDeptProdDays = fileProdDays
					currentDeptMonth = finalFileMainMonth
					processedFirstDeptInFile = true
				}

				deptData, exists := core.AllDepartmentsData[deptShiftToImport]
				if !exists {
					deptData = &core.DepartmentData{DepartmentShiftName: deptShiftToImport}
					core.AllDepartmentsData[deptShiftToImport] = deptData
				}
				deptData.TotalHours = currentDeptTotalHours
				deptData.ProductionDays = currentDeptProdDays
				deptData.MonthName = currentDeptMonth
				deptData.Employees = employees
				for i := range deptData.Employees {
					deptData.Employees[i].MonthType = currentDeptMonth
				}
				importedDeptShiftsSuccess = append(importedDeptShiftsSuccess, deptShiftToImport)
			}

			if len(importedDeptShiftsSuccess) > 0 {
				currentSelectedInCombo := ui.deptComboBox.Selected
				needsUIFullRefreshForCurrent := false
				for _, imported := range importedDeptShiftsSuccess {
					if imported == currentSelectedInCombo {
						needsUIFullRefreshForCurrent = true
						break
					}
				}
				if needsUIFullRefreshForCurrent {
					ui.refreshUIForCurrentDepartment()
				}
			}
			if len(skippedDeptsMessages) > 0 {
				dialog.ShowInformation("واحدهای رد شده", strings.Join(skippedDeptsMessages, "\n"), ui.Window)
			}
		}()
	}, ui.Window)
	fileOpenDialog.Show()
}
func (ui *MainUI) onManageCloudLinks() {
	linkManagerDialog := CreateCloudLinkManagerDialog(ui.App, ui.Window, func(changed bool) {
		if changed {
			fmt.Println("لینک‌های ابری از دیالوگ تغییر کردند.")
			dialog.ShowInformation("لینک‌ها به‌روز شد", "تغییرات در لینک‌های ابری ذخیره شد. برای مشاهده اثر تغییرات در داده‌های واحد، ممکن است نیاز به 'به‌روزرسانی از سرور' مجدد باشد.", ui.Window)
		}
	})
	linkManagerDialog.Show()
}
func parseURL(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}
	return u
}
