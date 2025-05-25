package config

import (
	"fyne.io/fyne/v2"
)

const (
	prefUsername   = "username"
	prefPassword   = "password"
	prefRememberMe = "remember_me"
)

type AppSettings struct {
	Username   string
	Password   string
	RememberMe bool
}

func LoadSettings(app fyne.App) AppSettings {
	return AppSettings{
		Username:   app.Preferences().StringWithFallback(prefUsername, ""),
		Password:   app.Preferences().StringWithFallback(prefPassword, ""),
		RememberMe: app.Preferences().BoolWithFallback(prefRememberMe, false),
	}
}

func SaveSettings(app fyne.App, settings AppSettings) {
	app.Preferences().SetString(prefUsername, settings.Username)
	app.Preferences().SetString(prefPassword, settings.Password)
	app.Preferences().SetBool(prefRememberMe, settings.RememberMe)
}
