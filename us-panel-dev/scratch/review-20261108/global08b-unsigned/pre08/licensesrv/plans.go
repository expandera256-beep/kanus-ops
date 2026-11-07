package licensesrv

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ParsePlans accepts both the original kind schema and exported live plan schema.
// PriceTRY is informational, never used to infer purchased credits.
func ParsePlans(data []byte) (map[int64]Plan, error) {
	var raw map[string]struct {
		Plan
		Name     string `json:"plan"`
		PriceTRY int    `json:"price_try"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("plan listesi boş")
	}
	out := make(map[int64]Plan, len(raw))
	for key, v := range raw {
		id, err := strconv.ParseInt(key, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("geçersiz product_id: %s", key)
		}
		p := v.Plan
		if v.Name != "" {
			kind := ""
			switch v.Name {
			case "credit", "monthly":
				kind = "sub"
			case "hourly":
				kind = "use"
			default:
				return nil, fmt.Errorf("bilinmeyen plan: %s", v.Name)
			}
			if p.Kind != "" && p.Kind != kind {
				return nil, fmt.Errorf("plan/kind çelişkisi: %s", key)
			}
			p.Kind = kind
		}
		if p.Kind == "sub" {
			if p.Prompts == 0 {
				p.Prompts = p.Credits
			}
			if p.Credits == 0 {
				p.Credits = p.Prompts
			}
		}
		// Hourly packages have no prompt-credit price policy. Keep them time-metered
		// in legacy mode; credit HTTP API explicitly rejects them rather than inventing credits.
		if p.Days <= 0 || p.Days > 3650 || p.Credits < 0 || p.Prompts < 0 || p.Seconds < 0 ||
			(p.Kind != "sub" && p.Kind != "use") || (p.Kind == "sub" && (p.Prompts <= 0 || p.Credits <= 0)) || (p.Kind == "use" && p.Seconds <= 0) {
			return nil, fmt.Errorf("geçersiz plan: %s", key)
		}
		out[id] = p
	}
	return out, nil
}
