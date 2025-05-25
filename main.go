package main

import (
	"fmt"
	// "path/filepath" // دیگر نیازی به این در main نیست چون GetExecutableDir منتقل شد

	"overtime_go/auth"
	"overtime_go/config"
	"overtime_go/core"
	"overtime_go/gui"
	"overtime_go/resources"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	// "overtime_go/utils" // GetExecutableDir مستقیما در main استفاده نمی‌شود، پس نیازی به وارد کردن utils در main نیست مگر جای دیگر لازم باشد
)

const (
	AppID = "com.hatefpr.overtimeapp.go"
)

var (
	currentUser *core.User
	mainWindow  fyne.Window
	loginWindow fyne.Window
	fyneApp     fyne.App
)

// GetExecutableDir از اینجا حذف شد و به پکیج utils منتقل گردید.

func main() {
	fyneApp = app.NewWithID(AppID)

	customTheme, err := gui.NewCustomTheme(resources.FaraFontData)
	if err != nil {
		fyne.LogError("Failed to create custom theme", err)
	} else {
		fyneApp.Settings().SetTheme(customTheme)
	}
	if resources.AppIconData != nil {
		fyneApp.SetIcon(fyne.NewStaticResource("app_icon.png", resources.AppIconData))
	}

	auth.InitializeDefaultUsers()
	// InitializeDefaultCloudLinks نیاز به GetExecutableDir ندارد چون فایل JSON از embed خوانده می‌شود
	// و فایل cloud_links.json قابل ویرایش توسط کاربر، مسیرش توسط cloud.LoadCloudLinks مدیریت می‌شود.
	core.InitializeDefaultCloudLinks(resources.DefaultCloudLinksJSON)

	showLoginScreen()
	fyneApp.Run()
}

func showLoginScreen() {
	fmt.Println("Showing login screen...")
	appSettings := config.LoadSettings(fyneApp)

	if loginWindow == nil {
		loginWindow = fyneApp.NewWindow("ورود به سیستم")
		loginWindow.SetFixedSize(true)
		loginWindow.CenterOnScreen()
		loginWindow.SetOnClosed(func() {
			fmt.Println("Login window closed by user. Exiting application.")
			fyneApp.Quit()
		})
	}

	loginUI := gui.CreateLoginUI(loginWindow, fyneApp, appSettings, func(user core.User) {
		fmt.Printf("User '%s' logged in successfully.\n", user.Username)
		currentUser = &user
		if loginWindow != nil {
			loginWindow.Hide()
		}
		showMainScreen()
	})

	loginWindow.SetContent(loginUI)
	loginWindow.Resize(fyne.NewSize(450, 350))
	loginWindow.Show()
}

func showMainScreen() {
	if currentUser == nil {
		fyne.LogError("Attempted to show main screen without a logged-in user.", nil)
		if mainWindow != nil {
			mainWindow.Hide()
		}
		showLoginScreen()
		return
	}
	fmt.Println("Showing main screen...")

	if mainWindow == nil {
		mainWindow = fyneApp.NewWindow("سیستم تخصیص ساعت اضافه کاری نورد میلگرد کاسپین")
		mainWindow.CenterOnScreen()
		mainWindow.SetOnClosed(func() {
			fmt.Println("Main window closed by user via 'X' button. Performing logout.")
			performLogout()
		})
	}

	mainUIContent := gui.CreateMainUI(mainWindow, fyneApp, currentUser, performLogout)
	mainWindow.SetContent(mainUIContent)
	mainWindow.Resize(fyne.NewSize(1200, 750))
	mainWindow.Show()
}

func performLogout() {
	fmt.Println("Performing logout...")
	currentUser = nil
	if mainWindow != nil {
		mainWindow.Hide()
	}
	showLoginScreen()
}
