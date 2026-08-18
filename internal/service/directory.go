package service

import (
	"sort"
	"time"

	"wiki/internal/model"
	"wiki/pkg/idgen"
)

func (s *Service) CreateDirectory(input model.Directory) (*model.Directory, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if input.ParentID != "" {
		if _, err := s.store.GetDirectory(input.ParentID); err != nil {
			return nil, err
		}
	}
	d := &model.Directory{
		ID:          idgen.Hex(),
		Name:        input.Name,
		ParentID:    input.ParentID,
		Description: input.Description,
		CreatedAt:   time.Now(),
	}
	if err := s.store.CreateDirectory(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) GetDirectory(id string) (*model.Directory, error) {
	return s.store.GetDirectory(id)
}

func (s *Service) ListDirectories(filter model.DirectoryFilter, page, size int) ([]*model.Directory, int, error) {
	all := s.store.ListDirectories()
	matched := make([]*model.Directory, 0, len(all))
	for _, d := range all {
		if filter.Match(d) {
			matched = append(matched, d)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].CreatedAt.After(matched[j].CreatedAt) })
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Directory{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateDirectory(id string, input model.Directory) (*model.Directory, error) {
	existing, err := s.store.GetDirectory(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.Description != "" {
		existing.Description = input.Description
	}
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateDirectory(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteDirectory(id string) error {
	// 跨实体校验：目录下仍有子目录或文档时不允许删除
	for _, d := range s.store.ListDirectories() {
		if d.ParentID == id {
			return model.NewValidationError("directory", "该目录下仍有子目录，无法删除")
		}
	}
	for _, doc := range s.store.ListDocuments() {
		if doc.DirectoryID == id {
			return model.NewValidationError("directory", "该目录下仍有文档，无法删除")
		}
	}
	return s.store.DeleteDirectory(id)
}

// DirectoryNode 目录树节点。
type DirectoryNode struct {
	Directory     model.Directory  `json:"directory"`
	DocumentCount int              `json:"document_count"`
	Children      []*DirectoryNode `json:"children,omitempty"`
}

// DirectoryTree 构建目录树（按父子关系嵌套）。
func (s *Service) DirectoryTree() ([]*DirectoryNode, error) {
	dirs := s.store.ListDirectories()
	docs := s.store.ListDocuments()

	// 统计每个目录直接文档数
	docCount := map[string]int{}
	for _, d := range docs {
		docCount[d.DirectoryID]++
	}

	nodes := make(map[string]*DirectoryNode)
	for _, d := range dirs {
		nodes[d.ID] = &DirectoryNode{Directory: *d, DocumentCount: docCount[d.ID]}
	}

	roots := make([]*DirectoryNode, 0)
	for _, d := range dirs {
		node := nodes[d.ID]
		if d.ParentID == "" {
			roots = append(roots, node)
		} else if parent, ok := nodes[d.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].Directory.Name < roots[j].Directory.Name })
	for _, n := range nodes {
		sort.Slice(n.Children, func(i, j int) bool { return n.Children[i].Directory.Name < n.Children[j].Directory.Name })
	}
	return roots, nil
}
