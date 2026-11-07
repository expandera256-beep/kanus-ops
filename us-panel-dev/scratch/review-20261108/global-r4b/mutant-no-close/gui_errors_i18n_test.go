package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"kiyas/i18n"
)

// Regression: TR texts are byte-identical to the literals used before i18n.
var guiErrorCases = []struct{ key, tr, en string }{
	{"err.gui.method", "yöntem yok", "method not allowed"},
	{"err.gui.task_read", "Görev kaydı okunamadı.", "Could not read the task record."},
	{"err.gui.prompt_too_big", "İstem 8 MiB sınırını aşıyor; daha küçük bölümlere ayırın.", "The prompt exceeds the 8 MiB limit; split it into smaller parts."},
	{"err.gui.request_bad", "İstek okunamadı veya çok büyük.", "The request could not be read or is too large."},
	{"err.gui.conn_busy", "İstek devam ederken bağlantı değiştirilemez.", "The connection cannot be changed while a request is running."},
	{"err.gui.cancel_save", "İptal kaydedilemedi.", "Could not save the cancellation."},
	{"err.gui.pending_change", "Önce bekleyen değişikliği uygulayın veya reddedin.", "Apply or discard the pending change first."},
	{"err.gui.no_op", "işlem yok", "no such operation"},
	{"err.gui.not_image", "görsel değil", "not an image"},
	{"err.gui.bad_host", "geçersiz yerel adres", "invalid local address"},
	{"err.gui.bad_page", "geçersiz sayfa", "invalid page"},
	{"err.gui.bad_range", "geçersiz aralık", "invalid range"},
	{"err.gui.list_files", "dosyalar listelenemedi", "could not list files"},
	{"err.gui.direct_model", "Ticari model yalnız Go lisans proxy üzerinden çağrılır; doğrudan model adresi kapalı.", "The commercial model is only called through the Go license proxy; direct model addresses are disabled."},
	{"err.gui.task_done", "Tamamlanan görev yeniden yürütülemez; yeni bir görev gönderin.", "A completed task cannot be run again; submit a new task."},
	{"err.gui.ssh_dev", "SSH bağlantısı yalnızca açıkça etkinleştirilmiş geliştirici modunda kullanılabilir.", "The SSH connection is only available in explicitly enabled developer mode."},
	{"err.gui.python_gateway", "Python hesap geçidi bu Go lisans sürümünde kapalıdır.", "The Python account gateway is disabled in this Go license build."},
	{"err.gui.tunnel_windows", "Otomatik tünel yalnızca Windows sürümünde kullanılabilir.", "The automatic tunnel is only available in the Windows build."},
	{"err.gui.model_busy", "Model isteği zaten sürüyor.", "A model request is already running."},
	{"err.gui.model_name", "Model adı geçersiz.", "Invalid model name."},
	{"err.gui.model_set", "Model adı ayarlanamadı.", "Could not set the model name."},
	{"err.gui.queue_id", "Kuyruk kimliği oluşturulamadı.", "Could not create a queue ID."},
	{"err.gui.source_verify", "Kaynak doğrulanamadı.", "Could not verify the source."},
	{"err.gui.task_running", "Görev sürerken çalışma alanı, model veya dosyalar değiştirilemez.", "The workspace, model or files cannot be changed while a task is running."},
	{"err.gui.task_save", "Görev kaydedilemedi.", "Could not save the task."},
	{"err.gui.task_resume", "Bu görev sürdürülemez.", "This task cannot be resumed."},
	{"err.gui.task_cancel", "Bu görev iptal edilemez.", "This task cannot be cancelled."},
	{"err.gui.conn_set", "Bağlantı ayarlanamadı.", "Could not configure the connection."},
}

func TestGUIErrorTextsPreserveTurkish(t *testing.T) {
	b, err := i18n.Load()
	if err != nil {
		t.Fatal(err)
	}
	src, _ := os.ReadFile("gui.go")
	for _, c := range guiErrorCases {
		if got := b.For("tr").T(c.key); got != c.tr {
			t.Errorf("%s tr=%q want %q", c.key, got, c.tr)
		}
		if got := b.For("en").T(c.key); got != c.en {
			t.Errorf("%s en=%q want %q", c.key, got, c.en)
		}
		if strings.Contains(string(src), "http.Error(w, \""+c.tr+"\"") {
			t.Errorf("literal still hardcoded: %q", c.tr)
		}
		if !strings.Contains(string(src), "\""+c.key+"\"") {
			t.Errorf("key unused in gui.go: %s", c.key)
		}
	}
}

func TestGUIErrorFollowsPreference(t *testing.T) {
	s := newView("errlang")
	s.root = t.TempDir()
	h := guiMux(&s)
	call := func() string {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/version", nil))
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatal(w.Code)
		}
		return w.Body.String()
	}
	_ = os.Remove(langPref.Path())
	if got := call(); got != "yöntem yok\n" {
		t.Fatalf("default TR changed: %q", got)
	}
	if err := langPref.Set("en"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(langPref.Path())
	if got := call(); got != "method not allowed\n" {
		t.Fatalf("en: %q", got)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://evil.example/", nil))
	if !strings.HasPrefix(w.Body.String(), "invalid local address") {
		t.Fatalf("host error not localized: %q", w.Body.String())
	}
}
