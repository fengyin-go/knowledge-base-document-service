package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wiki/internal/config"
	"wiki/internal/model"
	"wiki/internal/service"
	"wiki/internal/store"
	"wiki/pkg/httpx"
	"wiki/pkg/logger"
)

func newTestServer(cfg *config.Config) *Server {
	if cfg == nil {
		cfg = &config.Config{MaxPageSize: 100}
	}
	log := logger.NewLevel(logger.LevelError)
	st := store.NewMemoryStore()
	svc := service.New(st, log, cfg)
	return NewServer(svc, log, cfg)
}

func doRequest(s *Server, method, path string, body interface{}) (*httptest.ResponseRecorder, *httpx.Response) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	resp := &httpx.Response{}
	_ = json.Unmarshal(rec.Body.Bytes(), resp)
	return rec, resp
}

func decodeData(t *testing.T, resp *httpx.Response, dst interface{}) {
	t.Helper()
	raw, _ := json.Marshal(resp.Data)
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatalf("decode data: %v", err)
	}
}

func setupAuthorAndDir(t *testing.T, s *Server) (*model.Author, *model.Directory) {
	t.Helper()
	_, resp := doRequest(s, http.MethodPost, "/api/authors", map[string]string{"name": "张三", "email": "z@x.com"})
	var a model.Author
	decodeData(t, resp, &a)
	_, resp = doRequest(s, http.MethodPost, "/api/directories", map[string]string{"name": "技术文档"})
	var d model.Directory
	decodeData(t, resp, &d)
	return &a, &d
}

func TestAuthorEndpoints(t *testing.T) {
	s := newTestServer(nil)
	_, resp := doRequest(s, http.MethodPost, "/api/authors", map[string]string{"name": "张三", "email": "z@x.com"})
	if resp.Code != 0 {
		t.Fatalf("create author: %s", resp.Message)
	}
	rec, _ := doRequest(s, http.MethodPost, "/api/authors", map[string]string{"name": "李四", "email": "z@x.com"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate author status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodGet, "/api/authors/missing", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d", rec.Code)
	}
}

func TestDocumentLifecycleEndpoints(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)

	_, resp := doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "编程语言", "directory_id": d.ID, "author_id": a.ID, "tags": []string{"Go"},
	})
	var doc model.Document
	decodeData(t, resp, &doc)

	// 发布
	rec, _ := doRequest(s, http.MethodPost, "/api/documents/"+doc.ID+"/publish", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("publish status = %d", rec.Code)
	}
	// 版本列表
	rec, _ = doRequest(s, http.MethodGet, "/api/documents/"+doc.ID+"/versions", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("versions status = %d", rec.Code)
	}
	// 搜索
	rec, _ = doRequest(s, http.MethodGet, "/api/search/documents?q=编程", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("search status = %d", rec.Code)
	}
	// 归档
	rec, _ = doRequest(s, http.MethodPost, "/api/documents/"+doc.ID+"/archive", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("archive status = %d", rec.Code)
	}
}

func TestDocumentRollbackEndpoint(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	_, resp := doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "v1", "directory_id": d.ID, "author_id": a.ID,
	})
	var doc model.Document
	decodeData(t, resp, &doc)
	doRequest(s, http.MethodPut, "/api/documents/"+doc.ID, map[string]interface{}{"content": "v2"})

	_, resp = doRequest(s, http.MethodGet, "/api/documents/"+doc.ID+"/versions", nil)
	var versions []model.Version
	decodeData(t, resp, &versions)

	rec, _ := doRequest(s, http.MethodPost, "/api/documents/"+doc.ID+"/rollback", map[string]string{"version_id": versions[0].ID})
	if rec.Code != http.StatusOK {
		t.Fatalf("rollback status = %d", rec.Code)
	}
}

