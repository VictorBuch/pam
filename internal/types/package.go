package types

type Package struct {
	PName       string `json:"pname"`
	Version     string `json:"version"`
	Description string `json:"description"`
	FullPath    string
}
