package reference

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// stubTransport serves canned bodies per URL and counts requests, so the
// fetch-once-cache-forever property is observable without the network.
type stubTransport struct {
	bodies   map[string][]byte
	requests []string
}

func (t *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.requests = append(t.requests, req.URL.String())
	body, ok := t.bodies[req.URL.String()]
	if !ok {
		return nil, fmt.Errorf("unexpected request to %s", req.URL)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{},
	}, nil
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// testFetcher points a Fetcher at the stub transport and canned bytes,
// standing in for the pinned URLs and checksums.
func testFetcher(t *testing.T, transport *stubTransport, cacheDir string) *Fetcher {
	t.Helper()
	cacheDir = filepath.Join(cacheDir, "cache")
	fetcher := newWithClient(cacheDir, &http.Client{Transport: transport})
	fetcher.referencePackURL = "https://example.test/reference-pack.yaml"
	fetcher.sdeURL = "https://example.test/sde.zip"
	return fetcher
}

func TestReferencePackIsFetchedOnceThenCached(t *testing.T) {
	pack := []byte("presets: []\n")
	transport := &stubTransport{bodies: map[string][]byte{
		"https://example.test/reference-pack.yaml": pack,
	}}
	fetcher := testFetcher(t, transport, t.TempDir())
	fetcher.referencePackSHA = sha256Hex(pack)

	first, err := fetcher.ReferencePack(context.Background())
	if err != nil {
		t.Fatalf("first ReferencePack: %v", err)
	}
	if !bytes.Equal(first, pack) {
		t.Fatalf("ReferencePack = %q, want %q", first, pack)
	}

	second, err := fetcher.ReferencePack(context.Background())
	if err != nil {
		t.Fatalf("second ReferencePack: %v", err)
	}
	if !bytes.Equal(second, pack) {
		t.Fatalf("second ReferencePack = %q, want %q", second, pack)
	}
	if got := len(transport.requests); got != 1 {
		t.Errorf("made %d network requests, want 1 (second call must read the cache)", got)
	}
	if _, err := os.Stat(fetcher.referencePackCachePath()); err != nil {
		t.Errorf("reference pack was not cached: %v", err)
	}
}

func TestReferencePackChecksumMismatchFailsAndDoesNotCache(t *testing.T) {
	pack := []byte("presets: []\n")
	transport := &stubTransport{bodies: map[string][]byte{
		"https://example.test/reference-pack.yaml": pack,
	}}
	fetcher := testFetcher(t, transport, t.TempDir())
	fetcher.referencePackSHA = sha256Hex([]byte("something else"))

	if _, err := fetcher.ReferencePack(context.Background()); err == nil {
		t.Fatal("ReferencePack succeeded despite a checksum mismatch")
	}
	if _, err := os.Stat(fetcher.referencePackCachePath()); !os.IsNotExist(err) {
		t.Errorf("a checksum mismatch must not populate the cache (stat err = %v)", err)
	}
}

func TestSDEGroupsExtractsGroupsAndCaches(t *testing.T) {
	groups := []byte(`{"_key":6,"name":{"en":"Sun"}}` + "\n")
	archive := zipWith(t, map[string][]byte{
		"sde/fsd/categories.jsonl": []byte(`{"_key":2}`),
		"sde/fsd/groups.jsonl":     groups,
	})
	transport := &stubTransport{bodies: map[string][]byte{
		"https://example.test/sde.zip": archive,
	}}
	fetcher := testFetcher(t, transport, t.TempDir())
	fetcher.sdeSHA = sha256Hex(archive)

	first, err := fetcher.SDEGroups(context.Background())
	if err != nil {
		t.Fatalf("first SDEGroups: %v", err)
	}
	if !bytes.Equal(first, groups) {
		t.Fatalf("SDEGroups = %q, want %q", first, groups)
	}
	cached, err := os.ReadFile(fetcher.sdeGroupsCachePath())
	if err != nil {
		t.Fatalf("SDE Group matrix was not cached: %v", err)
	}
	if !bytes.Equal(cached, groups) {
		t.Errorf("cached bytes = %q, want the extracted groups.jsonl", cached)
	}

	second, err := fetcher.SDEGroups(context.Background())
	if err != nil {
		t.Fatalf("second SDEGroups: %v", err)
	}
	if !bytes.Equal(second, groups) {
		t.Fatalf("second SDEGroups = %q, want %q", second, groups)
	}
	if got := len(transport.requests); got != 1 {
		t.Errorf("made %d network requests, want 1 (second call must read the cache)", got)
	}
}

func TestSDEGroupsChecksumMismatchFailsAndDoesNotCache(t *testing.T) {
	groups := []byte(`{"_key":6,"name":{"en":"Sun"}}` + "\n")
	archive := zipWith(t, map[string][]byte{"sde/fsd/groups.jsonl": groups})
	transport := &stubTransport{bodies: map[string][]byte{
		"https://example.test/sde.zip": archive,
	}}
	fetcher := testFetcher(t, transport, t.TempDir())
	fetcher.sdeSHA = sha256Hex([]byte("something else"))

	if _, err := fetcher.SDEGroups(context.Background()); err == nil {
		t.Fatal("SDEGroups succeeded despite a checksum mismatch")
	}
	if _, err := os.Stat(fetcher.sdeGroupsCachePath()); !os.IsNotExist(err) {
		t.Errorf("a checksum mismatch must not populate the cache (stat err = %v)", err)
	}
}

func TestSDEGroupsWithoutGroupsFileFails(t *testing.T) {
	archive := zipWith(t, map[string][]byte{"sde/fsd/categories.jsonl": []byte("{}")})
	transport := &stubTransport{bodies: map[string][]byte{
		"https://example.test/sde.zip": archive,
	}}
	fetcher := testFetcher(t, transport, t.TempDir())
	fetcher.sdeSHA = sha256Hex(archive)

	if _, err := fetcher.SDEGroups(context.Background()); err == nil {
		t.Fatal("SDEGroups succeeded for an archive with no groups.jsonl")
	}
}

func zipWith(t *testing.T, entries map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, body := range entries {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create %s in archive: %v", name, err)
		}
		if _, err := file.Write(body); err != nil {
			t.Fatalf("write %s in archive: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}
	return buf.Bytes()
}
