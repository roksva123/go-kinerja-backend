package model

type SpaceInfo struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}


type Folder struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	OrderIndex int       `json:"orderindex"`
	Archived   bool      `json:"archived"`
	TaskCount  string    `json:"task_count"` 
	Lists      []List    `json:"lists"`
	Space      SpaceInfo `json:"space"`
}

type List struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Archived bool   `json:"archived"`
	FolderID string `json:"-"` 
	SpaceID  string `json:"-"` 
}

type ClickUpFoldersResponse struct {
	Folders []Folder `json:"folders"`
}