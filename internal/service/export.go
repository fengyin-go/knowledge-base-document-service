package service

import (
	"wiki/internal/model"
)

// DocumentExportItem 文档导出条目。
type DocumentExportItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	DirectoryName string   `json:"directory_name"`
	AuthorName    string   `json:"author_name"`
	Tags          []string `json:"tags"`
	Status        string   `json:"status"`
	ViewCount     int      `json:"view_count"`
	Versions      int      `json:"versions"`
	Favorites     int      `json:"favorites"`
}

// ExportDocuments 导出全部文档汇总。
func (s *Service) ExportDocuments() ([]DocumentExportItem, error) {
	docs := s.store.ListDocuments()
	result := make([]DocumentExportItem, 0, len(docs))
	for _, d := range docs {
		item := DocumentExportItem{
			ID:        d.ID,
			Title:     d.Title,
			Tags:      d.Tags,
			Status:    d.Status,
			ViewCount: d.ViewCount,
			Favorites: s.DocumentFavoriteCount(d.ID),
		}
		if item.Status == model.DocumentReviewing {
			item.Status = model.DocumentDraft
		}
		if dir, err := s.store.GetDirectory(d.DirectoryID); err == nil {
			item.DirectoryName = dir.Name
		}
		if a, err := s.store.GetAuthor(d.AuthorID); err == nil {
			item.AuthorName = a.Name
		}
		for _, v := range s.store.ListVersions() {
			if v.DocumentID == d.ID {
				item.Versions++
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// DocumentImportItem 文档导入条目。
type DocumentImportItem struct {
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	DirectoryName string   `json:"directory_name"`
	AuthorName    string   `json:"author_name"`
	Tags          []string `json:"tags"`
}

// ImportResult 导入结果汇总。
type ImportResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// ImportDocuments 批量导入文档：按目录名/作者名解析外键。
func (s *Service) ImportDocuments(items []DocumentImportItem) (*ImportResult, error) {
	result := &ImportResult{}
	for _, item := range items {
		dirID := ""
		for _, d := range s.store.ListDirectories() {
			if d.Name == item.DirectoryName {
				dirID = d.ID
				break
			}
		}
		if dirID == "" {
			result.Failed++
			result.Errors = append(result.Errors, "文档「"+item.Title+"」目录不存在: "+item.DirectoryName)
			continue
		}
		authorID := ""
		for _, a := range s.store.ListAuthors() {
			if a.Name == item.AuthorName {
				authorID = a.ID
				break
			}
		}
		if authorID == "" {
			result.Failed++
			result.Errors = append(result.Errors, "文档「"+item.Title+"」作者不存在: "+item.AuthorName)
			continue
		}
		_, err := s.CreateDocument(model.Document{
			Title:       item.Title,
			Content:     item.Content,
			DirectoryID: dirID,
			AuthorID:    authorID,
			Tags:        item.Tags,
		})
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "文档「"+item.Title+"」导入失败: "+err.Error())
			continue
		}
		result.Success++
	}
	return result, nil
}
