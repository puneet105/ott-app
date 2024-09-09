package manifest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// VideoResolution defines the resolution and bitrates
type VideoResolution struct {
	Name   string
	Width  int
	Height int
}

const numberOfThreads = 4	
// GenerateManifests generates HLS and DASH manifests for different resolutions
func GenerateManifest(inputFile, outputDir string) error {
	resolutions := []VideoResolution{
		{"360p", 640, 360},
		{"480p", 854, 480},
		{"720p", 1280, 720},
		{"1080p", 1920, 1080},
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(resolutions)*2)

	// Generate HLS and DASH manifests for each resolution concurrently
	for _, res := range resolutions {
		wg.Add(2) // Adding two tasks to the WaitGroup, one for HLS and one for DASH

		go func(res VideoResolution) {
			defer wg.Done()
			hlsOutput := filepath.Join(outputDir, res.Name, "hls")

			// Create the HLS output directory
			if err := os.MkdirAll(hlsOutput, 0755); err != nil {
				errChan <- fmt.Errorf("error creating HLS output directory: %v", err)
				return
			}

			// Generate HLS manifest
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

			// Create the DASH output directory
			if err := os.MkdirAll(dashOutput, 0755); err != nil {
				errChan <- fmt.Errorf("error creating DASH output directory: %v", err)
				return
			}

			// Generate DASH manifest
			dashCmd := exec.Command("ffmpeg", "-threads", fmt.Sprintf("%d", numberOfThreads), "-i", inputFile, "-vf", fmt.Sprintf("scale=%d:%d", res.Width, res.Height),
				"-c:a", "aac", "-ar", "48000", "-b:a", "128k", "-c:v", "h264", "-b:v", "1500k", "-maxrate", "1500k", "-bufsize", "3000k",
				"-seg_duration", "10", "-adaptation_sets", "id=0,streams=v id=1,streams=a", "-f", "dash", filepath.Join(dashOutput, "manifest.mpd"))

			if err := dashCmd.Run(); err != nil {
				errChan <- fmt.Errorf("error generating DASH manifest: %v", err)
			}
		}(res)
	}

	// Wait for all go routines to complete
	wg.Wait()
	close(errChan)

	// Check if any errors occurred
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Generate Master HLS Manifest
	if err := generateMasterHLSManifest(outputDir, resolutions); err != nil {
		return fmt.Errorf("error generating master HLS manifest: %v", err)
	}
	// Generate Master DASH Manifest
	if err := generateMasterDASHManifest(outputDir, resolutions); err != nil {
		return fmt.Errorf("error generating master DASH manifest: %v", err)
	}


	return nil

	// for _, res := range resolutions {
	// 	hlsOutput := filepath.Join(outputDir, res.Name, "hls")
	// 	dashOutput := filepath.Join(outputDir, res.Name, "dash")

	// 	// Create directories
	// 	if err := os.MkdirAll(hlsOutput, 0755); err != nil {
	// 		return fmt.Errorf("error creating HLS output directory: %v", err)
	// 	}
	// 	if err := os.MkdirAll(dashOutput, 0755); err != nil {
	// 		return fmt.Errorf("error creating DASH output directory: %v", err)
	// 	}

	// 	// Generate HLS manifest
	// 	hlsCmd := exec.Command("ffmpeg", "-i", inputFile, "-vf", fmt.Sprintf("scale=%d:%d", res.Width, res.Height),
	// 		"-c:a", "aac", "-ar", "48000", "-b:a", "128k", "-c:v", "h264", "-b:v", "1500k", "-maxrate", "1500k", "-bufsize", "3000k",
	// 		"-hls_time", "10", "-hls_list_size", "0", "-f", "hls", filepath.Join(hlsOutput, "manifest.m3u8"))
	// 	if err := hlsCmd.Run(); err != nil {
	// 		return fmt.Errorf("error generating HLS manifest: %v", err)
	// 	}

	// 	// Generate DASH manifest
	// 	dashCmd := exec.Command("ffmpeg", "-i", inputFile, "-vf", fmt.Sprintf("scale=%d:%d", res.Width, res.Height),
	// 		"-c:a", "aac", "-ar", "48000", "-b:a", "128k", "-c:v", "h264", "-b:v", "1500k", "-maxrate", "1500k", "-bufsize", "3000k",
	// 		"-seg_duration", "10", "-adaptation_sets", "id=0,streams=v id=1,streams=a", "-f", "dash", filepath.Join(dashOutput, "manifest.mpd"))
	// 	if err := dashCmd.Run(); err != nil {
	// 		return fmt.Errorf("error generating DASH manifest: %v", err)
	// 	}
	// }

	// // Generate Master HLS Manifest
	// if err := generateMasterHLSManifest(outputDir, resolutions); err != nil {
	// 	return fmt.Errorf("error generating master HLS manifest: %v", err)
	// }

	// return nil
}

// generateMasterHLSManifest creates a master HLS playlist
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

// generateMasterDASHManifest creates a master DASH manifest
func generateMasterDASHManifest(outputDir string, resolutions []VideoResolution) error {
	masterManifestPath := filepath.Join(outputDir, "master.mpd")
	file, err := os.Create(masterManifestPath)
	if err != nil {
		return fmt.Errorf("error creating master DASH manifest file: %v", err)
	}
	defer file.Close()

	// Write the basic structure of a DASH master manifest
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


// package manifest

// import (
// 	"fmt"
// 	"os/exec"
// 	"path/filepath"
// )

// // GenerateManifest generates a HLS manifest for a given MP4 file
// func GenerateManifest(inputFile, outputDir string) error {
// 	cmd := exec.Command(
// 		"ffmpeg",
// 		"-i", inputFile,
// 		"-codec:","copy", // Ensure that we are copying codecs
// 		"-start_number", "0", // Starting segment number
// 		"-hls_time", "10", // Segment duration in seconds
// 		"-hls_list_size", "0", // Unlimited number of segments in playlist
// 		"-hls_segment_filename", filepath.Join(outputDir, "manifest%03d.ts"), // Pattern for segment filenames
// 		filepath.Join(outputDir, "manifest.m3u8"), // Output manifest file
// 	)
// 	// cmd := exec.Command("ffmpeg", 
// 	// 	"-i", inputFile, 
// 	// 	"-codec:", "copy", 
// 	// 	"-start_number", "0", 
// 	// 	"-hls_time", "10", 
// 	// 	"-hls_list_size", "0", 
// 	// 	"-f", "hls", 
// 	// 	outputDir+"/manifest.m3u8")

// 	err := cmd.Run()
// 	if err != nil {
// 		return fmt.Errorf("error generating manifest: %v", err)
// 	}

// 	return nil
// }
