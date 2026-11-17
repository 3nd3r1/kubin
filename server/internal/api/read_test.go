package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/server/internal/snapshot/snapshottest"
)

// withFixture uploads snapshottest.Fixture and returns the server and the ID.
func withFixture(t *testing.T) (testServer, string) {
	t.Helper()
	s := newUploadServer(t, "", 1<<20)
	resp, body := s.upload(t, snapshottest.Fixture(t))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload: %d %s", resp.StatusCode, body)
	}
	var up uploadResponse
	if err := json.Unmarshal(body, &up); err != nil {
		t.Fatal(err)
	}
	return s, up.ID
}

func (s testServer) get(t *testing.T, path string) (int, []byte) {
	t.Helper()
	resp, err := http.Get(s.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("GET %s: Content-Type %q", path, ct)
	}
	return resp.StatusCode, b
}

func TestGetSnapshot(t *testing.T) {
	s, id := withFixture(t)
	status, body := s.get(t, "/api/v1/snapshots/"+id)
	if status != http.StatusOK {
		t.Fatalf("got %d %s", status, body)
	}
	var got snapshotResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != id || time.Since(got.UploadedAt) > time.Minute {
		t.Errorf("id %q uploaded_at %v", got.ID, got.UploadedAt)
	}
	var m struct {
		Cluster struct{ Context string } `json:"cluster"`
	}
	if err := json.Unmarshal(got.Manifest, &m); err != nil || m.Cluster.Context != "kind-kubin" {
		t.Errorf("manifest %s doesn't carry the cluster context", got.Manifest)
	}
	want := []kindCount{{"core", "v1", "Namespace", 2}, {"core", "v1", "Pod", 2}}
	if len(got.Counts) != len(want) || got.Counts[0] != want[0] || got.Counts[1] != want[1] {
		t.Errorf("counts %+v, want %+v", got.Counts, want)
	}
}

func TestListResources(t *testing.T) {
	s, id := withFixture(t)
	for _, tt := range []struct {
		query string
		want  string
	}{
		{"", "Namespace:/default Namespace:/kube-system Pod:default/web-0 Pod:kube-system/coredns-1"},
		{"?kind=Pod", "Pod:default/web-0 Pod:kube-system/coredns-1"},
		{"?kind=Pod&namespace=kube-system", "Pod:kube-system/coredns-1"},
		{"?namespace=_cluster", "Namespace:/default Namespace:/kube-system"},
		{"?kind=Deployment", ""},
	} {
		status, body := s.get(t, "/api/v1/snapshots/"+id+"/resources"+tt.query)
		if status != http.StatusOK {
			t.Fatalf("%s: got %d %s", tt.query, status, body)
		}
		var got resourceList
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatal(err)
		}
		if got.Items == nil {
			t.Errorf("%s: items is null, want an array", tt.query)
		}
		var names []string
		for _, it := range got.Items {
			names = append(names, it.Kind+":"+it.Namespace+"/"+it.Name)
			if !json.Valid(it.Object) || it.Version != "v1" || it.Group != "core" {
				t.Errorf("%s: bad item %+v", tt.query, it)
			}
		}
		if strings.Join(names, " ") != tt.want {
			t.Errorf("%s: got %v, want %s", tt.query, names, tt.want)
		}
	}
}

func TestGetResource(t *testing.T) {
	s, id := withFixture(t)
	status, body := s.get(t, "/api/v1/snapshots/"+id+"/resources/Pod/default/web-0")
	if status != http.StatusOK {
		t.Fatalf("got %d %s", status, body)
	}
	var pod struct {
		Kind     string
		Metadata struct{ Name, Namespace string }
		Status   struct{ Phase string }
	}
	if err := json.Unmarshal(body, &pod); err != nil {
		t.Fatal(err)
	}
	if pod.Kind != "Pod" || pod.Metadata.Name != "web-0" || pod.Status.Phase != "Running" {
		t.Errorf("got %s", body)
	}

	status, body = s.get(t, "/api/v1/snapshots/"+id+"/resources/Namespace/_cluster/kube-system")
	if status != http.StatusOK || !strings.Contains(string(body), `"kube-system"`) {
		t.Errorf("cluster-scoped: got %d %s", status, body)
	}
}

func TestReadNotFound(t *testing.T) {
	s, id := withFixture(t)
	for _, tt := range []struct{ path, msg string }{
		{"/api/v1/snapshots/aaaaaaaaaaaaaaaaaaaaaaaa", "snapshot not found"},
		{"/api/v1/snapshots/aaaaaaaaaaaaaaaaaaaaaaaa/resources", "snapshot not found"},
		{"/api/v1/snapshots/aaaaaaaaaaaaaaaaaaaaaaaa/resources/Pod/default/web-0", "snapshot not found"},
		{"/api/v1/snapshots/" + id + "/resources/Pod/default/missing", "resource not found"},
		{"/api/v1/snapshots/" + id + "/resources/Pod/_cluster/web-0", "resource not found"},
	} {
		status, body := s.get(t, tt.path)
		if status != http.StatusNotFound || !strings.Contains(string(body), tt.msg) {
			t.Errorf("GET %s: got %d %s, want 404 %q", tt.path, status, body, tt.msg)
		}
	}
}
