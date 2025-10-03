package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func GetVideoAspectRatio(filePath string) (string,error) {

	PATH_TO_VIDEO := filePath

	cmd := exec.Command("ffprobe", "-v", "error" ,"-print_format", "json" ,"-show_streams",PATH_TO_VIDEO)

	var byteReader bytes.Buffer
	cmd.Stdout = &byteReader

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	var jsonOutput map[string]interface{}

	err = json.Unmarshal(byteReader.Bytes(),&jsonOutput)
	if err != nil {
		return "", err
	}

	streams, ok := jsonOutput["streams"]
	if !ok {
		return "", fmt.Errorf("streams key not found in JSON")
	}

	streamsArr, ok := streams.([]interface{})
	if !ok || len(streamsArr) == 0 {
		return "", fmt.Errorf("streams is not a valid array")
	}

	firstStream, ok := streamsArr[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("first stream is not a JSON object")
	}

	aspectRatio, ok := firstStream["display_aspect_ratio"].(string)
	if !ok {
		return "", fmt.Errorf("display_aspect_ratio not found or not a string")
	}

	return aspectRatio, nil
}



func ProcessVideoForFastStart(filePath string) (string,error) {

	outputFilePath := filePath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return outputFilePath, nil
}

