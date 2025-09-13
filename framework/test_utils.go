package framework

import (
	"encoding/json"
	"os"
)

func CreateTempConfig(deps map[string][]string) string {
	config := struct {
		Operators map[string]struct {
			Type   string                 `json:"type"`
			Config map[string]interface{} `json:"config"`
		} `json:"operators"`
		Dependencies map[string]string `json:"dependencies"`
	}{
		Operators: map[string]struct {
			Type   string                 `json:"type"`
			Config map[string]interface{} `json:"config"`
		}{
			"op1": {Type: "op1"},
			"op2": {Type: "op2"},
		},
		Dependencies: make(map[string]string),
	}

	for from, toList := range deps {
		for _, to := range toList {
			config.Dependencies[from] = to
		}
	}

	file, _ := os.CreateTemp("", "config-*.json")
	defer file.Close()
	json.NewEncoder(file).Encode(config)
	return file.Name()
}
