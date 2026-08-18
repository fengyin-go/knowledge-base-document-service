package model

import (
	"strings"
	"time"
)

// Version 文档历史版本。
type Version struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	VersionNo  int       `json:"version_no"`
	Content    string    `json:"content"`
	Changelog  string    `json:"changelog"`
	CreatedAt  time.Time `json:"created_at"`
}

func (v *Version) Validate() error {
	v.Changelog = strings.TrimSpace(v.Changelog)
	if v.DocumentID == "" {
		return NewValidationError("document_id", "文档不能为空")
	}
	if v.VersionNo <= 0 {
		return NewValidationError("version_no", "版本号必须大于 0")
	}
	return nil
}
