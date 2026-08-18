package model

import (
	"strings"
	"time"
)

const (
	DocumentDraft     = "draft"
	DocumentReviewing = "reviewing"
	DocumentPublished = "published"
	DocumentArchived  = "archived"
)

// documentTransitions 文档状态机。
var documentTransitions = map[string]map[string]bool{
	DocumentDraft:     {DocumentReviewing: true, DocumentArchived: true},
	DocumentReviewing: {DocumentPublished: true, DocumentArchived: true},
	DocumentPublished: {DocumentArchived: true},
	DocumentArchived:  {DocumentPublished: true},
}

// CanTransitionDocument 判断文档状态流转是否合法。
func CanTransitionDocument(from, to string) bool {
	if m, ok := documentTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Document 文档。
type Document struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	DirectoryID string     `json:"directory_id"`
	AuthorID    string     `json:"author_id"`
	Tags        []string   `json:"tags"`
	Status      string     `json:"status"`
	ViewCount   int        `json:"view_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

func (d *Document) Validate() error {
	d.Title = strings.TrimSpace(d.Title)
	if d.Title == "" {
		return NewValidationError("title", "文档标题不能为空")
	}
	if d.DirectoryID == "" {
		return NewValidationError("directory_id", "所属目录不能为空")
	}
	if d.AuthorID == "" {
		return NewValidationError("author_id", "作者不能为空")
	}
	if d.Status == "" {
		d.Status = DocumentDraft
	}
	if d.Status != DocumentDraft && d.Status != DocumentReviewing && d.Status != DocumentPublished && d.Status != DocumentArchived {
		return NewValidationError("status", "文档状态不合法")
	}
	return nil
}

type DocumentFilter struct {
	DirectoryID string
	AuthorID    string
	Status      string
	Tag         string
	Keyword     string
}

func (f DocumentFilter) Match(d *Document) bool {
	if f.DirectoryID != "" && d.DirectoryID != f.DirectoryID {
		return false
	}
	if f.AuthorID != "" && d.AuthorID != f.AuthorID {
		return false
	}
	if f.Status != "" && d.Status != f.Status {
		return false
	}
	if f.Tag != "" {
		found := false
		for _, t := range d.Tags {
			if t == f.Tag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(d.Title), k) &&
			!strings.Contains(strings.ToLower(d.Content), k) {
			return false
		}
	}
	return true
}

// HasTag 判断文档是否包含指定标签。
func (d *Document) HasTag(tag string) bool {
	for _, t := range d.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
