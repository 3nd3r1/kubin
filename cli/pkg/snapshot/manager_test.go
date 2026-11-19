package snapshot

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
	"github.com/3nd3r1/kubin/cli/pkg/kube/kubetest"
	"github.com/3nd3r1/kubin/cli/pkg/redact"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// readArchive returns the entry names in order and their contents.
func readArchive(t *testing.T, data []byte) ([]string, map[string][]byte) {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	tr := tar.NewReader(gz)
	var names []string
	files := map[string][]byte{}
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		assert.Equal(t, byte(tar.TypeReg), hdr.Typeflag, "%s is a regular file", hdr.Name)
		b, err := io.ReadAll(tr)
		require.NoError(t, err)
		names = append(names, hdr.Name)
		files[hdr.Name] = b
	}
	return names, files
}

func TestCreateWritesLayout(t *testing.T) {
	managed := []metav1.ManagedFieldsEntry{{Manager: "kubectl", Operation: metav1.ManagedFieldsOperationApply}}
	cs := kubetest.New(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "prod"}},
		// Same name in two namespaces: both must be kept.
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web", ManagedFields: managed}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "web", ManagedFields: managed}},
		&corev1.Event{ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "web.17f3a"}, Reason: "BackOff"},
	)

	mgr := NewManager(cs, "v0.0.0-test", Options{Logs: collector.DefaultLogOptions()})
	created := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	mgr.now = func() time.Time { return created }

	var buf bytes.Buffer
	_, err := mgr.Create(context.Background(), &buf)
	require.NoError(t, err)

	names, files := readArchive(t, buf.Bytes())
	require.NotEmpty(t, names)
	assert.Equal(t, "manifest.json", names[0], "manifest is the first entry")
	assert.True(t, sort.StringsAreSorted(names[1:]), "entries after the manifest are sorted: %v", names)
	assert.Equal(t, []string{
		"manifest.json",
		"resources/core/v1/Event/prod/web.17f3a.json",
		"resources/core/v1/Namespace/_cluster/default.json",
		"resources/core/v1/Namespace/_cluster/prod.json",
		"resources/core/v1/Pod/default/web.json",
		"resources/core/v1/Pod/prod/web.json",
	}, names)

	var m Manifest
	require.NoError(t, json.Unmarshal(files["manifest.json"], &m))
	assert.Equal(t, Manifest{
		FormatVersion: 3,
		CreatedAt:     created,
		CLIVersion:    "v0.0.0-test",
		Cluster:       Cluster{Context: "kind-kubin", Name: "kind-kubin", ServerVersion: "v1.33.1"},
		Resources: []KindCount{
			{Group: "core", Version: "v1", Kind: "Event", Count: 1},
			{Group: "core", Version: "v1", Kind: "Namespace", Count: 2},
			{Group: "core", Version: "v1", Kind: "Pod", Count: 2},
		},
		Logs:       Logs{TailLines: 1000, MaxBytes: 1 << 20, Files: []LogFile{}},
		Redactions: []string{"secret-data", "env-values", "last-applied-configuration"},
	}, m)

	var pod map[string]any
	require.NoError(t, json.Unmarshal(files["resources/core/v1/Pod/prod/web.json"], &pod))
	assert.Equal(t, "v1", pod["apiVersion"])
	assert.Equal(t, "Pod", pod["kind"])
	meta := pod["metadata"].(map[string]any)
	assert.Equal(t, "prod", meta["namespace"])
	assert.NotContains(t, meta, "managedFields")
}

