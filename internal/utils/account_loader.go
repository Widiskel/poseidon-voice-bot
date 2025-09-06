package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadAccounts(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open accounts file: %w", err)
	}
	defer f.Close()

	var accs []string
	if err := json.NewDecoder(f).Decode(&accs); err != nil {
		return nil, fmt.Errorf("decode accounts: %w", err)
	}
	return accs, nil
}
