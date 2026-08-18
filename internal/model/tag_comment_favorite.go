package model

import (
	"strings"
	"time"
)

// Tag 标签。
type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	DocCount  int       `json:"doc_count"`
	CreatedAt time.Time `json:"created_at"`
}

func (t *Tag) Validate() error {
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" {
		return NewValidationError("name", "标签名称不能为空")
	}
	return nil
}

// Comment 文档评论。
type Comment struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	AuthorID   string    `json:"author_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

func (c *Comment) Validate() error {
	c.Content = strings.TrimSpace(c.Content)
	if c.DocumentID == "" {
		return NewValidationError("document_id", "文档不能为空")
	}
	if c.AuthorID == "" {
		return NewValidationError("author_id", "评论人不能为空")
	}
	if c.Content == "" {
		return NewValidationError("content", "评论内容不能为空")
	}
	return nil
}

type CommentFilter struct {
	DocumentID string
	AuthorID   string
}

func (f CommentFilter) Match(c *Comment) bool {
	if f.DocumentID != "" && c.DocumentID != f.DocumentID {
		return false
	}
	if f.AuthorID != "" && c.AuthorID != f.AuthorID {
		return false
	}
	return true
}

// Favorite 文档收藏。
type Favorite struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	UserID     string    `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (f *Favorite) Validate() error {
	f.UserID = strings.TrimSpace(f.UserID)
	if f.DocumentID == "" {
		return NewValidationError("document_id", "文档不能为空")
	}
	if f.UserID == "" {
		return NewValidationError("user_id", "用户不能为空")
	}
	return nil
}

type FavoriteFilter struct {
	UserID     string
	DocumentID string
}

func (f FavoriteFilter) Match(fv *Favorite) bool {
	if f.UserID != "" && fv.UserID != f.UserID {
		return false
	}
	if f.DocumentID != "" && fv.DocumentID != f.DocumentID {
		return false
	}
	return true
}
