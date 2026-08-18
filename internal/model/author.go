package model

import (
	"strings"
	"time"
)

// Author 文档作者。
type Author struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *Author) Validate() error {
	a.Name = strings.TrimSpace(a.Name)
	a.Email = strings.TrimSpace(a.Email)
	if a.Name == "" {
		return NewValidationError("name", "作者姓名不能为空")
	}
	if a.Email == "" {
		return NewValidationError("email", "作者邮箱不能为空")
	}
	a.Bio = strings.TrimSpace(a.Bio)
	return nil
}

type AuthorFilter struct {
	Keyword string
}

func (f AuthorFilter) Match(a *Author) bool {
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(a.Name), k) &&
			!strings.Contains(strings.ToLower(a.Email), k) {
			return false
		}
	}
	return true
}
