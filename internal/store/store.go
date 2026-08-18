// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"wiki/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法。
type Store interface {
	// 作者
	CreateAuthor(a *model.Author) error
	GetAuthor(id string) (*model.Author, error)
	ListAuthors() []*model.Author
	UpdateAuthor(a *model.Author) error
	DeleteAuthor(id string) error

	// 目录
	CreateDirectory(d *model.Directory) error
	GetDirectory(id string) (*model.Directory, error)
	ListDirectories() []*model.Directory
	UpdateDirectory(d *model.Directory) error
	DeleteDirectory(id string) error

	// 文档
	CreateDocument(d *model.Document) error
	GetDocument(id string) (*model.Document, error)
	ListDocuments() []*model.Document
	UpdateDocument(d *model.Document) error
	DeleteDocument(id string) error

	// 版本
	CreateVersion(v *model.Version) error
	GetVersion(id string) (*model.Version, error)
	ListVersions() []*model.Version
	DeleteVersion(id string) error

	// 标签
	CreateTag(t *model.Tag) error
	GetTag(id string) (*model.Tag, error)
	GetTagByName(name string) (*model.Tag, error)
	ListTags() []*model.Tag
	UpdateTag(t *model.Tag) error
	DeleteTag(id string) error

	// 评论
	CreateComment(c *model.Comment) error
	GetComment(id string) (*model.Comment, error)
	ListComments() []*model.Comment
	DeleteComment(id string) error

	// 收藏
	CreateFavorite(f *model.Favorite) error
	GetFavorite(id string) (*model.Favorite, error)
	ListFavorites() []*model.Favorite
	DeleteFavorite(id string) error
}
