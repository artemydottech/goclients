package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/artemydottech/goclients/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()

	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "e2e.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := repository.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	server := httptest.NewServer(New(db))
	t.Cleanup(server.Close)

	return server
}

type client struct {
	t    *testing.T
	base string
}

func (c client) do(method, path string, body any) (int, []byte) {
	c.t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		c.t.Fatalf("request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)

	return resp.StatusCode, buf.Bytes()
}

func (c client) create(path string, body any) int {
	c.t.Helper()

	status, raw := c.do(http.MethodPost, path, body)
	if status != http.StatusCreated {
		c.t.Fatalf("POST %s = %d: %s", path, status, raw)
	}

	var created struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(raw, &created); err != nil {
		c.t.Fatalf("decode id: %v", err)
	}

	return created.ID
}

func (c client) expect(method, path string, body any, want int) []byte {
	c.t.Helper()

	status, raw := c.do(method, path, body)
	if status != want {
		c.t.Fatalf("%s %s = %d, want %d: %s", method, path, status, want, raw)
	}

	return raw
}

func newClient(t *testing.T) client {
	t.Helper()

	return client{t: t, base: newServer(t).URL}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
