package store

import (
	"sync"

	"wiki/internal/model"
)

// MemoryStore 基于内存的 Store 实现，使用读写锁保证并发安全。
type MemoryStore struct {
	mu          sync.RWMutex
	authors     map[string]*model.Author
	directories map[string]*model.Directory
	documents   map[string]*model.Document
	versions    map[string]*model.Version
	tags        map[string]*model.Tag
	comments    map[string]*model.Comment
	favorites   map[string]*model.Favorite
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		authors:     make(map[string]*model.Author),
		directories: make(map[string]*model.Directory),
		documents:   make(map[string]*model.Document),
		versions:    make(map[string]*model.Version),
		tags:        make(map[string]*model.Tag),
		comments:    make(map[string]*model.Comment),
		favorites:   make(map[string]*model.Favorite),
	}
}

var _ Store = (*MemoryStore)(nil)
