package redact

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func obj(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func asJSON(t *testing.T, m map[string]any) string {
	t.Helper()
	b, err := json.Marshal(m)
	require.NoError(t, err)
	return string(b)
}

func TestSecretKeepsKeysWithoutValues(t *testing.T) {
	o := obj(t, `{"kind":"Secret","data":{"password":"aHVudGVyMg=="},"stringData":{"token":"abc"},"type":"Opaque"}`)
	New(Options{}).Object("", "Secret", o)
	assert.JSONEq(t, `{"kind":"Secret","data":{"password":""},"stringData":{"token":""},"type":"Opaque"}`, asJSON(t, o))
}

func TestSecretInOtherGroupIsNotTreatedAsCoreSecret(t *testing.T) {
	o := obj(t, `{"data":{"x":"y"}}`)
	New(Options{}).Object("example.com", "Secret", o)
	assert.JSONEq(t, `{"data":{"x":"y"}}`, asJSON(t, o))
}

func TestEnvValuesDropped(t *testing.T) {
	// A Deployment-shaped object: the pod spec sits under spec.template.
	o := obj(t, `{"spec":{"template":{"spec":{
		"initContainers":[{"name":"init","env":[{"name":"A","value":"init-secret"}]}],
		"containers":[{"name":"app","env":[
			{"name":"DATABASE_URL","value":"postgres://u:p@db"},
			{"name":"TOKEN","valueFrom":{"secretKeyRef":{"name":"api","key":"token"}}}
		]}],
		"ephemeralContainers":[{"name":"debug","env":[{"name":"B","value":"eph-secret"}]}]
	}}}}`)
	New(Options{}).Object("apps", "Deployment", o)
	assert.JSONEq(t, `{"spec":{"template":{"spec":{
		"initContainers":[{"name":"init","env":[{"name":"A"}]}],
		"containers":[{"name":"app","env":[
			{"name":"DATABASE_URL"},
			{"name":"TOKEN","valueFrom":{"secretKeyRef":{"name":"api","key":"token"}}}
		]}],
		"ephemeralContainers":[{"name":"debug","env":[{"name":"B"}]}]
	}}}}`, asJSON(t, o))
}

func TestKeepEnvValues(t *testing.T) {
	r := New(Options{KeepEnvValues: true})
	o := obj(t, `{"spec":{"containers":[{"name":"app","env":[{"name":"A","value":"kept"}]}]}}`)
	r.Object("", "Pod", o)
	assert.JSONEq(t, `{"spec":{"containers":[{"name":"app","env":[{"name":"A","value":"kept"}]}]}}`, asJSON(t, o))
	assert.Equal(t, []string{SecretData, LastApplied}, r.Rules())
	assert.Equal(t, []string{SecretData, EnvValues, LastApplied}, New(Options{}).Rules())
}

func TestLastAppliedIsRedactedTheSameWay(t *testing.T) {
	applied := `{"apiVersion":"v1","kind":"Secret","data":{"password":"aHVudGVyMg=="},"metadata":{"name":"db"},"replicas":12345678901234567}`
	o := map[string]any{
		"kind": "Secret",
		"data": map[string]any{"password": "aHVudGVyMg=="},
		"metadata": map[string]any{"annotations": map[string]any{
			"kubectl.kubernetes.io/last-applied-configuration": applied + "\n",
			"other": "kept",
		}},
	}
	New(Options{}).Object("", "Secret", o)
	ann := o["metadata"].(map[string]any)["annotations"].(map[string]any)
	assert.Equal(t, "kept", ann["other"])
	assert.JSONEq(t,
		`{"apiVersion":"v1","kind":"Secret","data":{"password":""},"metadata":{"name":"db"},"replicas":12345678901234567}`,
		ann["kubectl.kubernetes.io/last-applied-configuration"].(string))
	assert.Contains(t, ann["kubectl.kubernetes.io/last-applied-configuration"], "12345678901234567", "numbers are not rounded")
}

func TestLastAppliedDroppedWhenUnreadable(t *testing.T) {
	for name, v := range map[string]any{"not json": "{password: hunter2", "not a string": 5, "null": "null"} {
		t.Run(name, func(t *testing.T) {
			o := map[string]any{"metadata": map[string]any{"annotations": map[string]any{
				"kubectl.kubernetes.io/last-applied-configuration": v,
			}}}
			New(Options{}).Object("", "Pod", o)
			assert.Empty(t, o["metadata"].(map[string]any)["annotations"])
		})
	}
}
