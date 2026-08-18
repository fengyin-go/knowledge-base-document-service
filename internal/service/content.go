package service

import (
	"sort"
	"time"

	"wiki/internal/model"
	"wiki/pkg/idgen"
)

func (s *Service) CreateTag(input model.Tag) (*model.Tag, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	t := &model.Tag{ID: idgen.Hex(), Name: input.Name, CreatedAt: time.Now()}
	if err := s.store.CreateTag(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) GetTag(id string) (*model.Tag, error) {
	return s.store.GetTag(id)
}

func (s *Service) ListTags(page, size int) ([]*model.Tag, int, error) {
	all := s.store.ListTags()
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	total := len(all)
	start := (page - 1) * size
	if start >= total {
		return []*model.Tag{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (s *Service) DeleteTag(id string) error {
	return s.store.DeleteTag(id)
}

// CreateComment 发表评论：校验文档与作者存在。
func (s *Service) CreateComment(input model.Comment) (*model.Comment, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetDocument(input.DocumentID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetAuthor(input.AuthorID); err != nil {
		return nil, err
	}
	c := &model.Comment{
		ID:         idgen.Hex(),
		DocumentID: input.DocumentID,
		AuthorID:   input.AuthorID,
		Content:    input.Content,
		CreatedAt:  time.Now(),
	}
	if err := s.store.CreateComment(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) GetComment(id string) (*model.Comment, error) {
	return s.store.GetComment(id)
}

func (s *Service) ListComments(filter model.CommentFilter, page, size int) ([]*model.Comment, int, error) {
	all := s.store.ListComments()
	matched := make([]*model.Comment, 0, len(all))
	for _, c := range all {
		if filter.Match(c) {
			matched = append(matched, c)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.After(matched[j].CreatedAt) })
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Comment{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) DeleteComment(id string) error {
	return s.store.DeleteComment(id)
}

// Favorite 收藏文档。
func (s *Service) Favorite(userID, documentID string) (*model.Favorite, error) {
	if _, err := s.store.GetDocument(documentID); err != nil {
		return nil, err
	}
	f := &model.Favorite{
		ID:         idgen.Hex(),
		DocumentID: documentID,
		UserID:     userID,
		CreatedAt:  time.Now(),
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateFavorite(f); err != nil {
		return nil, err
	}
	return f, nil
}

// Unfavorite 取消收藏。
func (s *Service) Unfavorite(userID, documentID string) error {
	for _, f := range s.store.ListFavorites() {
		if f.UserID == userID && f.DocumentID == documentID {
			return s.store.DeleteFavorite(f.ID)
		}
	}
	return model.NewValidationError("favorite", "未收藏该文档")
}

func (s *Service) IsFavorited(userID, documentID string) bool {
	for _, f := range s.store.ListFavorites() {
		if f.UserID == userID && f.DocumentID == documentID {
			return true
		}
	}
	return false
}

func (s *Service) DocumentFavoriteCount(documentID string) int {
	count := 0
	for _, f := range s.store.ListFavorites() {
		if f.DocumentID == documentID {
			count++
		}
	}
	return count
}

func (s *Service) ListFavorites(filter model.FavoriteFilter, page, size int) ([]*model.Favorite, int, error) {
	all := s.store.ListFavorites()
	matched := make([]*model.Favorite, 0, len(all))
	for _, f := range all {
		if filter.Match(f) {
			matched = append(matched, f)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.After(matched[j].CreatedAt) })
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Favorite{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}
