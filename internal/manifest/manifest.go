package manifest

import (
	"os"
)


func StoreManifest(outputDir string, manifestData []byte) error {
	err := os.WriteFile(outputDir+"/manifest.m3u8", manifestData, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
