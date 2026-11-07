package main

import "testing"

func TestCompletionFeatureIsNotProjectRepair(t *testing.T) {
	if projectRepairRequest("Kapsamlı proje oluştur: görev ekleme, listeleme, tamamlama ve silme") {
		t.Fatal("task completion was mistaken for a request to repair an existing project")
	}
	for _, text := range []string{"projeyi tamamla", "projeyi tamamlayın", "projeyi onar", "uygulamayı düzelt"} {
		if !projectRepairRequest(text) { t.Fatalf("repair request missed: %s", text) }
	}
}