func TestCommentAndFavoriteEndpoints(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	_, resp := doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "内容", "directory_id": d.ID, "author_id": a.ID,
	})
	var doc model.Document
	decodeData(t, resp, &doc)

	rec, _ := doRequest(s, http.MethodPost, "/api/comments", map[string]string{"document_id": doc.ID, "author_id": a.ID, "content": "好文章"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("comment status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodPost, "/api/favorites", map[string]string{"document_id": doc.ID, "user_id": "u1"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("favorite status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodDelete, "/api/favorites?user_id=u1&document_id="+doc.ID, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("unfavorite status = %d", rec.Code)
	}
}

func TestStatsAndExportEndpoints(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "内容", "directory_id": d.ID, "author_id": a.ID, "tags": []string{"Go"},
	})
	for _, path := range []string{
		"/api/stats/documents", "/api/stats/tags", "/api/stats/authors", "/api/stats/popular",
		"/api/export/documents",
	} {
		rec, _ := doRequest(s, http.MethodGet, path, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, rec.Code)
		}
	}
	// 导入
	rec, _ := doRequest(s, http.MethodPost, "/api/import/documents", map[string]interface{}{
		"documents": []map[string]interface{}{
			{"title": "Python 入门", "content": "x", "directory_name": "技术文档", "author_name": "张三"},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("import status = %d", rec.Code)
	}
}

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{MaxPageSize: 100, AuthToken: "secret"}
	s := newTestServer(cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token status = %d, want 401", rec.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid token status = %d, want 200", rec.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := &config.Config{MaxPageSize: 100, RateLimit: 2}
	s := newTestServer(cfg)
	h := s.Routes()
	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	req.RemoteAddr = "10.0.0.1:1000"
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d", i, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("third request status = %d, want 429", rec.Code)
	}
}

func TestMalformedJSONAndNotFound(t *testing.T) {
	s := newTestServer(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/documents", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed status = %d", rec.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/no-such", nil)
	rec = httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d", rec.Code)
	}
}

func TestDirectoryEndpoints(t *testing.T) {
	s := newTestServer(nil)
	_, resp := doRequest(s, http.MethodPost, "/api/directories", map[string]string{"name": "根目录"})
	var d model.Directory
	decodeData(t, resp, &d)
	// 子目录
	_, resp = doRequest(s, http.MethodPost, "/api/directories", map[string]string{"name": "子目录", "parent_id": d.ID})
	decodeData(t, resp, &d)
	// 目录树
	rec, _ := doRequest(s, http.MethodGet, "/api/directories/tree", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("tree status = %d", rec.Code)
	}
	// 重复 -> 409
	rec, _ = doRequest(s, http.MethodPost, "/api/directories", map[string]string{"name": "根目录"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d", rec.Code)
	}
	// 有子目录不能删除
	rec, _ = doRequest(s, http.MethodDelete, "/api/directories/"+d.ID, nil)
	_ = rec
}

func TestTagEndpoints(t *testing.T) {
	s := newTestServer(nil)
	_, resp := doRequest(s, http.MethodPost, "/api/tags", map[string]string{"name": "Go"})
	var tag model.Tag
	decodeData(t, resp, &tag)
	rec, _ := doRequest(s, http.MethodGet, "/api/tags", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list tags status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodGet, "/api/tags/"+tag.ID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get tag status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodDelete, "/api/tags/"+tag.ID, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete tag status = %d", rec.Code)
	}
}

func TestBatchEndpoints(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	var ids []string
	for i := 0; i < 2; i++ {
		_, resp := doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
			"title": "文档" + string(rune('A'+i)), "content": "内容", "directory_id": d.ID, "author_id": a.ID,
		})
		var doc model.Document
		decodeData(t, resp, &doc)
		doRequest(s, http.MethodPost, "/api/documents/"+doc.ID+"/publish", nil)
		ids = append(ids, doc.ID)
	}
	rec, _ := doRequest(s, http.MethodPost, "/api/documents/batch-archive", map[string]interface{}{"ids": ids})
	if rec.Code != http.StatusOK {
		t.Fatalf("batch archive status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodPost, "/api/documents/batch-delete", map[string]interface{}{"ids": ids})
	if rec.Code != http.StatusOK {
		t.Fatalf("batch delete status = %d", rec.Code)
	}
}

func TestOverviewAndMonthlyEndpoints(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "内容", "directory_id": d.ID, "author_id": a.ID,
	})
	for _, path := range []string{"/api/stats/overview", "/api/stats/monthly"} {
		rec, _ := doRequest(s, http.MethodGet, path, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, rec.Code)
		}
	}
}

