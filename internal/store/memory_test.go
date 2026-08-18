package store

import (
	"sync"
	"testing"
	"time"

	"wiki/internal/model"
)

func newTestStore() *MemoryStore { return NewMemoryStore() }

func TestAuthorCRUD(t *testing.T) {
	s := newTestStore()
	if err := s.CreateAuthor(&model.Author{ID: "a1", Name: "张三", Email: "z@x.com"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateAuthor(&model.Author{ID: "a2", Name: "李四", Email: "z@x.com"}); err != ErrConflict {
		t.Fatalf("expect conflict by email, got %v", err)
	}
	got, _ := s.GetAuthor("a1")
	got.Bio = "资深作者"
	if err := s.UpdateAuthor(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteAuthor("a1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetAuthor("a1"); err != ErrNotFound {
		t.Fatalf("expect not found, got %v", err)
	}
}

func TestDirectoryCRUD(t *testing.T) {
	s := newTestStore()
	if err := s.CreateDirectory(&model.Directory{ID: "d1", Name: "技术文档"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateDirectory(&model.Directory{ID: "d2", Name: "技术文档"}); err != ErrConflict {
		t.Fatalf("expect conflict, got %v", err)
	}
	got, _ := s.GetDirectory("d1")
	got.Description = "技术类文档"
	if err := s.UpdateDirectory(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteDirectory("d1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestDocumentCRUD(t *testing.T) {
	s := newTestStore()
	doc := &model.Document{ID: "doc1", Title: "Go 入门", DirectoryID: "d1", AuthorID: "a1"}
	if err := s.CreateDocument(doc); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateDocument(&model.Document{ID: "doc2", Title: "Go 入门", DirectoryID: "d1", AuthorID: "a1"}); err != ErrConflict {
		t.Fatalf("expect conflict by title, got %v", err)
	}
	got, _ := s.GetDocument("doc1")
	got.Content = "正文"
	if err := s.UpdateDocument(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteDocument("doc1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestVersionCRUD(t *testing.T) {
	s := newTestStore()
	v := &model.Version{ID: "v1", DocumentID: "doc1", VersionNo: 1, Content: "内容"}
	if err := s.CreateVersion(v); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := s.GetVersion("v1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if n := len(s.ListVersions()); n != 1 {
		t.Fatalf("list len = %d", n)
	}
	if err := s.DeleteVersion("v1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestTagCRUD(t *testing.T) {
	s := newTestStore()
	if err := s.CreateTag(&model.Tag{ID: "t1", Name: "Go"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateTag(&model.Tag{ID: "t2", Name: "Go"}); err != ErrConflict {
		t.Fatalf("expect conflict, got %v", err)
	}
	if _, err := s.GetTagByName("Go"); err != nil {
		t.Fatalf("get by name: %v", err)
	}
	got, _ := s.GetTag("t1")
	got.Name = "Golang"
	if err := s.UpdateTag(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.DeleteTag("t1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestCommentCRUD(t *testing.T) {
	s := newTestStore()
	c := &model.Comment{ID: "c1", DocumentID: "doc1", AuthorID: "a1", Content: "好文章", CreatedAt: time.Now()}
	if err := s.CreateComment(c); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := s.GetComment("c1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if n := len(s.ListComments()); n != 1 {
		t.Fatalf("list len = %d", n)
	}
	if err := s.DeleteComment("c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestFavoriteCRUD(t *testing.T) {
	s := newTestStore()
	f := &model.Favorite{ID: "f1", DocumentID: "doc1", UserID: "u1"}
	if err := s.CreateFavorite(f); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateFavorite(&model.Favorite{ID: "f2", DocumentID: "doc1", UserID: "u1"}); err != ErrConflict {
		t.Fatalf("expect conflict, got %v", err)
	}
	if _, err := s.GetFavorite("f1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := s.DeleteFavorite("f1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestUpdateNotFound(t *testing.T) {
	s := newTestStore()
	if err := s.UpdateAuthor(&model.Author{ID: "missing"}); err != ErrNotFound {
		t.Fatalf("expect not found, got %v", err)
	}
	if err := s.UpdateDirectory(&model.Directory{ID: "missing"}); err != ErrNotFound {
		t.Fatalf("expect not found, got %v", err)
	}
	if err := s.UpdateDocument(&model.Document{ID: "missing"}); err != ErrNotFound {
		t.Fatalf("expect not found, got %v", err)
	}
	if err := s.UpdateTag(&model.Tag{ID: "missing"}); err != ErrNotFound {
		t.Fatalf("expect not found, got %v", err)
	}
}

func TestUpdateConflicts(t *testing.T) {
	s := newTestStore()

	s.CreateAuthor(&model.Author{ID: "a1", Name: "张三", Email: "a@x.com"})
	s.CreateAuthor(&model.Author{ID: "a2", Name: "李四", Email: "b@x.com"})
	a2, _ := s.GetAuthor("a2")
	a2.Email = "a@x.com"
	if err := s.UpdateAuthor(a2); err != ErrConflict {
		t.Fatalf("author update conflict: got %v", err)
	}

	s.CreateDirectory(&model.Directory{ID: "d1", Name: "技术"})
	s.CreateDirectory(&model.Directory{ID: "d2", Name: "产品"})
	d2, _ := s.GetDirectory("d2")
	d2.Name = "技术"
	if err := s.UpdateDirectory(d2); err != ErrConflict {
		t.Fatalf("directory update conflict: got %v", err)
	}

	s.CreateDocument(&model.Document{ID: "doc1", Title: "A", DirectoryID: "d1", AuthorID: "a1"})
	s.CreateDocument(&model.Document{ID: "doc2", Title: "B", DirectoryID: "d1", AuthorID: "a1"})
	doc2, _ := s.GetDocument("doc2")
	doc2.Title = "A"
	if err := s.UpdateDocument(doc2); err != ErrConflict {
		t.Fatalf("document update conflict: got %v", err)
	}

	s.CreateTag(&model.Tag{ID: "t1", Name: "Go"})
	s.CreateTag(&model.Tag{ID: "t2", Name: "K8s"})
	t2, _ := s.GetTag("t2")
	t2.Name = "Go"
	if err := s.UpdateTag(t2); err != ErrConflict {
		t.Fatalf("tag update conflict: got %v", err)
	}
}

func TestListEmpty(t *testing.T) {
	s := newTestStore()
	if n := len(s.ListAuthors()); n != 0 {
		t.Fatalf("empty authors len = %d", n)
	}
	if n := len(s.ListDirectories()); n != 0 {
		t.Fatalf("empty directories len = %d", n)
	}
	if n := len(s.ListDocuments()); n != 0 {
		t.Fatalf("empty documents len = %d", n)
	}
	if n := len(s.ListVersions()); n != 0 {
		t.Fatalf("empty versions len = %d", n)
	}
	if n := len(s.ListTags()); n != 0 {
		t.Fatalf("empty tags len = %d", n)
	}
	if n := len(s.ListComments()); n != 0 {
		t.Fatalf("empty comments len = %d", n)
	}
	if n := len(s.ListFavorites()); n != 0 {
		t.Fatalf("empty favorites len = %d", n)
	}
}

func TestGetTagByNameNotFound(t *testing.T) {
	s := newTestStore()
	if _, err := s.GetTagByName("不存在"); err != ErrNotFound {
		t.Fatalf("expect not found, got %v", err)
	}
}

func TestConcurrentDocumentListAndWrite(t *testing.T) {
	s := newTestStore()
	for i := 0; i < 20; i++ {
		id := string(rune('a' + i))
		if err := s.CreateDocument(&model.Document{ID: id, Title: "文档" + id, DirectoryID: "d1", AuthorID: "u1"}); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				id := string(rune('k'+offset)) + string(rune('a'+j%20))
				_ = s.CreateDocument(&model.Document{ID: id, Title: "新增" + id, DirectoryID: "d1", AuthorID: "u1"})
				_ = s.DeleteDocument(id)
			}
		}(i)
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 400; j++ {
				_ = s.ListDocuments()
			}
		}()
	}
	close(start)
	wg.Wait()
}
