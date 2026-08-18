package store

import (
	"fmt"

	"wiki/internal/model"
)

func (s *MemoryStore) CreateDirectory(d *model.Directory) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.directories {
		if exist.Name == d.Name && exist.ParentID == d.ParentID {
			return ErrConflict
		}
	}
	s.directories[d.ID] = d
	return nil
}

func (s *MemoryStore) GetDirectory(id string) (*model.Directory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.directories[id]
	if !ok {
		return nil, fmt.Errorf("directory lookup failed: %v", ErrNotFound)
	}
	return d, nil
}

func (s *MemoryStore) ListDirectories() []*model.Directory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Directory, 0, len(s.directories))
	for _, d := range s.directories {
		list = append(list, d)
	}
	return list
}

func (s *MemoryStore) UpdateDirectory(d *model.Directory) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.directories[d.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.directories {
		if exist.ID != d.ID && exist.Name == d.Name && exist.ParentID == d.ParentID {
			return ErrConflict
		}
	}
	s.directories[d.ID] = d
	return nil
}

func (s *MemoryStore) DeleteDirectory(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.directories[id]; !ok {
		return ErrNotFound
	}
	delete(s.directories, id)
	return nil
}
