package service

import (
	"testing"

	"wiki/internal/config"
	"wiki/internal/model"
	"wiki/internal/store"
	"wiki/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

func setupDoc(t *testing.T) (*Service, *model.Author, *model.Directory, *model.Document) {
	t.Helper()
	s := newTestService()
	author, err := s.CreateAuthor(model.Author{Name: "张三", Email: "z@x.com"})
	if err != nil {
		t.Fatalf("create author: %v", err)
	}
	dir, err := s.CreateDirectory(model.Directory{Name: "技术文档"})
	if err != nil {
		t.Fatalf("create directory: %v", err)
	}
	doc, err := s.CreateDocument(model.Document{
		Title: "Go 入门", Content: "Go 是一门编程语言", DirectoryID: dir.ID, AuthorID: author.ID, Tags: []string{"Go"},
	})
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	return s, author, dir, doc
}

func TestCreateDocumentValidations(t *testing.T) {
	s := newTestService()
	if _, err := s.CreateDocument(model.Document{Title: "x", DirectoryID: "missing", AuthorID: "missing"}); err == nil {
		t.Fatalf("expect error for missing directory")
	}
}

func TestDocumentFullLifecycle(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	if doc.Status != model.DocumentDraft {
		t.Fatalf("initial status = %s", doc.Status)
	}
	// 版本 1 已建立
	versions, _ := s.DocumentVersions(doc.ID)
	if len(versions) != 1 || versions[0].VersionNo != 1 {
		t.Fatalf("initial versions = %d", len(versions))
	}
	// 发布
	doc, err := s.PublishDocument(doc.ID)
	if err != nil || doc.Status != model.DocumentPublished {
		t.Fatalf("publish: %v status=%s", err, doc.Status)
	}
	// 更新 -> 版本 2
	doc, _ = s.UpdateDocument(doc.ID, model.Document{Content: "Go 是一门编译型语言"})
	versions, _ = s.DocumentVersions(doc.ID)
	if len(versions) != 2 || versions[1].VersionNo != 2 {
		t.Fatalf("after update versions = %d", len(versions))
	}
	// 归档
	doc, err = s.ArchiveDocument(doc.ID)
	if err != nil || doc.Status != model.DocumentArchived {
		t.Fatalf("archive: %v status=%s", err, doc.Status)
	}
}

func TestDocumentPublishRequiresContent(t *testing.T) {
	s, author, dir, _ := setupDoc(t)
	// 空内容文档无法发布
	empty, _ := s.CreateDocument(model.Document{Title: "空文档", DirectoryID: dir.ID, AuthorID: author.ID})
	if _, err := s.PublishDocument(empty.ID); err == nil {
		t.Fatalf("expect error publishing empty document")
	}
}

func TestDocumentInvalidTransitions(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	s.ArchiveDocument(doc.ID)
	// archived 不能直接 publish（需通过 republish，但当前状态机 archived->published 允许）
	// 已归档文档不可更新
	if _, err := s.UpdateDocument(doc.ID, model.Document{Content: "修改"}); err == nil {
		t.Fatalf("expect error updating archived document")
	}
}

func TestDocumentRollback(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	doc, _ = s.UpdateDocument(doc.ID, model.Document{Content: "第二版内容"})
	versions, _ := s.DocumentVersions(doc.ID)
	// 回滚到版本 1
	doc, err := s.RollbackDocument(doc.ID, versions[0].ID)
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if doc.Content != "Go 是一门编程语言" {
		t.Fatalf("after rollback content = %s", doc.Content)
	}
	// 回滚产生新版本
	versions, _ = s.DocumentVersions(doc.ID)
	if len(versions) != 3 {
		t.Fatalf("after rollback versions = %d, want 3", len(versions))
	}
}

func TestDocumentViewCount(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	s.GetDocument(doc.ID)
	s.GetDocument(doc.ID)
	d, _ := s.GetDocument(doc.ID)
	if d.ViewCount != 3 {
		t.Fatalf("view count = %d, want 3", d.ViewCount)
	}
}

func TestSearchDocuments(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	items, total, err := s.SearchDocuments("编程语言", 1, 10)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("search: %v total=%d len=%d", err, total, len(items))
	}
	_, total, _ = s.SearchDocuments("不存在关键词", 1, 10)
	if total != 0 {
		t.Fatalf("empty search total = %d", total)
	}
}

