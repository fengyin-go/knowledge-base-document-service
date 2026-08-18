package model

import (
	"strings"
	"time"
)

// Directory 文档目录（支持层级，ParentID 为空表示根目录）。
type Directory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ParentID    string    `json:"parent_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (d *Directory) Validate() error {
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		return NewValidationError("name", "目录名称不能为空")
	}
	if d.ParentID != "" && d.ParentID == d.ID {
		return NewValidationError("parent_id", "父目录不能是自身")
	}
	d.Description = strings.TrimSpace(d.Description)
	return nil
}

type DirectoryFilter struct {
	ParentID string
	Keyword  string
}

func (f DirectoryFilter) Match(d *Directory) bool {
	if f.ParentID != "" && d.ParentID != f.ParentID {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(d.Name), k) {
			return false
		}
	}
	return true
}
