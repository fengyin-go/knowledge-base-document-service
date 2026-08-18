package model

import (
	"testing"
)

func TestAuthorValidate(t *testing.T) {
	a := &Author{Name: "张三", Email: "z@x.com"}
	if err := a.Validate(); err != nil {
		t.Fatalf("valid: %v", err)
	}
	a.Name = ""
	if err := a.Validate(); err == nil {
		t.Fatalf("expect error for empty name")
	}
}

func TestAuthorFilter(t *testing.T) {
	a := &Author{Name: "张三", Email: "z@x.com"}
	if !(AuthorFilter{Keyword: "张"}).Match(a) {
		t.Fatalf("expect match by keyword")
	}
	if (AuthorFilter{Keyword: "李"}).Match(a) {
		t.Fatalf("expect no match")
	}
}

func TestDirectoryValidate(t *testing.T) {
	d := &Directory{Name: "技术文档"}
	if err := d.Validate(); err != nil {
		t.Fatalf("valid: %v", err)
	}
	d.ParentID = "self"
	d.ID = "self"
	if err := d.Validate(); err == nil {
		t.Fatalf("expect error for self parent")
	}
}

func TestDirectoryFilter(t *testing.T) {
	d := &Directory{Name: "技术文档", ParentID: "root"}
	if !(DirectoryFilter{ParentID: "root"}).Match(d) {
		t.Fatalf("expect match by parent")
	}
	if (DirectoryFilter{ParentID: "other"}).Match(d) {
		t.Fatalf("expect no match by parent")
	}
}

func TestDocumentValidate(t *testing.T) {
	d := &Document{Title: "Go 入门", DirectoryID: "d1", AuthorID: "a1"}
	if err := d.Validate(); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if d.Status != DocumentDraft {
		t.Fatalf("default status = %s", d.Status)
	}
	d.Title = ""
	if err := d.Validate(); err == nil {
		t.Fatalf("expect error for empty title")
	}
}

func TestDocumentTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{DocumentDraft, DocumentPublished, true},
		{DocumentDraft, DocumentArchived, true},
		{DocumentPublished, DocumentArchived, true},
		{DocumentArchived, DocumentPublished, true},
		{DocumentPublished, DocumentDraft, false},
		{DocumentArchived, DocumentDraft, false},
	}
	for _, c := range cases {
		if got := CanTransitionDocument(c.from, c.to); got != c.want {
			t.Fatalf("transition %s->%s = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestDocumentFilter(t *testing.T) {
	d := &Document{Title: "Go 入门", Content: "编程语言", DirectoryID: "d1", AuthorID: "a1", Tags: []string{"Go"}, Status: DocumentPublished}
	if !(DocumentFilter{Status: DocumentPublished}).Match(d) {
		t.Fatalf("expect match by status")
	}
	if !(DocumentFilter{Tag: "Go"}).Match(d) {
		t.Fatalf("expect match by tag")
	}
	if (DocumentFilter{Tag: "Java"}).Match(d) {
		t.Fatalf("expect no match by tag")
	}
	if !(DocumentFilter{Keyword: "编程"}).Match(d) {
		t.Fatalf("expect match by keyword")
	}
}

func TestDocumentHasTag(t *testing.T) {
	d := &Document{Tags: []string{"Go", "K8s"}}
	if !d.HasTag("Go") {
		t.Fatalf("expect has tag Go")
	}
	if d.HasTag("Java") {
		t.Fatalf("expect no tag Java")
	}
}

func TestVersionValidate(t *testing.T) {
	v := &Version{DocumentID: "d1", VersionNo: 1, Content: "内容"}
	if err := v.Validate(); err != nil {
		t.Fatalf("valid: %v", err)
	}
	v.VersionNo = 0
	if err := v.Validate(); err == nil {
		t.Fatalf("expect error for version no 0")
	}
}

func TestTagValidate(t *testing.T) {
	tag := &Tag{Name: "Go"}
	if err := tag.Validate(); err != nil {
		t.Fatalf("valid: %v", err)
	}
	tag.Name = ""
	if err := tag.Validate(); err == nil {
		t.Fatalf("expect error for empty name")
	}
}

func TestCommentValidate(t *testing.T) {
	c := &Comment{DocumentID: "d1", AuthorID: "a1", Content: "好文章"}
	if err := c.Validate(); err != nil {
		t.Fatalf("valid: %v", err)
	}
	c.Content = ""
	if err := c.Validate(); err == nil {
		t.Fatalf("expect error for empty content")
	}
}

func TestCommentFilter(t *testing.T) {
	c := &Comment{DocumentID: "d1", AuthorID: "a1"}
	if !(CommentFilter{DocumentID: "d1"}).Match(c) {
		t.Fatalf("expect match by document")
	}
	if (CommentFilter{AuthorID: "a2"}).Match(c) {
		t.Fatalf("expect no match by author")
	}
}

func TestFavoriteValidate(t *testing.T) {
	f := &Favorite{DocumentID: "d1", UserID: "u1"}
	if err := f.Validate(); err != nil {
		t.Fatalf("valid: %v", err)
	}
	f.UserID = ""
	if err := f.Validate(); err == nil {
		t.Fatalf("expect error for empty user")
	}
}

func TestFavoriteFilter(t *testing.T) {
	f := &Favorite{DocumentID: "d1", UserID: "u1"}
	if !(FavoriteFilter{UserID: "u1"}).Match(f) {
		t.Fatalf("expect match by user")
	}
	if !(FavoriteFilter{DocumentID: "d1"}).Match(f) {
		t.Fatalf("expect match by document")
	}
	if (FavoriteFilter{DocumentID: "d2"}).Match(f) {
		t.Fatalf("expect no match")
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("field", "消息")
	if !IsValidationError(err) {
		t.Fatalf("expect validation error")
	}
	if err.Error() != "field: 消息" {
		t.Fatalf("error string = %s", err.Error())
	}
}

func TestDocumentStatusConstants(t *testing.T) {
	if DocumentDraft != "draft" || DocumentPublished != "published" || DocumentArchived != "archived" {
		t.Fatalf("document status constants mismatch")
	}
}

func TestDocumentFilterAuthorAndDirectory(t *testing.T) {
	d := &Document{Title: "Go", DirectoryID: "d1", AuthorID: "a1", Status: DocumentPublished}
	if !(DocumentFilter{DirectoryID: "d1"}).Match(d) {
		t.Fatalf("expect match by directory")
	}
	if !(DocumentFilter{AuthorID: "a1"}).Match(d) {
		t.Fatalf("expect match by author")
	}
	if (DocumentFilter{DirectoryID: "d2"}).Match(d) {
		t.Fatalf("expect no match by directory")
	}
	if (DocumentFilter{AuthorID: "a2"}).Match(d) {
		t.Fatalf("expect no match by author")
	}
}

func TestDirectoryValidateEmptyName(t *testing.T) {
	d := &Directory{Name: "  "}
	if err := d.Validate(); err == nil {
		t.Fatalf("expect error for empty name")
	}
}

func TestVersionValidateMissingDocument(t *testing.T) {
	v := &Version{VersionNo: 1}
	if err := v.Validate(); err == nil {
		t.Fatalf("expect error for empty document")
	}
}
