//go:build !linux && !windows

package main

func descendantPids(root int) []int { return nil }
