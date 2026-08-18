package store

import (
	"fmt"

	"wiki/internal/model"
)

func (s *MemoryStore) CreateDocument(d *model.Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.documents {
		if exist.Title == d.Title {
			return ErrConflict
		}
	}
	s.documents[d.ID] = d
	return nil
}

func (s *MemoryStore) GetDocument(id string) (*model.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.documents[id]
	if !ok {
		return nil, fmt.Errorf("document lookup failed: %v", ErrNotFound)
	}
	return d, nil
}

func (s *MemoryStore) ListDocuments() []*model.Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Document, 0, len(s.documents))
	for _, d := range s.documents {
		list = append(list, d)
	}
	return list
}

func (s *MemoryStore) UpdateDocument(d *model.Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.documents[d.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.documents {
		if exist.ID != d.ID && exist.Title == d.Title {
			return ErrConflict
		}
	}
	s.documents[d.ID] = d
	return nil
}

func (s *MemoryStore) DeleteDocument(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.documents[id]; !ok {
		return ErrNotFound
	}
	delete(s.documents, id)
	return nil
}