func TestTagNormalization(t *testing.T) {
	s, author, dir, _ := setupDoc(t)
	// 创建带新标签的文档，标签应自动建立
	s.CreateDocument(model.Document{Title: "K8s 入门", Content: "x", DirectoryID: dir.ID, AuthorID: author.ID, Tags: []string{"K8s", "Go"}})
	// "Go" 标签已存在，"K8s" 标签应被创建
	if _, err := s.store.GetTagByName("K8s"); err != nil {
		t.Fatalf("tag K8s should be auto-created")
	}
	stats, _ := s.TagStats()
	foundGo := false
	for _, tc := range stats {
		if tc.Name == "Go" && tc.DocCount == 2 {
			foundGo = true
		}
	}
	if !foundGo {
		t.Fatalf("Go tag count should be 2, got %+v", stats)
	}
}

func TestCommentService(t *testing.T) {
	s, author, _, doc := setupDoc(t)
	c, err := s.CreateComment(model.Comment{DocumentID: doc.ID, AuthorID: author.ID, Content: "好文章"})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	// 评论不存在的文档 -> 404
	if _, err := s.CreateComment(model.Comment{DocumentID: "missing", AuthorID: author.ID, Content: "x"}); err == nil {
		t.Fatalf("expect error commenting missing document")
	}
	items, total, _ := s.ListComments(model.CommentFilter{DocumentID: doc.ID}, 1, 10)
	if total != 1 || len(items) != 1 || items[0].ID != c.ID {
		t.Fatalf("list comments: total=%d len=%d", total, len(items))
	}
}

func TestFavoriteService(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	if _, err := s.Favorite("u1", doc.ID); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	if _, err := s.Favorite("u1", doc.ID); err == nil {
		t.Fatalf("expect conflict on duplicate favorite")
	}
	if !s.IsFavorited("u1", doc.ID) {
		t.Fatalf("expect favorited")
	}
	if s.DocumentFavoriteCount(doc.ID) != 1 {
		t.Fatalf("favorite count = %d", s.DocumentFavoriteCount(doc.ID))
	}
	if err := s.Unfavorite("u1", doc.ID); err != nil {
		t.Fatalf("unfavorite: %v", err)
	}
	if s.IsFavorited("u1", doc.ID) {
		t.Fatalf("expect not favorited")
	}
	if err := s.Unfavorite("u1", doc.ID); err == nil {
		t.Fatalf("expect error on re-unfavorite")
	}
}

