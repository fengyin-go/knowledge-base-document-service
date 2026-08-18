package store

import (
	"wiki/internal/model"
)

func (s *MemoryStore) CreateAuthor(a *model.Author) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.authors {
		if exist.Email == a.Email {
			return ErrConflict
		}
	}
	s.authors[a.ID] = a
	return nil
}

func (s *MemoryStore) GetAuthor(id string) (*model.Author, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.authors[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *MemoryStore) ListAuthors() []*model.Author {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Author, 0, len(s.authors))
	for _, a := range s.authors {
		list = append(list, a)
	}
	return list
}

func (s *MemoryStore) UpdateAuthor(a *model.Author) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.authors[a.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.authors {
		if exist.ID != a.ID && exist.Email == a.Email {
			return ErrConflict
		}
	}
	s.authors[a.ID] = a
	return nil
}

func (s *MemoryStore) DeleteAuthor(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.authors[id]; !ok {
		return ErrNotFound
	}
	delete(s.authors, id)
	return nil
}
