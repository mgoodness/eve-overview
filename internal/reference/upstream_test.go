package reference

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"testing"
)

func TestNewestZSPack(t *testing.T) {
	cases := []struct {
		name  string
		names []string
		want  string
	}{
		{"empty", nil, ""},
		{"only legacy", []string{"zs_full_v9.00.0347.yaml"}, "zs_full_v9.00.0347.yaml"},
		{"newest patch", []string{"zs_full_v10.06.09.yaml", "zs_full_v10.07.00.yaml", "zs_full_v9.00.0347.yaml"}, "zs_full_v10.07.00.yaml"},
		{"newest minor", []string{"zs_full_v10.06.09.yaml", "zs_full_v10.10.00.yaml"}, "zs_full_v10.10.00.yaml"},
		{"newest major", []string{"zs_full_v9.99.99.yaml", "zs_full_v10.00.00.yaml"}, "zs_full_v10.00.00.yaml"},
		{"ignores non-packs", []string{"fenris_default_v24.01.yaml", "zs_full_v10.06.09.yaml"}, "zs_full_v10.06.09.yaml"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := newestZSPack(tc.names); got != tc.want {
				t.Errorf("newestZSPack(%v) = %q, want %q", tc.names, got, tc.want)
			}
		})
	}
}

func TestCheckSDEUpstream(t *testing.T) {
	manifest := func(build int) []byte {
		return []byte(fmt.Sprintf(`{"_key":"sde","buildNumber":%d,"releaseDate":"2026-10-12T11:08:38Z"}`, build))
	}

	t.Run("current build", func(t *testing.T) {
		transport := &stubTransport{bodies: map[string][]byte{sdeLatestURL: manifest(123)}}
		status, err := checkSDEUpstream(context.Background(), &http.Client{Transport: transport}, "123")
		if err != nil {
			t.Fatalf("checkSDEUpstream: %v", err)
		}
		if !status.Current {
			t.Errorf("Current = false for a matching build; status = %+v", status)
		}
		if status.Pinned != "123" || status.Latest != "123" {
			t.Errorf("pinned/latest = %q/%q, want 123/123", status.Pinned, status.Latest)
		}
	})

	t.Run("new build available", func(t *testing.T) {
		transport := &stubTransport{bodies: map[string][]byte{sdeLatestURL: manifest(124)}}
		status, err := checkSDEUpstream(context.Background(), &http.Client{Transport: transport}, "123")
		if err != nil {
			t.Fatalf("checkSDEUpstream: %v", err)
		}
		if status.Current {
			t.Errorf("Current = true for a newer build; status = %+v", status)
		}
		if status.Latest != "124" {
			t.Errorf("Latest = %q, want 124", status.Latest)
		}
	})
}

func TestCheckZSPackUpstream(t *testing.T) {
	pinnedName := path.Base(ReferencePackURL)
	rawURL := zsDefaultsRawBase + "/" + pinnedName
	listing := func(names ...string) []byte {
		body := []byte("[")
		for i, name := range names {
			if i > 0 {
				body = append(body, ',')
			}
			body = append(body, fmt.Sprintf(`{"name":%q}`, name)...)
		}
		return append(body, ']')
	}
	pack := []byte("presets: []\n")

	t.Run("pinned pack is current", func(t *testing.T) {
		transport := &stubTransport{bodies: map[string][]byte{
			zsDefaultsAPI: listing(pinnedName, "zs_full_v9.00.0347.yaml"),
			rawURL:        pack,
		}}
		status, err := checkZSPackUpstream(context.Background(), &http.Client{Transport: transport}, "", sha256Hex(pack))
		if err != nil {
			t.Fatalf("checkZSPackUpstream: %v", err)
		}
		if !status.Current {
			t.Errorf("Current = false for the pinned pack; status = %+v", status)
		}
	})

	t.Run("newer pack is bundled", func(t *testing.T) {
		transport := &stubTransport{bodies: map[string][]byte{
			zsDefaultsAPI: listing(pinnedName, "zs_full_v10.07.00.yaml"),
		}}
		status, err := checkZSPackUpstream(context.Background(), &http.Client{Transport: transport}, "", "irrelevant")
		if err != nil {
			t.Fatalf("checkZSPackUpstream: %v", err)
		}
		if status.Current {
			t.Errorf("Current = true when a newer pack exists; status = %+v", status)
		}
		if status.Latest != "zs_full_v10.07.00.yaml" {
			t.Errorf("Latest = %q, want the newer pack", status.Latest)
		}
	})

	t.Run("pinned pack changed in place", func(t *testing.T) {
		transport := &stubTransport{bodies: map[string][]byte{
			zsDefaultsAPI: listing(pinnedName),
			rawURL:        pack,
		}}
		status, err := checkZSPackUpstream(context.Background(), &http.Client{Transport: transport}, "", sha256Hex([]byte("a different pack")))
		if err != nil {
			t.Fatalf("checkZSPackUpstream: %v", err)
		}
		if status.Current {
			t.Errorf("Current = true when the pinned filename's contents changed; status = %+v", status)
		}
	})
}