func TestDeleteDocumentCascade(t *testing.T) {
	s, author, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	s.CreateComment(model.Comment{DocumentID: doc.ID, AuthorID: author.ID, Content: "评论"})
	s.Favorite("u1", doc.ID)

	if err := s.DeleteDocument(doc.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// 评论、收藏、版本应级联删除
	if _, err := s.GetDocument(doc.ID); err == nil {
		t.Fatalf("expect document deleted")
	}
	versions, _ := s.DocumentVersions(doc.ID)
	if len(versions) != 0 {
		t.Fatalf("versions should be empty")
	}
	comments, total, _ := s.ListComments(model.CommentFilter{DocumentID: doc.ID}, 1, 10)
	if total != 0 || len(comments) != 0 {
		t.Fatalf("comments should be empty")
	}
}

func TestAuthorDeleteBlocked(t *testing.T) {
	s, author, _, _ := setupDoc(t)
	if err := s.DeleteAuthor(author.ID); err == nil {
		t.Fatalf("expect error deleting author with documents")
	}
}

func TestDirectoryDeleteBlocked(t *testing.T) {
	s, _, dir, _ := setupDoc(t)
	if err := s.DeleteDirectory(dir.ID); err == nil {
		t.Fatalf("expect error deleting directory with documents")
	}
}

func TestDocumentStats(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	stats, err := s.DocumentStats()
	if err != nil || stats.Total != 1 {
		t.Fatalf("document stats: %v total=%d", err, stats.Total)
	}
	popular, err := s.PopularDocuments(10)
	if err != nil || len(popular) != 1 {
		t.Fatalf("popular docs: %v len=%d", err, len(popular))
	}
	authorStats, err := s.ListAuthorStats()
	if err != nil || len(authorStats) != 1 {
		t.Fatalf("author stats: %v len=%d", err, len(authorStats))
	}
	if authorStats[0].Documents != 1 {
		t.Fatalf("author documents = %d", authorStats[0].Documents)
	}
}

func TestExportAndImport(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)

	exported, err := s.ExportDocuments()
	if err != nil || len(exported) != 1 {
		t.Fatalf("export: %v len=%d", err, len(exported))
	}
	if exported[0].DirectoryName != "技术文档" || exported[0].AuthorName != "张三" {
		t.Fatalf("exported = %+v", exported[0])
	}
	if exported[0].Versions != 1 {
		t.Fatalf("exported versions = %d", exported[0].Versions)
	}

	// 导入新文档
	result, err := s.ImportDocuments([]DocumentImportItem{
		{Title: "Python 入门", Content: "x", DirectoryName: "技术文档", AuthorName: "张三", Tags: []string{"Python"}},
		{Title: "无效文档", Content: "x", DirectoryName: "不存在目录", AuthorName: "张三"},
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if result.Success != 1 || result.Failed != 1 {
		t.Fatalf("import success=%d failed=%d", result.Success, result.Failed)
	}
}

func TestListPagination(t *testing.T) {
	s, author, dir, _ := setupDoc(t)
	for i := 0; i < 5; i++ {
		s.CreateDocument(model.Document{Title: "文档" + string(rune('A'+i)), Content: "x", DirectoryID: dir.ID, AuthorID: author.ID})
	}
	items, total, _ := s.ListDocuments(model.DocumentFilter{}, 1, 2)
	if total != 6 || len(items) != 2 {
		t.Fatalf("page1 total=%d len=%d", total, len(items))
	}
	// 按目录过滤
	items, total, _ = s.ListDocuments(model.DocumentFilter{DirectoryID: dir.ID}, 1, 20)
	if total != 6 || len(items) != 6 {
		t.Fatalf("by directory total=%d len=%d", total, len(items))
	}
}

func TestDirectoryTree(t *testing.T) {
	s := newTestService()
	root, _ := s.CreateDirectory(model.Directory{Name: "技术文档"})
	child, _ := s.CreateDirectory(model.Directory{Name: "后端", ParentID: root.ID})
	s.CreateDirectory(model.Directory{Name: "前端", ParentID: root.ID})
	author, _ := s.CreateAuthor(model.Author{Name: "张三", Email: "z@x.com"})
	s.CreateDocument(model.Document{Title: "Go 入门", Content: "x", DirectoryID: child.ID, AuthorID: author.ID})

	tree, err := s.DirectoryTree()
	if err != nil {
		t.Fatalf("tree: %v", err)
	}
	if len(tree) != 1 || tree[0].Directory.Name != "技术文档" {
		t.Fatalf("root len=%d", len(tree))
	}
	if len(tree[0].Children) != 2 {
		t.Fatalf("children len = %d, want 2", len(tree[0].Children))
	}
	// 后端目录下有 1 篇文档
	for _, c := range tree[0].Children {
		if c.Directory.Name == "后端" && c.DocumentCount != 1 {
			t.Fatalf("后端 document count = %d", c.DocumentCount)
		}
	}
}

func TestBatchArchiveAndDelete(t *testing.T) {
	s, author, dir, _ := setupDoc(t)
	var ids []string
	for i := 0; i < 3; i++ {
		d, _ := s.CreateDocument(model.Document{Title: "批量" + string(rune('A'+i)), Content: "x", DirectoryID: dir.ID, AuthorID: author.ID})
		s.PublishDocument(d.ID)
		ids = append(ids, d.ID)
	}
	n, err := s.BatchArchiveDocuments(ids)
	if err != nil || n != 3 {
		t.Fatalf("batch archive: %v n=%d", err, n)
	}
	// 已归档的不能再归档
	n, _ = s.BatchArchiveDocuments(ids)
	if n != 0 {
		t.Fatalf("re-archive success = %d, want 0", n)
	}
	// 批量删除
	n, err = s.BatchDeleteDocuments(ids)
	if err != nil || n != 3 {
		t.Fatalf("batch delete: %v n=%d", err, n)
	}
}

func TestDocumentFilterByTag(t *testing.T) {
	s, author, dir, _ := setupDoc(t)
	s.CreateDocument(model.Document{Title: "K8s 文档", Content: "x", DirectoryID: dir.ID, AuthorID: author.ID, Tags: []string{"K8s"}})
	items, total, _ := s.ListDocuments(model.DocumentFilter{Tag: "Go"}, 1, 20)
	if total != 1 || len(items) != 1 {
		t.Fatalf("filter by Go tag: total=%d", total)
	}
	items, total, _ = s.ListDocuments(model.DocumentFilter{Tag: "K8s"}, 1, 20)
	if total != 1 {
		t.Fatalf("filter by K8s tag: total=%d", total)
	}
}

func TestDirectoryHierarchy(t *testing.T) {
	s := newTestService()
	parent, _ := s.CreateDirectory(model.Directory{Name: "根"})
	child, _ := s.CreateDirectory(model.Directory{Name: "子", ParentID: parent.ID})
	// 有子目录的父目录不能删除
	if err := s.DeleteDirectory(parent.ID); err == nil {
		t.Fatalf("expect error deleting parent with child")
	}
	// 删除子目录后父目录可删
	if err := s.DeleteDirectory(child.ID); err != nil {
		t.Fatalf("delete child: %v", err)
	}
	if err := s.DeleteDirectory(parent.ID); err != nil {
		t.Fatalf("delete parent: %v", err)
	}
}

func TestOverviewAndMonthlyStats(t *testing.T) {
	s, author, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	s.CreateComment(model.Comment{DocumentID: doc.ID, AuthorID: author.ID, Content: "评论"})
	s.Favorite("u1", doc.ID)

	overview, err := s.OverviewStats()
	if err != nil {
		t.Fatalf("overview: %v", err)
	}
	if overview.Documents != 1 || overview.Published != 1 || overview.Comments != 1 || overview.Favorites != 1 {
		t.Fatalf("overview = %+v", overview)
	}

	monthly, err := s.MonthlyStats()
	if err != nil || len(monthly) != 1 {
		t.Fatalf("monthly: %v len=%d", err, len(monthly))
	}
}

func TestAuthorServiceEdgeCases(t *testing.T) {
	s := newTestService()
	author, _ := s.CreateAuthor(model.Author{Name: "张三", Email: "z@x.com"})
	// 重复邮箱冲突
	if _, err := s.CreateAuthor(model.Author{Name: "李四", Email: "z@x.com"}); err == nil {
		t.Fatalf("expect conflict on duplicate email")
	}
	// 更新
	updated, err := s.UpdateAuthor(author.ID, model.Author{Bio: "资深"})
	if err != nil || updated.Bio != "资深" {
		t.Fatalf("update author: %v", err)
	}
	// 列表过滤
	items, total, _ := s.ListAuthors(model.AuthorFilter{Keyword: "张"}, 1, 10)
	if total != 1 || len(items) != 1 {
		t.Fatalf("list authors: total=%d len=%d", total, len(items))
	}
}

func TestTagService(t *testing.T) {
	s := newTestService()
	tag, err := s.CreateTag(model.Tag{Name: "Go"})
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if _, err := s.CreateTag(model.Tag{Name: "Go"}); err == nil {
		t.Fatalf("expect conflict on duplicate tag")
	}
	if _, err := s.GetTag(tag.ID); err != nil {
		t.Fatalf("get tag: %v", err)
	}
	items, total, _ := s.ListTags(1, 10)
	if total != 1 || len(items) != 1 {
		t.Fatalf("list tags: total=%d len=%d", total, len(items))
	}
	if err := s.DeleteTag(tag.ID); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
}

func TestCommentAndFavoriteEdgeCases(t *testing.T) {
	s, author, _, doc := setupDoc(t)
	// 评论内容为空 -> 校验失败
	if _, err := s.CreateComment(model.Comment{DocumentID: doc.ID, AuthorID: author.ID, Content: ""}); err == nil {
		t.Fatalf("expect error for empty comment")
	}
	// 评论不存在的作者 -> 404
	if _, err := s.CreateComment(model.Comment{DocumentID: doc.ID, AuthorID: "missing", Content: "x"}); err == nil {
		t.Fatalf("expect error for missing author")
	}
	// 收藏不存在的文档 -> 404
	if _, err := s.Favorite("u1", "missing"); err == nil {
		t.Fatalf("expect error favoriting missing document")
	}
}

func TestDocumentGetViewCountNotIncrementForDraft(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	// 草稿状态查看不增加浏览量
	s.GetDocument(doc.ID)
	d, _ := s.GetDocument(doc.ID)
	if d.ViewCount != 0 {
		t.Fatalf("draft view count = %d, want 0", d.ViewCount)
	}
}

func TestRollbackWrongVersion(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	// 用另一个文档的版本回滚 -> 报错
	other, _ := s.CreateDocument(model.Document{Title: "另一个文档", Content: "x", DirectoryID: doc.DirectoryID, AuthorID: doc.AuthorID})
	otherVersion, _ := s.DocumentVersions(other.ID)
	if _, err := s.RollbackDocument(doc.ID, otherVersion[0].ID); err == nil {
		t.Fatalf("expect error rolling back to another document's version")
	}
}

func TestUpdateDocumentTags(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	doc, err := s.UpdateDocument(doc.ID, model.Document{Tags: []string{"Go", "K8s"}})
	if err != nil {
		t.Fatalf("update tags: %v", err)
	}
	if len(doc.Tags) != 2 || !doc.HasTag("K8s") {
		t.Fatalf("updated tags = %v", doc.Tags)
	}
	// K8s 标签被自动创建
	if _, err := s.store.GetTagByName("K8s"); err != nil {
		t.Fatalf("K8s tag should be auto-created")
	}
}

func TestSearchDocumentsPagination(t *testing.T) {
	s, author, dir, _ := setupDoc(t)
	// 创建多篇含关键词的文档
	for i := 0; i < 5; i++ {
		d, _ := s.CreateDocument(model.Document{Title: "Go 进阶" + string(rune('A'+i)), Content: "Go 编程", DirectoryID: dir.ID, AuthorID: author.ID})
		s.PublishDocument(d.ID)
	}
	items, total, _ := s.SearchDocuments("Go", 1, 2)
	if total < 5 || len(items) != 2 {
		t.Fatalf("search page1 total=%d len=%d", total, len(items))
	}
}

func TestDirectoryUpdateAndList(t *testing.T) {
	s := newTestService()
	d, _ := s.CreateDirectory(model.Directory{Name: "技术文档"})
	updated, err := s.UpdateDirectory(d.ID, model.Directory{Description: "技术类"})
	if err != nil || updated.Description != "技术类" {
		t.Fatalf("update directory: %v", err)
	}
	items, total, _ := s.ListDirectories(model.DirectoryFilter{Keyword: "技术"}, 1, 10)
	if total != 1 || len(items) != 1 {
		t.Fatalf("list directories: total=%d len=%d", total, len(items))
	}
}

func TestGetNotFound(t *testing.T) {
	s, _, _, _ := setupDoc(t)
	if _, err := s.GetAuthor("missing"); err == nil {
		t.Fatalf("expect not found for author")
	}
	if _, err := s.GetDirectory("missing"); err == nil {
		t.Fatalf("expect not found for directory")
	}
	if _, err := s.GetDocument("missing"); err == nil {
		t.Fatalf("expect not found for document")
	}
	if _, err := s.GetTag("missing"); err == nil {
		t.Fatalf("expect not found for tag")
	}
	if _, err := s.GetComment("missing"); err == nil {
		t.Fatalf("expect not found for comment")
	}
}

func TestRepublishArchivedDocument(t *testing.T) {
	s, _, _, doc := setupDoc(t)
	s.PublishDocument(doc.ID)
	s.ArchiveDocument(doc.ID)
	// archived -> published 重新发布
	doc, err := s.PublishDocument(doc.ID)
	if err != nil || doc.Status != model.DocumentPublished {
		t.Fatalf("republish: %v status=%s", err, doc.Status)
	}
}

func TestSearchEmptyKeyword(t *testing.T) {
	s, author, dir, _ := setupDoc(t)
	for i := 0; i < 3; i++ {
		d, _ := s.CreateDocument(model.Document{Title: "文档" + string(rune('A'+i)), Content: "x", DirectoryID: dir.ID, AuthorID: author.ID})
		s.PublishDocument(d.ID)
	}
	items, total, _ := s.SearchDocuments("", 1, 10)
	if total != 3 || len(items) != 3 {
		t.Fatalf("empty search total=%d len=%d", total, len(items))
	}
}

func TestDocumentDeleteNonExistent(t *testing.T) {
	s, _, _, _ := setupDoc(t)
	if err := s.DeleteDocument("missing"); err == nil {
		t.Fatalf("expect error deleting missing document")
	}
}

func TestVersionListOnMissingDocument(t *testing.T) {
	s, _, _, _ := setupDoc(t)
	if _, err := s.DocumentVersions("missing"); err == nil {
		t.Fatalf("expect error listing versions of missing document")
	}
}

func TestListAuthorsPagination(t *testing.T) {
	s := newTestService()
	for i := 0; i < 5; i++ {
		s.CreateAuthor(model.Author{Name: "作者" + string(rune('A'+i)), Email: string(rune('a'+i)) + "@x.com"})
	}
	items, total, _ := s.ListAuthors(model.AuthorFilter{}, 1, 2)
	if total != 5 || len(items) != 2 {
		t.Fatalf("authors page1 total=%d len=%d", total, len(items))
	}
}
