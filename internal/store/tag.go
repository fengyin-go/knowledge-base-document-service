package store

import (
	"wiki/internal/model"
)

func (s *MemoryStore) CreateTag(t *model.Tag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.tags {
		if exist.Name == t.Name {
			return ErrConflict
		}
	}
	s.tags[t.ID] = t
	return nil
}

func (s *MemoryStore) GetTag(id string) (*model.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tags[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (s *MemoryStore) GetTagByName(name string) (*model.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tags {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListTags() []*model.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Tag, 0, len(s.tags))
	for _, t := range s.tags {
		list = append(list, t)
	}
	return list
}

func (s *MemoryStore) UpdateTag(t *model.Tag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tags[t.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.tags {
		if exist.ID != t.ID && exist.Name == t.Name {
			return ErrConflict
		}
	}
	s.tags[t.ID] = t
	return nil
}

func (s *MemoryStore) DeleteTag(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tags[id]; !ok {
		return ErrNotFound
	}
	delete(s.tags, id)
	return nil
}
