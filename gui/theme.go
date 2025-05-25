package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type customTheme struct {
	regular fyne.Resource
}

var _ fyne.Theme = (*customTheme)(nil)

func NewCustomTheme(fontData []byte) (fyne.Theme, error) {
	fontResource := fyne.NewStaticResource("Fara-Light.otf", fontData)
	return &customTheme{regular: fontResource}, nil
}

func (m *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantDark {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0x28, G: 0x28, B: 0x28, A: 0xff} // پس‌زمینه تیره‌تر
		case theme.ColorNameForeground:
			return color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff} // متن روشن‌تر
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0x40, G: 0xa0, B: 0xff, A: 0xff} // رنگ اصلی آبی روشن
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 0x32, G: 0x32, B: 0x32, A: 0xff} // پس‌زمینه فیلدهای ورودی
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return m.regular
}

func (m *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *customTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14 // اندازه متن بزرگتر
	case theme.SizeNameHeadingText:
		return 16 // اندازه عنوان‌ها بزرگتر
	case theme.SizeNameSubHeadingText:
		return 15
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInlineIcon:
		return 24 // آیکون‌های بزرگتر
	case theme.SizeNamePadding:
		return 6 // فاصله‌گذاری بیشتر
	case theme.SizeNameScrollBar:
		return 12 // اسکرول‌بار ضخیم‌تر
	case theme.SizeNameSeparatorThickness:
		return 2
	}
	return theme.DefaultTheme().Size(name)
}
