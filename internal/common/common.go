package common

import (
	"os"
	"path/filepath"
)

func CreateConfigDir() (string, error) {
	sysConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(sysConfigDir, "landlord-rts")
	err = os.MkdirAll(configDir, 0o755)
	if err != nil {
		return "", err
	}
	return configDir, nil
}
