package gui

import (
	"fmt"
	"os"

	// "path/filepath" // اگر نیاز به کار با مسیرها باشد
	"sort"
	"strings"

	"overtime_go/cloud" // اطمینان از صحت نام ماژول
	"overtime_go/core"
	"overtime_go/excel"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	// "overtime_go/resources" // اگر مستقیما استفاده شود
)

type cloudLinkManagerDialog struct {
	dialog       dialog.Dialog
	app          fyne.App
	parentWindow fyne.Window
	linksTable   *widget.Table

	editableLinks map[string]string

	onCloseCallback func(changed bool)

	sortedDisplayDepts []string
}

func CreateCloudLinkManagerDialog(app fyne.App, parent fyne.Window, onCloseCallback func(changed bool)) dialog.Dialog {
	loadedLinks := cloud.LoadCloudLinks()
	editableLinksMap := make(map[string]string)
	for k, v := range loadedLinks {
		editableLinksMap[k] = v
	}

	displayDepts := make([]string, 0, len(core.ManageableDepartments))
	for _, dept := range core.ManageableDepartments {
		displayDepts = append(displayDepts, dept)
	}
	sort.Strings(displayDepts)

	manager := &cloudLinkManagerDialog{
		app:                app,
		parentWindow:       parent,
		editableLinks:      editableLinksMap,
		onCloseCallback:    onCloseCallback,
		sortedDisplayDepts: displayDepts,
	}

	manager.linksTable = widget.NewTable(
		manager.tableLength,
		manager.createCell,
		manager.updateCell,
	)
	manager.linksTable.SetColumnWidth(0, 280)
	manager.linksTable.SetColumnWidth(1, 380)

	testLinkButton := widget.NewButtonWithIcon("تست لینک انتخاب شده", theme.SearchIcon(), manager.onTestSelectedLink)

	buttonsTop := container.NewHBox(testLinkButton)

	helpTextContent := fmt.Sprintf(`- لینک دانلود مستقیم فایل اکسل (e.g., Dropbox dl=1) را برای هر واحد ویرایش کنید.
- لینک انتخاب شده را تست کنید (تست، سلول‌های %s, %s, %s را در فایل اکسل بررسی می‌کند).
- تغییرات در فایل cloud_links.json (کنار فایل اجرایی برنامه) ذخیره می‌شوند.`,
		core.SeranehCell, core.ProductionDaysCell, core.MonthCell)

	helpLabel := widget.NewLabel(helpTextContent)
	helpLabel.Wrapping = fyne.TextWrapWord

	mainDialogContent := container.NewBorder(
		buttonsTop,
		helpLabel,
		nil, nil,
		container.NewScroll(manager.linksTable),
	)

	manager.dialog = dialog.NewCustomConfirm(
		"مدیریت لینک‌های دانلود سرور",
		"ذخیره تغییرات",
		"انصراف",
		mainDialogContent,
		func(saveConfirmed bool) {
			if saveConfirmed {
				originalLinks := cloud.LoadCloudLinks()
				changed := false
				if len(manager.editableLinks) != len(originalLinks) {
					changed = true
				} else {
					for k, v := range manager.editableLinks {
						if ov, ok := originalLinks[k]; !ok || ov != v {
							changed = true
							break
						}
					}
				}

				if !changed {
					if manager.onCloseCallback != nil {
						manager.onCloseCallback(false)
					}
					return
				}

				err := cloud.SaveCloudLinks(manager.editableLinks)
				if err != nil {
					dialog.ShowError(fmt.Errorf("خطا در ذخیره فایل لینک‌ها: %w", err), manager.parentWindow)
					if manager.onCloseCallback != nil {
						manager.onCloseCallback(false)
					}
					return
				}
				dialog.ShowInformation("ذخیره شد", "تغییرات لینک‌ها با موفقیت در فایل cloud_links.json ذخیره شد.", manager.parentWindow)
				if manager.onCloseCallback != nil {
					manager.onCloseCallback(true)
				}
			} else {
				if manager.onCloseCallback != nil {
					manager.onCloseCallback(false)
				}
			}
		},
		parent,
	)
	manager.dialog.Resize(fyne.NewSize(750, 550))
	return manager.dialog
}

func (m *cloudLinkManagerDialog) tableLength() (rows, cols int) {
	return len(m.sortedDisplayDepts), 2
}

func (m *cloudLinkManagerDialog) createCell() fyne.CanvasObject {
	return container.NewMax()
}

