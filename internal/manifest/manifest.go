// internal/manifest/manifest.go
package manifest

import (
	"os"
)

// StoreManifest stores the generated manifest in a specified directory
func StoreManifest(outputDir string, manifestData []byte) error {
	err := os.WriteFile(outputDir+"/manifest.m3u8", manifestData, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
