package service

import (
	"sort"

	"wiki/internal/model"
)

// KeyCount 通用键值计数。
type KeyCount struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// DocumentStats 文档统计。
type DocumentStats struct {
	Total       int        `json:"total"`
	ByStatus    []KeyCount `json:"by_status"`
	ByDirectory []KeyCount `json:"by_directory"`
	ByAuthor    []KeyCount `json:"by_author"`
}

// DocumentStats 计算文档多维度统计。
func (s *Service) DocumentStats() (*DocumentStats, error) {
	docs := s.store.ListDocuments()
	stats := &DocumentStats{Total: len(docs)}
	statusMap := map[string]int{}
	dirMap := map[string]int{}
	authorMap := map[string]int{}
	for _, d := range docs {
		statusMap[d.Status]++
		dirMap[d.DirectoryID]++
		authorMap[d.AuthorID]++
	}
	stats.ByStatus = keyCounts(statusMap, func(k string) string { return k })
	stats.ByDirectory = keyCounts(dirMap, func(k string) string {
		if d, err := s.store.GetDirectory(k); err == nil {
			return d.Name
		}
		return k
	})
	stats.ByAuthor = keyCounts(authorMap, func(k string) string {
		if a, err := s.store.GetAuthor(k); err == nil {
			return a.Name
		}
		return k
	})
	return stats, nil
}

// TagCount 标签统计。
type TagCount struct {
	Name     string `json:"name"`
	DocCount int    `json:"doc_count"`
}

// TagStats 统计各标签的文档数。
func (s *Service) TagStats() ([]TagCount, error) {
	docs := s.store.ListDocuments()
	countMap := map[string]int{}
	for _, d := range docs {
		for _, t := range d.Tags {
			countMap[t]++
		}
	}
	result := make([]TagCount, 0, len(countMap))
	for name, c := range countMap {
		result = append(result, TagCount{Name: name, DocCount: c})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DocCount > result[j].DocCount })
	return result, nil
}

// AuthorStats 作者统计。
type AuthorStats struct {
	AuthorID   string `json:"author_id"`
	AuthorName string `json:"author_name"`
	Documents  int    `json:"documents"`
	TotalViews int    `json:"total_views"`
}

// ListAuthorStats 统计各作者的文档数与浏览量。
func (s *Service) ListAuthorStats() ([]AuthorStats, error) {
	authors := s.store.ListAuthors()
	result := make([]AuthorStats, 0, len(authors))
	for _, a := range authors {
		as := AuthorStats{AuthorID: a.ID, AuthorName: a.Name}
		for _, d := range s.store.ListDocuments() {
			if d.AuthorID == a.ID {
				as.Documents++
				as.TotalViews += d.ViewCount
			}
		}
		result = append(result, as)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Documents > result[j].Documents })
	return result, nil
}

// PopularDocument 热门文档。
type PopularDocument struct {
	DocumentID    string `json:"document_id"`
	DocumentTitle string `json:"document_title"`
	Views         int    `json:"views"`
	Favorites     int    `json:"favorites"`
}

// PopularDocuments 按浏览量返回热门文档 TOP N。
func (s *Service) PopularDocuments(n int) ([]PopularDocument, error) {
	if n <= 0 {
		n = 10
	}
	docs := s.store.ListDocuments()
	result := make([]PopularDocument, 0, len(docs))
	for _, d := range docs {
		if d.Status != model.DocumentPublished {
			continue
		}
		result = append(result, PopularDocument{
			DocumentID:    d.ID,
			DocumentTitle: d.Title,
			Views:         d.ViewCount,
			Favorites:     s.DocumentFavoriteCount(d.ID),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Views > result[j].Views })
	if len(result) > n {
		result = result[:n]
	}
	return result, nil
}

// keyCounts 将 map 计数转为排序后的 KeyCount 列表。
func keyCounts(m map[string]int, name func(string) string) []KeyCount {
	result := make([]KeyCount, 0, len(m))
	for k, c := range m {
		result = append(result, KeyCount{Key: name(k), Count: c})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Count > result[j].Count })
	return result
}

// MonthlyCount 按月聚合的文档数量。
type MonthlyCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

// MonthlyStats 按月统计文档创建数量。
func (s *Service) MonthlyStats() ([]MonthlyCount, error) {
	docs := s.store.ListDocuments()
	countMap := map[string]int{}
	for _, d := range docs {
		key := d.CreatedAt.Format("2006-01")
		countMap[key]++
	}
	result := make([]MonthlyCount, 0, len(countMap))
	for m, c := range countMap {
		result = append(result, MonthlyCount{Month: m, Count: c})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Month > result[j].Month })
	return result, nil
}

// OverviewStats 知识库总览统计。
type OverviewStats struct {
	Documents   int `json:"documents"`
	Published   int `json:"published"`
	Drafts      int `json:"drafts"`
	Archived    int `json:"archived"`
	Authors     int `json:"authors"`
	Directories int `json:"directories"`
	Tags        int `json:"tags"`
	Comments    int `json:"comments"`
	Favorites   int `json:"favorites"`
	TotalViews  int `json:"total_views"`
}

// OverviewStats 计算知识库总览指标。
func (s *Service) OverviewStats() (*OverviewStats, error) {
	stats := &OverviewStats{}
	for _, d := range s.store.ListDocuments() {
		stats.Documents++
		stats.TotalViews += d.ViewCount
		switch d.Status {
		case model.DocumentPublished:
			stats.Published++
		case model.DocumentDraft:
			stats.Drafts++
		case model.DocumentArchived:
			stats.Archived++
		}
	}
	stats.Authors = len(s.store.ListAuthors())
	stats.Directories = len(s.store.ListDirectories())
	stats.Tags = len(s.store.ListTags())
	stats.Comments = len(s.store.ListComments())
	stats.Favorites = len(s.store.ListFavorites())
	return stats, nil
}