func (m *cloudLinkManagerDialog) updateCell(id widget.TableCellID, template fyne.CanvasObject) {
	cellContainer := template.(*fyne.Container)
	// اطمینان از اینکه id.Row در محدوده m.sortedDisplayDepts قرار دارد
	if id.Row < 0 || id.Row >= len(m.sortedDisplayDepts) {
		// fyne.LogError("Row index out of bounds in links dialog table", fmt.Errorf("index %d, len %d", id.Row, len(m.sortedDisplayDepts)))
		cellContainer.Objects = []fyne.CanvasObject{widget.NewLabel("خطا در ردیف")}
		cellContainer.Refresh()
		return
	}
	deptShiftName := m.sortedDisplayDepts[id.Row]

	switch id.Col {
	case 0:
		nameLabel := widget.NewLabel(deptShiftName)
		cellContainer.Objects = []fyne.CanvasObject{nameLabel}
	case 1:
		linkEntry := widget.NewEntry()
		linkEntry.SetText(m.editableLinks[deptShiftName])
		linkEntry.Wrapping = fyne.TextTruncate
		linkEntry.OnChanged = func(newLink string) {
			m.editableLinks[deptShiftName] = newLink
		}
		cellContainer.Objects = []fyne.CanvasObject{linkEntry}
	}
	cellContainer.Refresh()
}

func (m *cloudLinkManagerDialog) onTestSelectedLink() {
	selectedDept := ""

	deptSelector := widget.NewSelect(m.sortedDisplayDepts, func(s string) {
		selectedDept = s
	})
	if len(m.sortedDisplayDepts) > 0 {
		deptSelector.SetSelected(m.sortedDisplayDepts[0])
		selectedDept = m.sortedDisplayDepts[0]
	}

	dialog.ShowForm("انتخاب واحد برای تست لینک", "تست کن", "انصراف", []*widget.FormItem{
		widget.NewFormItem("واحد:", deptSelector),
	}, func(confirm bool) {
		if !confirm || selectedDept == "" {
			return
		}

		linkToTest, ok := m.editableLinks[selectedDept]
		if !ok || linkToTest == "" {
			dialog.ShowError(fmt.Errorf("لینکی برای واحد '%s' جهت تست یافت نشد.", selectedDept), m.parentWindow)
			return
		}

		progress := dialog.NewProgressInfinite("تست لینک: "+selectedDept, "در حال دانلود و بررسی...", m.parentWindow)
		progress.Show()
		go func() {
			defer progress.Hide()
			downloadURL := cloud.ConvertToDownloadLink(linkToTest)

			tempFilePath, err := cloud.DownloadToTempFile(downloadURL, "test_link_mgr_*.xlsx")
			if err != nil {
				// Directly execute the code for UI updates
				dialog.ShowError(fmt.Errorf("خطا در دانلود لینک تست (%s) برای واحد '%s': %w", downloadURL, selectedDept, err), m.parentWindow)
				return
			}
			defer os.Remove(tempFilePath)

			_, _, monthF3, errExcel := excel.ReadBasicDataFromExcel(tempFilePath)
			// fyne.CurrentApp().Driver().RunOnMain(func() {
			if errExcel != nil {
				errMsgDetails := fmt.Sprintf("جزئیات بررسی محتوا: %v.", errExcel)
				if monthF3 == "" && (errExcel != nil && strings.Contains(errExcel.Error(), core.MonthCell)) {
					errMsgDetails += fmt.Sprintf(" مقدار ماه (%s) نیز خوانده نشد یا نامعتبر بود.", core.MonthCell)
				}
				// Replace dialog.ShowWarning with dialog.ShowInformation
				dialog.ShowInformation("نتیجه تست (هشدار)", fmt.Sprintf("لینک برای واحد '%s' قابل دانلود است، اما محتوای فایل اکسل (سلول‌های F1,F2,F3) ممکن است مشکل داشته باشد یا خالی باشد.\n%s", selectedDept, errMsgDetails), m.parentWindow)
			} else {
				finalMonthName := monthF3
				warningMsg := ""
				if finalMonthName == "" {
					warningMsg = fmt.Sprintf("\nهشدار: مقدار ماه در سلول %s فایل اکسل خالی یا نامعتبر است.", core.MonthCell)
				}
				dialog.ShowInformation("نتیجه تست (موفق)", fmt.Sprintf("لینک برای واحد '%s' معتبر به نظر می‌رسد و سلول‌های F1,F2,F3 با موفقیت خوانده شدند (یا خطای بحرانی در خواندن آن‌ها نبود).%s", selectedDept, warningMsg), m.parentWindow)
			}
			// })
		}()
	}, m.parentWindow)
}
