package auth

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"zion-english/internal/conf"
)

func TestCSRFMiddlewareRejectsMissingToken(t *testing.T) {
	cfg := &conf.Config{Secret: "test-secret", BasePath: "/admin"}
	called := false
	handler := CSRFMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/students", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if called {
		t.Fatal("handler should not run without CSRF token")
	}
}

func TestCSRFMiddlewareAcceptsMatchingToken(t *testing.T) {
	cfg := &conf.Config{Secret: "test-secret", BasePath: "/admin"}
	called := false
	handler := CSRFMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/students", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cookie := rec.Result().Cookies()[0]
	req2 := httptest.NewRequest(http.MethodPost, "/admin/students", nil)
	req2.AddCookie(cookie)
	req2.Header.Set(csrfHeaderName, cookie.Value)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec2.Code, http.StatusOK)
	}
	if !called {
		t.Fatal("handler should run with valid CSRF token")
	}
}

func TestCSRFMiddlewareAcceptsMultipartFormToken(t *testing.T) {
	cfg := &conf.Config{Secret: "test-secret", BasePath: "/admin"}
	called := false
	handler := CSRFMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/profile/document", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	cookie := rec.Result().Cookies()[0]

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField(csrfFormField, cookie.Value); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("document", "id.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(part, "%PDF-1.4"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/admin/profile/document", body)
	req2.Header.Set("Content-Type", writer.FormDataContentType())
	req2.AddCookie(cookie)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec2.Code, http.StatusOK)
	}
	if !called {
		t.Fatal("handler should run with valid multipart CSRF token")
	}
}
