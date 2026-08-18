package service

import (
	"sort"
	"time"

	"wiki/internal/model"
	"wiki/pkg/idgen"
)

func (s *Service) CreateAuthor(input model.Author) (*model.Author, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	a := &model.Author{
		ID:        idgen.Hex(),
		Name:      input.Name,
		Email:     input.Email,
		Bio:       input.Bio,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateAuthor(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) GetAuthor(id string) (*model.Author, error) {
	return s.store.GetAuthor(id)
}

func (s *Service) ListAuthors(filter model.AuthorFilter, page, size int) ([]*model.Author, int, error) {
	all := s.store.ListAuthors()
	matched := make([]*model.Author, 0, len(all))
	for _, a := range all {
		if filter.Match(a) {
			matched = append(matched, a)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.After(matched[j].CreatedAt) })
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Author{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateAuthor(id string, input model.Author) (*model.Author, error) {
	existing, err := s.store.GetAuthor(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.Email != "" {
		existing.Email = input.Email
	}
	if input.Bio != "" {
		existing.Bio = input.Bio
	}
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateAuthor(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteAuthor(id string) error {
	// 跨实体校验：作者仍有文档时不允许删除
	for _, d := range s.store.ListDocuments() {
		if d.AuthorID == id {
			return model.NewValidationError("author", "该作者仍有文档，无法删除")
		}
	}
	return s.store.DeleteAuthor(id)
}
