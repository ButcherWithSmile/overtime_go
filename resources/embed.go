package resources

import _ "embed"

// فایل Fara-Light.otf باید در همین پوشه (resources) موجود باشد
//go:embed Fara-Light.otf
var FaraFontData []byte

// فایل enterprise.ico باید در همین پوشه (resources) موجود باشد
//go:embed enterprise.ico
var AppIconData []byte

// فایل default_cloud_links.json باید در همین پوشه (resources) موجود باشد
//go:embed default_cloud_links.json
var DefaultCloudLinksJSON []byte
