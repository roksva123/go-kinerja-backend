package model

import "fmt"

type TaskStatus struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Color string `json:"color"`
}

type UserSimple struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Color    string `json:"color"`
	RoleKey  string `json:"role_key"`
}

type TaskResponse struct {
	ID          string  `json:"id"`
	TaskID        string  `json:"task_id"`
	Name        string  `json:"name"`
	TextContent string  `json:"text_content"`
	Description string  `json:"description"`
	Status      struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Color string `json:"color"`
	} `json:"status"`
	DateDone       *int64 `json:"date_done,omitempty"`
	DateClosed     *int64 `json:"date_closed,omitempty"`

	Username string `json:"username"`
	Email    string `json:"email"`
	Color    string `json:"color"`

	TimeEstimateMs *int64 `json:"time_estimate_ms,omitempty"`
	TimeSpentMs    *int64 `json:"time_spent_ms,omitempty"`

	StartDate    *int64 `json:"start_date"`
	DueDate      *int64 `json:"due_date"`
	DateCreated *int64 `json:"date_created"`
	TimeEstimate *int64 `json:"time_estimate"`
}

type FolderStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Type   string `json:"type"`
	Order  int    `json:"orderindex"`
	Color  string `json:"color"`
}

type SpaceInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Folder struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	TaskCount int            `json:"task_count"`
	Archived  bool           `json:"archived"`
	Space     SpaceInfo      `json:"space"`
	Statuses  []FolderStatus `json:"statuses"`
}

type FolderResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Orderindex int    `json:"orderindex"`
	Hidden     bool   `json:"hidden"`
}

func ParseClickUpTask(raw map[string]interface{}) *TaskResponse {
	t := &TaskResponse{}

	if v, ok := raw["id"].(string); ok {
		t.ID = v
	}
	if v, ok := raw["name"].(string); ok {
		t.Name = v
	}
	if v, ok := raw["text_content"].(string); ok {
		t.TextContent = v
	}
	if v, ok := raw["description"].(string); ok {
		t.Description = v
	}

	if st, ok := raw["status"].(map[string]interface{}); ok {
		if v, ok := st["id"].(string); ok {
			t.Status.ID = v
		}
		if v, ok := st["status"].(string); ok {
			t.Status.Name = v
		}
		if v, ok := st["type"].(string); ok {
			t.Status.Type = v
		}
		if v, ok := st["color"].(string); ok {
			t.Status.Color = v
		}
	}

	if arr, ok := raw["assignees"].([]interface{}); ok && len(arr) > 0 {
		if user, ok := arr[0].(map[string]interface{}); ok {
			if v, ok := user["username"].(string); ok {
				t.Username = v
			}
			if v, ok := user["email"].(string); ok {
				t.Email = v
			}
			if v, ok := user["color"].(string); ok {
				t.Color = v
			}
		}
	}

	if cr, ok := raw["creator"].(map[string]interface{}); ok {
		if t.Username == "" {
			if v, ok := cr["username"].(string); ok {
				t.Username = v
			}
		}
		if t.Email == "" {
			if v, ok := cr["email"].(string); ok {
				t.Email = v
			}
		}
	}

	if v, ok := raw["date_done"].(string); ok && v != "" {
		parsed := parseInt64(v)
		t.DateDone = &parsed
	} else if v, ok := raw["date_done"].(float64); ok {
		x := int64(v)
		t.DateDone = &x
	}

	if v, ok := raw["date_closed"].(string); ok && v != "" {
		parsed := parseInt64(v)
		t.DateClosed = &parsed
	} else if v, ok := raw["date_closed"].(float64); ok {
		x := int64(v)
		t.DateClosed = &x
	}

	return t
}

func ParseClickUpUser(raw map[string]interface{}) *UserResponse {
	u := &UserResponse{}

	if user, ok := raw["user"].(map[string]interface{}); ok {
		if v, ok := user["id"].(float64); ok {
			u.ID = intToString(v)
		}
		if v, ok := user["username"].(string); ok {
			u.Username = v
		}
		if v, ok := user["email"].(string); ok {
			u.Email = v
		}
		if v, ok := user["color"].(string); ok {
			u.Color = v
		}
	}

	return u
}

func ParseClickUpFolder(raw map[string]interface{}) *FolderResponse {
	f := &FolderResponse{}

	if v, ok := raw["id"].(string); ok {
		f.ID = v
	}
	if v, ok := raw["name"].(string); ok {
		f.Name = v
	}
	if v, ok := raw["orderindex"].(float64); ok {
		f.Orderindex = int(v)
	}
	if v, ok := raw["hidden"].(bool); ok {
		f.Hidden = v
	}

	return f
}

func parseInt64(s string) int64 {
	var out int64
	fmt.Sscan(s, &out)
	return out
}

func intToString(v float64) string {
	return fmt.Sprintf("%.0f", v)
}
