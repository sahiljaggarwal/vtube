package services

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func ProcessVideo(inputDir, outputDir string) error {
	resolutions := []string{"144p", "240p", "360p"}
	var wg sync.WaitGroup
	indexFile := filepath.Join(outputDir, "video/index.txt")
	err := filepath.Walk(inputDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir(){
			wg.Add(1)
			go func(p string ){
				defer wg.Done()
				err = convertAndUploadVideo(p, outputDir, resolutions)
				if err != nil {
					fmt.Printf("Error processing file %s: %v\n", p, err)
				}
			}(path)
		}
		return nil

	})
	
	if err != nil {
		fmt.Printf("Error walking through directory: %v\n", err)
		return err
	}

	wg.Wait()
	fmt.Printf("Video processing completed. Index file: %s\n", indexFile)
	return nil


}

func convertAndUploadVideo(inputPath, outputDir string, resolutions []string)error {
	var m3u8Content strings.Builder
	m3u8Content.WriteString("#EXTM3U\n")

	for _, res := range resolutions {
		// outputFile := fmt.Sprintf("%s_%s.mp4", filepath.Base(inputPath), res)
		cmd := exec.Command("ffmpeg", "-i", inputPath, "-vf", fmt.Sprintf("scale=-1:%s", res), "-f", "hls", "-hls_time", "10", "-hls_list_size", "0", "-hls_segment_filename", fmt.Sprintf("%s/%%03d_%s.ts", outputDir, res), fmt.Sprintf("%s/index_%s.m3u8", outputDir, res))
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error converting video to %s resolution: %v", res, err)
		}


		m3u8Content.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=,%s\n%s/index_%s.m3u8\n", getBandwidth(res), res, outputDir, res))
	}
	m3u8File := fmt.Sprintf("%s/index.m3u8", outputDir)
	err := writeM3U8File(m3u8File, m3u8Content.String())
	if err != nil {
		return fmt.Errorf("error writing master m3u8 file: %v", err)
	}

	return nil
}

func getBandwidth(resolution string) int {
	switch resolution {
	case "144p":
		return 300
	case "240p":
		return 500
	case "360p":
		return 800
	default:
		return 1000
	}
}

func writeM3U8File(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}
