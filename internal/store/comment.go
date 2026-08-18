package store

import (
	"wiki/internal/model"
)

func (s *MemoryStore) CreateComment(c *model.Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.comments[c.ID] = c
	return nil
}

func (s *MemoryStore) GetComment(id string) (*model.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.comments[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *MemoryStore) ListComments() []*model.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Comment, 0, len(s.comments))
	for _, c := range s.comments {
		list = append(list, c)
	}
	return list
}

func (s *MemoryStore) DeleteComment(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.comments[id]; !ok {
		return ErrNotFound
	}
	delete(s.comments, id)
	return nil
}