func TestCreateWritesLogs(t *testing.T) {
	running := corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
	crashed := corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1}}
	cs := kubetest.New(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "web"},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{Name: "migrate"}},
			Containers:     []corev1.Container{{Name: "app"}, {Name: "sidecar"}},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{{Name: "migrate", State: crashed}},
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "app", State: running, LastTerminationState: crashed, RestartCount: 3},
				// Never started: no log to fetch.
				{Name: "sidecar", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"}}},
			},
		},
	})

	mgr := NewManager(cs, "v0.0.0-test", Options{Logs: collector.DefaultLogOptions()})
	var buf bytes.Buffer
	m, err := mgr.Create(context.Background(), &buf)
	require.NoError(t, err)

	names, files := readArchive(t, buf.Bytes())
	assert.True(t, sort.StringsAreSorted(names[1:]), "entries after the manifest are sorted: %v", names)
	assert.Subset(t, names, []string{
		"logs/prod/web/app.log",
		"logs/prod/web/app.previous.log",
		"logs/prod/web/migrate.log",
	})
	assert.NotContains(t, names, "logs/prod/web/sidecar.log")
	// client-go's fake answers every log request with this body.
	assert.Equal(t, "fake logs", string(files["logs/prod/web/app.log"]))

	assert.Equal(t, []LogFile{
		{Namespace: "prod", Pod: "web", Container: "app"},
		{Namespace: "prod", Pod: "web", Container: "app", Previous: true},
		{Namespace: "prod", Pod: "web", Container: "migrate"},
	}, m.Logs.Files)

	var written Manifest
	require.NoError(t, json.Unmarshal(files["manifest.json"], &written))
	assert.Equal(t, m.Logs, written.Logs)

	t.Run("skip logs", func(t *testing.T) {
		mgr := NewManager(cs, "v0.0.0-test", Options{Logs: collector.DefaultLogOptions(), SkipLogs: true})
		var buf bytes.Buffer
		m, err := mgr.Create(context.Background(), &buf)
		require.NoError(t, err)
		names, _ := readArchive(t, buf.Bytes())
		for _, n := range names {
			assert.NotContains(t, n, "logs/")
		}
		assert.Contains(t, names, "resources/core/v1/Pod/prod/web.json")
		assert.Empty(t, m.Logs.Files)
	})
}

// staticCollector returns fixed resources, for kinds the CLI doesn't
// collect yet.
type staticCollector []collector.Resource

func (staticCollector) Name() string { return "static" }

func (c staticCollector) Collect(context.Context) ([]collector.Resource, error) { return c, nil }

func typed(t *testing.T, group, version, kind string, obj any) collector.Resource {
	t.Helper()
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	require.NoError(t, err)
	m["apiVersion"], m["kind"] = version, kind
	if group != "" {
		m["apiVersion"] = group + "/" + version
	}
	meta := m["metadata"].(map[string]any)
	ns, _ := meta["namespace"].(string)
	return collector.Resource{Group: group, Version: version, Kind: kind, Namespace: ns, Name: meta["name"].(string), Object: m}
}

func TestCreateKeepsSecretsOut(t *testing.T) {
	const (
		secretValue = "hunter2-secret-value"
		envValue    = "postgres://admin:env-secret@db"
	)
	lastApplied := `{"apiVersion":"v1","kind":"Secret","metadata":{"name":"db","namespace":"prod"},"stringData":{"password":"` + secretValue + `"}}`
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "db", Annotations: map[string]string{
			"kubectl.kubernetes.io/last-applied-configuration": lastApplied,
		}},
		Data: map[string][]byte{"password": []byte(secretValue)},
	}
	env := []corev1.EnvVar{
		{Name: "DATABASE_URL", Value: envValue},
		{Name: "PASSWORD", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "db"}, Key: "password",
		}}},
	}
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "api"},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "api", Env: env}},
		}}},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "api-1"},
		Spec:       corev1.PodSpec{InitContainers: []corev1.Container{{Name: "migrate", Env: env}}},
	}

	create := func(t *testing.T, opts redact.Options) []byte {
		cs := kubetest.New(pod, deploy)
		mgr := NewManager(cs, "v0.0.0-test",
			Options{Logs: collector.DefaultLogOptions(), Redact: opts})
		mgr.collectors = append(mgr.collectors, staticCollector{typed(t, "", "v1", "Secret", secret)})
		var buf bytes.Buffer
		_, err := mgr.Create(context.Background(), &buf)
		require.NoError(t, err)

		names, files := readArchive(t, buf.Bytes())
		require.Contains(t, names, "resources/core/v1/Secret/prod/db.json")
		require.Contains(t, names, "resources/apps/v1/Deployment/prod/api.json")
		var all []byte
		for _, b := range files {
			all = append(all, b...)
		}
		return all
	}

	t.Run("default", func(t *testing.T) {
		all := string(create(t, redact.Options{}))
		for _, s := range []string{secretValue, base64.StdEncoding.EncodeToString([]byte(secretValue)), envValue} {
			assert.NotContains(t, all, s)
		}
		assert.Contains(t, all, `"secretKeyRef"`, "valueFrom references stay")
		assert.Contains(t, all, `"DATABASE_URL"`, "env names stay")
		assert.Contains(t, all, `"password": ""`, "Secret keys stay")
		assert.Contains(t, all, `"env-values"`, "the manifest lists the rule")
	})

	t.Run("keep env values", func(t *testing.T) {
		all := string(create(t, redact.Options{KeepEnvValues: true}))
		assert.Contains(t, all, envValue)
		assert.NotContains(t, all, secretValue)
		assert.NotContains(t, all, base64.StdEncoding.EncodeToString([]byte(secretValue)))
		assert.NotContains(t, all, `"env-values"`)
	})
}
