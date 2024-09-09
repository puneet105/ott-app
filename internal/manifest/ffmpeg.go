package manifest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type VideoResolution struct {
	Name   string
	Width  int
	Height int
}

const numberOfThreads = 4	

func GenerateManifest(inputFile, outputDir string) error {
	resolutions := []VideoResolution{
		{"360p", 640, 360},
		{"480p", 854, 480},
		{"720p", 1280, 720},
		{"1080p", 1920, 1080},
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(resolutions)*2)

	for _, res := range resolutions {
		wg.Add(2)

		go func(res VideoResolution) {
			defer wg.Done()
			hlsOutput := filepath.Join(outputDir, res.Name, "hls")

			if err := os.MkdirAll(hlsOutput, 0755); err != nil {
				errChan <- fmt.Errorf("error creating HLS output directory: %v", err)
				return
			}

			hlsCmd := exec.Command("ffmpeg", "-threads", fmt.Sprintf("%d", numberOfThreads), "-i", inputFile, "-vf", fmt.Sprintf("scale=%d:%d", res.Width, res.Height),
				"-c:a", "aac", "-ar", "48000", "-b:a", "128k", "-c:v", "h264", "-b:v", "1500k", "-maxrate", "1500k", "-bufsize", "3000k",
				"-hls_time", "10", "-hls_list_size", "0", "-f", "hls", filepath.Join(hlsOutput, "manifest.m3u8"))

			if err := hlsCmd.Run(); err != nil {
				errChan <- fmt.Errorf("error generating HLS manifest: %v", err)
			}
		}(res)

		go func(res VideoResolution) {
			defer wg.Done()
			dashOutput := filepath.Join(outputDir, res.Name, "dash")

			if err := os.MkdirAll(dashOutput, 0755); err != nil {
				errChan <- fmt.Errorf("error creating DASH output directory: %v", err)
				return
			}

			dashCmd := exec.Command("ffmpeg", "-threads", fmt.Sprintf("%d", numberOfThreads), "-i", inputFile, "-vf", fmt.Sprintf("scale=%d:%d", res.Width, res.Height),
				"-c:a", "aac", "-ar", "48000", "-b:a", "128k", "-c:v", "h264", "-b:v", "1500k", "-maxrate", "1500k", "-bufsize", "3000k",
				"-seg_duration", "10", "-adaptation_sets", "id=0,streams=v id=1,streams=a", "-f", "dash", filepath.Join(dashOutput, "manifest.mpd"))

			if err := dashCmd.Run(); err != nil {
				errChan <- fmt.Errorf("error generating DASH manifest: %v", err)
			}
		}(res)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	if err := generateMasterHLSManifest(outputDir, resolutions); err != nil {
		return fmt.Errorf("error generating master HLS manifest: %v", err)
	}
	if err := generateMasterDASHManifest(outputDir, resolutions); err != nil {
		return fmt.Errorf("error generating master DASH manifest: %v", err)
	}


	return nil

}

func generateMasterHLSManifest(outputDir string, resolutions []VideoResolution) error {
	masterManifestPath := filepath.Join(outputDir, "master.m3u8")
	file, err := os.Create(masterManifestPath)
	if err != nil {
		return fmt.Errorf("error creating master manifest file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString("#EXTM3U\n")
	if err != nil {
		return fmt.Errorf("error writing to master manifest file: %v", err)
	}

	for _, res := range resolutions {
		_, err = file.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=1500000,RESOLUTION=%dx%d\n", res.Width, res.Height))
		if err != nil {
			return fmt.Errorf("error writing to master manifest file: %v", err)
		}
		_, err = file.WriteString(filepath.Join(res.Name, "hls", "manifest.m3u8") + "\n")
		if err != nil {
			return fmt.Errorf("error writing to master manifest file: %v", err)
		}
	}

	return nil
}

func generateMasterDASHManifest(outputDir string, resolutions []VideoResolution) error {
	masterManifestPath := filepath.Join(outputDir, "master.mpd")
	file, err := os.Create(masterManifestPath)
	if err != nil {
		return fmt.Errorf("error creating master DASH manifest file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-live:2011" type="static" mediaPresentationDuration="PT0H20M0.00S" minBufferTime="PT1.5S">
  <Period>
    <AdaptationSet mimeType="video/mp4" segmentAlignment="true" startWithSAP="1" bitstreamSwitching="true">
`)
	if err != nil {
		return fmt.Errorf("error writing to master DASH manifest file: %v", err)
	}

	for _, res := range resolutions {
		_, err = file.WriteString(fmt.Sprintf(`
      <Representation id="%s" width="%d" height="%d" bandwidth="1500000">
        <BaseURL>%s/</BaseURL>
        <SegmentTemplate media="manifest$Number$.m4s" initialization="manifest-init.mp4" startNumber="1" />
      </Representation>`, res.Name, res.Width, res.Height, filepath.Join(res.Name, "dash")))
		if err != nil {
			return fmt.Errorf("error writing to master DASH manifest file: %v", err)
		}
	}

	_, err = file.WriteString(`
    </AdaptationSet>
  </Period>
</MPD>
`)
	if err != nil {
		return fmt.Errorf("error writing to master DASH manifest file: %v", err)
	}

	return nil
}