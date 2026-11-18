// Package redact removes credentials from objects before they go into a
// snapshot. A snapshot link is effectively public, so every collected object
// goes through a Redactor, whichever collector produced it.
package redact

import (
	"bytes"
	"encoding/json"
)

// Rule names, as recorded in the snapshot manifest.
const (
	// SecretData empties the values of a Secret's data and stringData and
	// keeps the keys.
	SecretData = "secret-data"
	// EnvValues drops the literal value of every container env var. Names
	// and valueFrom references stay (decision-5).
	EnvValues = "env-values"
	// LastApplied applies the other rules inside the
	// kubectl.kubernetes.io/last-applied-configuration annotation, which
	// repeats the object as it was applied, Secret data included.
	LastApplied = "last-applied-configuration"
)

const lastAppliedAnnotation = "kubectl.kubernetes.io/last-applied-configuration"

type Options struct {
	// KeepEnvValues turns off EnvValues.
	KeepEnvValues bool
}

type Redactor struct {
	opts Options
}

func New(opts Options) *Redactor {
	return &Redactor{opts: opts}
}

// Rules lists the rules this Redactor applies, in a fixed order.
func (r *Redactor) Rules() []string {
	rules := []string{SecretData}
	if !r.opts.KeepEnvValues {
		rules = append(rules, EnvValues)
	}
	return append(rules, LastApplied)
}

// Object redacts obj in place. group is "" for the core group.
func (r *Redactor) Object(group, kind string, obj map[string]any) {
	if group == "" && kind == "Secret" {
		for _, field := range []string{"data", "stringData"} {
			if m, ok := obj[field].(map[string]any); ok {
				for k := range m {
					m[k] = ""
				}
			}
		}
	}
	if !r.opts.KeepEnvValues {
		dropEnvValues(obj)
	}
	r.lastApplied(group, kind, obj)
}

// lastApplied redacts the object inside the last-applied annotation the same
// way, or drops the annotation when it can't be read.
func (r *Redactor) lastApplied(group, kind string, obj map[string]any) {
	meta, ok := obj["metadata"].(map[string]any)
	if !ok {
		return
	}
	annotations, ok := meta["annotations"].(map[string]any)
	if !ok {
		return
	}
	raw, ok := annotations[lastAppliedAnnotation]
	if !ok {
		return
	}
	s, ok := raw.(string)
	var applied map[string]any
	if ok {
		// UseNumber keeps numbers as written instead of rounding them
		// through float64.
		dec := json.NewDecoder(bytes.NewReader([]byte(s)))
		dec.UseNumber()
		if dec.Decode(&applied) != nil {
			applied = nil
		}
	}
	if applied == nil {
		delete(annotations, lastAppliedAnnotation)
		return
	}
	r.Object(group, kind, applied)
	out, err := json.Marshal(applied)
	if err != nil {
		delete(annotations, lastAppliedAnnotation)
		return
	}
	annotations[lastAppliedAnnotation] = string(out)
}

// dropEnvValues removes env[].value from every container list anywhere in v.
// Looking everywhere covers pods, workload templates and custom resources
// that embed a pod spec, without a list of where each kind keeps one.
func dropEnvValues(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			if k == "containers" || k == "initContainers" || k == "ephemeralContainers" {
				dropContainerEnv(child)
			}
			dropEnvValues(child)
		}
	case []any:
		for _, child := range t {
			dropEnvValues(child)
		}
	}
}

func dropContainerEnv(containers any) {
	list, ok := containers.([]any)
	if !ok {
		return
	}
	for _, c := range list {
		ctr, ok := c.(map[string]any)
		if !ok {
			continue
		}
		env, ok := ctr["env"].([]any)
		if !ok {
			continue
		}
		for _, e := range env {
			if ev, ok := e.(map[string]any); ok {
				delete(ev, "value")
			}
		}
	}
}
