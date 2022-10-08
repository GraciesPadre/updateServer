package main

type UpdateMetaData struct {
	Url        string `json:"url"`
	Version    string `json:"version"`
	DeviceType string `json:"deviceType"`
}

func NewUpdateMetaData(url string, version string) UpdateMetaData {
	return UpdateMetaData{
		Url:        url,
		Version:    version,
		DeviceType: "Spitz",
	}
}
