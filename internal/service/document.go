package service

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"wiki/internal/model"
	"wiki/internal/store"
	"wiki/pkg/idgen"
)

// CreateDocument 创建文档：校验目录/作者存在，规范化标签并建立初始版本。
func (s *Service) CreateDocument(input model.Document) (*model.Document, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetDirectory(input.DirectoryID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetAuthor(input.AuthorID); err != nil {
		return nil, err
	}
	now := time.Now()
	d := &model.Document{
		ID:          idgen.Hex(),
		Title:       input.Title,
		Content:     input.Content,
		DirectoryID: input.DirectoryID,
		AuthorID:    input.AuthorID,
		Tags:        s.normalizeTags(input.Tags),
		Status:      model.DocumentDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.store.CreateDocument(d); err != nil {
		return nil, err
	}
	s.saveVersion(d.ID, 1, d.Content, "初始版本")
	return d, nil
}

// normalizeTags 确保标签存在，返回规范化后的标签名列表。
func (s *Service) normalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, name := range tags {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		if _, err := s.store.GetTagByName(name); err != nil {
			tag := &model.Tag{ID: idgen.Hex(), Name: name, CreatedAt: time.Now()}
			_ = s.store.CreateTag(tag)
		}
		result = append(result, name)
	}
	return result
}

// saveVersion 保存文档版本。
func (s *Service) saveVersion(documentID string, versionNo int, content, changelog string) {
	v := &model.Version{
		ID:         idgen.Hex(),
		DocumentID: documentID,
		VersionNo:  versionNo,
		Content:    content,
		Changelog:  changelog,
		CreatedAt:  time.Now(),
	}
	_ = s.store.CreateVersion(v)
}

// nextVersionNo 返回文档的下一个版本号。
func (s *Service) nextVersionNo(documentID string) int {
	max := 0
	for _, v := range s.store.ListVersions() {
		if v.DocumentID == documentID && v.VersionNo > max {
			max = v.VersionNo
		}
	}
	return max + 1
}

// GetDocument 查询文档；已发布文档浏览量 +1。
func (s *Service) GetDocument(id string) (*model.Document, error) {
	d, err := s.store.GetDocument(id)
	if err != nil {
		return nil, err
	}
	if d.Status == model.DocumentPublished {
		d.ViewCount++
		_ = s.store.UpdateDocument(d)
	}
	return d, nil
}

func (s *Service) ListDocuments(filter model.DocumentFilter, page, size int) ([]*model.Document, int, error) {
	all := s.store.ListDocuments()
	matched := make([]*model.Document, 0, len(all))
	for _, d := range all {
		if filter.Match(d) {
			matched = append(matched, d)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].UpdatedAt.After(matched[j].UpdatedAt) })
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Document{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateDocument 更新文档：仅草稿/已发布可编辑，更新后建立新版本。
func (s *Service) UpdateDocument(id string, input model.Document) (*model.Document, error) {
	existing, err := s.store.GetDocument(id)
	if err != nil {
		return nil, err
	}
	if existing.Status == model.DocumentArchived {
		return nil, model.NewValidationError("status", "已归档文档不可编辑")
	}
	changed := false
	if input.Title != "" && input.Title != existing.Title {
		existing.Title = input.Title
		changed = true
	}
	if input.Content != "" && input.Content != existing.Content {
		existing.Content = input.Content
		changed = true
	}
	if input.Tags != nil {
		existing.Tags = s.normalizeTags(input.Tags)
		changed = true
	}
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateDocument(existing); err != nil {
		return nil, err
	}
	if changed {
		s.saveVersion(id, s.nextVersionNo(id), existing.Content, "更新文档")
	}
	return existing, nil
}

// PublishDocument 发布文档：draft -> published。
func (s *Service) PublishDocument(id string) (*model.Document, error) {
	d, err := s.store.GetDocument(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionDocument(d.Status, model.DocumentPublished) {
		return nil, store.ErrConflict
	}
	if strings.TrimSpace(d.Content) == "" {
		return nil, model.NewValidationError("content", "文档内容为空，无法发布")
	}
	now := time.Now()
	d.Status = model.DocumentPublished
	d.PublishedAt = &now
	d.UpdatedAt = now
	if err := s.store.UpdateDocument(d); err != nil {
		return nil, err
	}
	return d, nil
}

// ArchiveDocument 归档文档：published -> archived。
func (s *Service) ArchiveDocument(id string) (*model.Document, error) {
	d, err := s.store.GetDocument(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionDocument(d.Status, model.DocumentArchived) {
		return nil, store.ErrConflict
	}
	d.Status = model.DocumentArchived
	d.UpdatedAt = time.Now()
	if err := s.store.UpdateDocument(d); err != nil {
		return nil, err
	}
	return d, nil
}

// BatchArchiveDocuments 批量归档文档，返回成功数。
func (s *Service) BatchArchiveDocuments(ids []string) (int, error) {
	success := 0
	for _, id := range ids {
		if _, err := s.ArchiveDocument(id); err == nil {
			success++
		}
	}
	return success, nil
}

// BatchDeleteDocuments 批量删除文档，返回成功数。
func (s *Service) BatchDeleteDocuments(ids []string) (int, error) {
	success := 0
	for _, id := range ids {
		if err := s.DeleteDocument(id); err == nil {
			success++
		}
	}
	return success, nil
}

// DeleteDocument 删除文档及其版本/评论/收藏。
func (s *Service) DeleteDocument(id string) error {
	if err := s.store.DeleteDocument(id); err != nil {
		return err
	}
	for _, v := range s.store.ListVersions() {
		if v.DocumentID == id {
			_ = s.store.DeleteVersion(v.ID)
		}
	}
	for _, c := range s.store.ListComments() {
		if c.DocumentID == id {
			_ = s.store.DeleteComment(c.ID)
		}
	}
	for _, f := range s.store.ListFavorites() {
		if f.DocumentID == id {
			_ = s.store.DeleteFavorite(f.ID)
		}
	}
	return nil
}

// DocumentVersions 返回文档版本列表（按版本号升序）。
func (s *Service) DocumentVersions(documentID string) ([]*model.Version, error) {
	if _, err := s.store.GetDocument(documentID); err != nil {
		return nil, err
	}
	result := make([]*model.Version, 0)
	for _, v := range s.store.ListVersions() {
		if v.DocumentID == documentID {
			result = append(result, v)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].VersionNo < result[j].VersionNo })
	return result, nil
}

// RollbackDocument 回滚文档到指定版本。
func (s *Service) RollbackDocument(documentID, versionID string) (*model.Document, error) {
	d, err := s.store.GetDocument(documentID)
	if err != nil {
		return nil, err
	}
	v, err := s.store.GetVersion(versionID)
	if err != nil {
		return nil, err
	}
	if v.DocumentID != documentID {
		return nil, model.NewValidationError("version_id", "版本不属于该文档")
	}
	d.Content = v.Content
	d.UpdatedAt = time.Now()
	if err := s.store.UpdateDocument(d); err != nil {
		return nil, err
	}
	s.saveVersion(documentID, s.nextVersionNo(documentID), v.Content, "回滚到版本 v"+strconv.Itoa(v.VersionNo))
	return d, nil
}

// SearchDocuments 全文搜索（标题 + 内容）。
func (s *Service) SearchDocuments(keyword string, page, size int) ([]*model.Document, int, error) {
	k := strings.ToLower(strings.TrimSpace(keyword))
	all := s.store.ListDocuments()
	matched := make([]*model.Document, 0)
	for _, d := range all {
		if d.Status != model.DocumentPublished {
			continue
		}
		if k == "" || strings.Contains(strings.ToLower(d.Title), k) ||
			strings.Contains(strings.ToLower(d.Content), k) {
			matched = append(matched, d)
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].ViewCount > matched[j].ViewCount })
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Document{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}