func TestListDocumentsFilters(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "内容", "directory_id": d.ID, "author_id": a.ID, "tags": []string{"Go"},
	})
	rec, _ := doRequest(s, http.MethodGet, "/api/documents?tag=Go&status=draft", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("filter status = %d", rec.Code)
	}
}

func TestAuthorUpdateAndDeleteEndpoints(t *testing.T) {
	s := newTestServer(nil)
	_, resp := doRequest(s, http.MethodPost, "/api/authors", map[string]string{"name": "张三", "email": "z@x.com"})
	var a model.Author
	decodeData(t, resp, &a)

	rec, _ := doRequest(s, http.MethodPut, "/api/authors/"+a.ID, map[string]string{"bio": "资深作者"})
	if rec.Code != http.StatusOK {
		t.Fatalf("update author status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodDelete, "/api/authors/"+a.ID, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete author status = %d", rec.Code)
	}
}

func TestDirectoryUpdateEndpoints(t *testing.T) {
	s := newTestServer(nil)
	_, resp := doRequest(s, http.MethodPost, "/api/directories", map[string]string{"name": "技术文档"})
	var d model.Directory
	decodeData(t, resp, &d)
	rec, _ := doRequest(s, http.MethodPut, "/api/directories/"+d.ID, map[string]string{"description": "技术类文档"})
	if rec.Code != http.StatusOK {
		t.Fatalf("update directory status = %d", rec.Code)
	}
}

func TestDocumentUpdateAndDeleteEndpoints(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	_, resp := doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "v1", "directory_id": d.ID, "author_id": a.ID,
	})
	var doc model.Document
	decodeData(t, resp, &doc)

	rec, _ := doRequest(s, http.MethodPut, "/api/documents/"+doc.ID, map[string]string{"content": "v2"})
	if rec.Code != http.StatusOK {
		t.Fatalf("update document status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodGet, "/api/documents/"+doc.ID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get document status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodDelete, "/api/documents/"+doc.ID, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete document status = %d", rec.Code)
	}
}

func TestCommentListAndGetEndpoints(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	_, resp := doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "内容", "directory_id": d.ID, "author_id": a.ID,
	})
	var doc model.Document
	decodeData(t, resp, &doc)
	_, resp = doRequest(s, http.MethodPost, "/api/comments", map[string]string{"document_id": doc.ID, "author_id": a.ID, "content": "好文章"})
	var c model.Comment
	decodeData(t, resp, &c)

	rec, _ := doRequest(s, http.MethodGet, "/api/comments?document_id="+doc.ID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list comments status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodGet, "/api/comments/"+c.ID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get comment status = %d", rec.Code)
	}
	rec, _ = doRequest(s, http.MethodDelete, "/api/comments/"+c.ID, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete comment status = %d", rec.Code)
	}
}

func TestFavoriteListEndpoint(t *testing.T) {
	s := newTestServer(nil)
	a, d := setupAuthorAndDir(t, s)
	_, resp := doRequest(s, http.MethodPost, "/api/documents", map[string]interface{}{
		"title": "Go 入门", "content": "内容", "directory_id": d.ID, "author_id": a.ID,
	})
	var doc model.Document
	decodeData(t, resp, &doc)
	doRequest(s, http.MethodPost, "/api/favorites", map[string]string{"document_id": doc.ID, "user_id": "u1"})

	rec, _ := doRequest(s, http.MethodGet, "/api/favorites?user_id=u1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list favorites status = %d", rec.Code)
	}
}
