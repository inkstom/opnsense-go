package captiveportal

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"testing"
)

func testTemplateZip(t *testing.T, body string) string {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("index.html")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestTemplate(t *testing.T) {
	c := newTestController(t)
	ctx := context.Background()

	id, err := c.SaveTemplate(ctx, "", "testtpl", testTemplateZip(t, "<html><body>one</body></html>"))
	if err != nil {
		t.Fatalf("SaveTemplate: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteTemplate(ctx, id) })

	tpl, err := c.GetTemplate(ctx, id)
	if err != nil || tpl == nil || tpl.Name != "testtpl" || tpl.FileId == "" || tpl.Content == "" {
		t.Fatalf("GetTemplate = %+v, %v", tpl, err)
	}
	if _, err := c.SaveTemplate(ctx, id, "testtpl2", testTemplateZip(t, "<html><body>two</body></html>")); err != nil {
		t.Fatalf("SaveTemplate update: %v", err)
	}
	if tpl, err = c.GetTemplate(ctx, id); err != nil || tpl == nil || tpl.Name != "testtpl2" {
		t.Fatalf("GetTemplate after update = %+v, %v", tpl, err)
	}
	if err := c.DeleteTemplate(ctx, id); err != nil {
		t.Fatalf("DeleteTemplate: %v", err)
	}
	if tpl, err = c.GetTemplate(ctx, id); err != nil || tpl != nil {
		t.Fatalf("GetTemplate after delete = %+v, %v", tpl, err)
	}
}
