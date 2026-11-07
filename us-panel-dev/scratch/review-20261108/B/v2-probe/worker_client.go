package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type accountWorkerState struct {
	Worker         string `json:"worker"`
	Model          string `json:"model"`
	Phase          string `json:"phase"`
	ProviderStatus string `json:"provider_status"`
	Active         int    `json:"active"`
	IdleSeconds    int    `json:"idle_seconds"`
	IdleRemaining  *int   `json:"idle_remaining"`
	ObservedAt     int64  `json:"observed_at"`
	Live           bool   `json:"live"`
	ModelReady     bool   `json:"model_ready"`
}

func accountWorkers(base, token string) ([]accountWorkerState, error) {
	if validateGateway(base) != nil || len(token) < 32 || len(token) > 256 || strings.ContainsAny(token, "\r\n") {
		return nil, errors.New("Sunucu durumu için hesap girişi gerekli.")
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(base, "/")+"/account/workers", strings.NewReader("{}"))
	if err != nil {
		return nil, errors.New("Sunucu durum isteği hazırlanamadı.")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Sunucu durumu alınamadı.")
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 16385))
	if err != nil || len(raw) > 16384 || response.StatusCode != 200 {
		return nil, errors.New("Sunucu yönetimi doğrulanamadı.")
	}
	var result struct {
		Workers []accountWorkerState `json:"workers"`
	}
	if json.Unmarshal(raw, &result) != nil || len(result.Workers) > 10 {
		return nil, errors.New("Geçersiz sunucu durumu.")
	}
	for _, item := range result.Workers {
		if item.Worker == "" || len(item.Worker) > 63 || item.Model == "" || len(item.Model) > 128 || item.Active < 0 || item.Active > 10 || item.IdleSeconds < 0 || item.IdleSeconds > 1800 || item.ObservedAt < 0 || item.IdleRemaining != nil && (*item.IdleRemaining < 0 || *item.IdleRemaining > item.IdleSeconds) {
			return nil, errors.New("Geçersiz sunucu durumu.")
		}
	}
	return result.Workers, nil
}

func probeSession(s *session) resourceSnap {
	if hookLicenseMode() {
		unlock := s.hold()
		base, key, worker := activeGateway(), s.ledger.Token, s.worker
		unlock()
		snap := resourceSnap{Error: "Model hazır durumu doğrulanmadı."}
		if key == "" {
			return snap
		}
		req, err := http.NewRequest("GET", base+"/v1/models", nil)
		if err != nil {
			return snap
		}
		req.Header.Set("Authorization", "Bearer "+key)
		client := &http.Client{Timeout: 6 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, err := client.Do(req)
		if err != nil {
			return snap
		}
		defer resp.Body.Close()
		var listing struct {
			Data []struct {
				ID     string `json:"id"`
				Status struct {
					Value string `json:"value"`
				} `json:"status"`
			} `json:"data"`
		}
		if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&listing) != nil {
			return snap
		}
		snap.DoorOK = true
		snap.Error = ""
		for _, v := range listing.Data {
			if v.ID == selectedModel(worker) {
				snap.Workers = append(snap.Workers, workerStatus{ID: string(worker), Name: v.ID, Awake: v.Status.Value == "" || v.Status.Value == "loaded"})
			}
		}
		return snap
	}
	unlock := s.hold()
	base, token, worker := s.accountGateway, s.key, s.worker
	unlock()
	if base == "" || base != strings.TrimRight(activeGateway(), "/") {
		return probe()
	}
	started := time.Now()
	items, err := accountWorkers(base, token)
	snap := resourceSnap{DoorMS: int(time.Since(started).Milliseconds())}
	if err != nil {
		snap.Error = err.Error()
		return snap
	}
	snap.DoorOK = true
	snap.Managed = items
	for _, item := range items {
		if item.Model == selectedModel(worker) {
			snap.Workers = append(snap.Workers, workerStatus{ID: string(worker), Name: item.Worker, Awake: item.ModelReady})
		}
	}
	return snap
}

func (s *session) connectionStatus() string {
	if hookLicenseMode() {
		return "Go sunucu model proxy · lisans/kredi sunucuda · doğrudan SSH/Cloud Run hesabı kapalı"
	}
	if s.accountGateway != "" && s.accountGateway == strings.TrimRight(activeGateway(), "/") {
		return "Hesap geçidine bağlı · model sunucusu yalnız yetkili model isteğiyle açılır"
	}
	if commercialRequired() {
		return "HTTPS geçidini ayarlayın ve hesabınıza giriş yapın"
	}
	return "Model adresini ayarlayın · panel açılışı sunucu başlatmaz"
}
