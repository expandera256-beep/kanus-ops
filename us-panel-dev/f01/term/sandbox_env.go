package main

import "strings"

func replaceSandboxEnv(env []string, values map[string]string) []string {
	out := make([]string, 0, len(env)+len(values))
	for _, v := range env {
		k, _, _ := strings.Cut(v, "=")
		if _, ok := values[k]; !ok {
			out = append(out, v)
		}
	}
	for k, v := range values {
		out = append(out, k+"="+v)
	}
	return out
}
