package store

import (
	"wiki/internal/model"
)

func (s *MemoryStore) CreateVersion(v *model.Version) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.versions[v.ID] = v
	return nil
}

func (s *MemoryStore) GetVersion(id string) (*model.Version, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.versions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *MemoryStore) ListVersions() []*model.Version {
	list := make([]*model.Version, 0, len(s.versions))
	for _, v := range s.versions {
		list = append(list, v)
	}
	return list
}

func (s *MemoryStore) DeleteVersion(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.versions[id]; !ok {
		return ErrNotFound
	}
	delete(s.versions, id)
	return nil
}
