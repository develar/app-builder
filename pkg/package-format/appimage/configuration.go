package appimage

type AppImageConfiguration struct {
	ProductName       string `json:"productName"`
	ProductFilename   string `json:"productFilename"`
	ExecutableName    string `json:"executableName"`
	SystemIntegration string `json:"systemIntegration"`

	DesktopEntry string `json:"desktopEntry"`

	Icons            []IconInfo        `json:"icons"`
	FileAssociations []FileAssociation `json:"fileAssociations"`
}

type IconInfo struct {
	File string `json:"file"`
	Size int    `json:"size"`
}

type FileAssociation struct {
	Ext      interface{} `json:"ext"` // Can be string or []string
	MimeType string      `json:"mimeType"`
}
