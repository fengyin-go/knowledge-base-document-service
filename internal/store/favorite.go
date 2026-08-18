package store

import (
	"wiki/internal/model"
)

func (s *MemoryStore) CreateFavorite(f *model.Favorite) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.favorites {
		if exist.DocumentID == f.DocumentID && exist.UserID == f.UserID {
			return ErrConflict
		}
	}
	s.favorites[f.ID] = f
	return nil
}

func (s *MemoryStore) GetFavorite(id string) (*model.Favorite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.favorites[id]
	if !ok {
		return nil, ErrNotFound
	}
	return f, nil
}

func (s *MemoryStore) ListFavorites() []*model.Favorite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Favorite, 0, len(s.favorites))
	for _, f := range s.favorites {
		list = append(list, f)
	}
	return list
}

func (s *MemoryStore) DeleteFavorite(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.favorites[id]; !ok {
		return ErrNotFound
	}
	delete(s.favorites, id)
	return nil
}
