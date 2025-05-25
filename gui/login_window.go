package gui

import (
	"fmt"
	"overtime_go/auth" // اطمینان از صحت نام ماژول
	"overtime_go/config"
	"overtime_go/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func CreateLoginUI(win fyne.Window, app fyne.App, settings config.AppSettings, onLoginSuccess func(user core.User)) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle("سیستم تخصیص ساعت اضافه کاری", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("نام کاربری واحد")
	if settings.RememberMe && settings.Username != "" {
		usernameEntry.SetText(settings.Username)
	}

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("رمز عبور")
	if settings.RememberMe && settings.Password != "" {
		passwordEntry.SetText(settings.Password)
	}

	rememberCheck := widget.NewCheck("مرا به خاطر بسپار", nil)
	rememberCheck.SetChecked(settings.RememberMe)

	loginButton := widget.NewButton("ورود", func() {
		username := usernameEntry.Text
		password := passwordEntry.Text

		if username == "" || password == "" {
			dialog.ShowInformation("خطا", "نام کاربری و رمز عبور را وارد کنید.", win)
			return
		}

		user, authenticated := auth.AuthenticateUser(username, password)
		if authenticated {
			if rememberCheck.Checked {
				settings.Username = username
				settings.Password = password
				settings.RememberMe = true
			} else {
				settings.Username = ""
				settings.Password = ""
				settings.RememberMe = false
			}
			config.SaveSettings(app, settings)

			// dialog.ShowInformation("موفقیت", "ورود موفقیت آمیز بود!", win) // این دیالوگ را حذف می‌کنیم تا بلافاصله به پنجره اصلی برود
			// win.Hide() // این کار در main.go انجام می‌شود
			onLoginSuccess(*user)
		} else {
			dialog.ShowError(fmt.Errorf("نام کاربری یا رمز عبور اشتباه است."), win)
			passwordEntry.SetText("")
		}
	})

	passwordEntry.OnSubmitted = func(s string) {
		loginButton.OnTapped()
	}
	usernameEntry.OnSubmitted = func(s string) {
		win.Canvas().Focus(passwordEntry)
	}

	form := container.NewVBox(
		widget.NewLabel("نام کاربری:"),
		usernameEntry,
		widget.NewLabel("رمز عبور:"),
		passwordEntry,
		rememberCheck,
		loginButton,
	)

	helpText := widget.NewLabelWithStyle("برای دریافت نام کاربری و رمز عبور با بخش سرمایه‌های انسانی تماس بگیرید.", fyne.TextAlignCenter, fyne.TextStyle{})

	return container.NewBorder(
		titleLabel,
		helpText,
		nil,
		nil,
		container.NewCenter(form),
	)
}
