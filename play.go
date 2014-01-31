package main

import (
  "fmt"
	"errors"
	"path/filepath"
	"os"
	"os/exec"
	"strings"
	"time"
)

type track struct {
	fullPath string
	info os.FileInfo
}

func main() {
	basePath, err := getBasePath(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	play(gatherTracks(basePath))
}

func play(tracks []track) {
	fmt.Println(len(tracks), "tracks found")
	queue := make(chan track)
	go playOne(queue)
	queue <- tracks[0]
	time.Sleep(time.Second)
}

func playOne(queue chan track) {
	t := <- queue
	cmd := exec.Command("afplay", t.fullPath )
	cmd.Run()
}

func gatherTracks(basePath string) []track {
	var tracks []track
	var bannedList []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.Name()[0] == '.' {
			bannedList = append(bannedList, path)
		} else if !info.IsDir() && !isInBannedList(path, bannedList) {
			tracks = append(tracks, track{fullPath: path, info: info})
		}
		return nil
	}
	filepath.Walk(basePath, walkFn)
	return tracks
}

func isInBannedList(path string, list []string) bool {
	for _, p := range list {
		if strings.HasPrefix(path, p) { return true }
	}
	return false
}

func getBasePath(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("Directory name was not provided")
	} else if !isDirectory(args[0]) {
		return args[0], errors.New("Error: The path provided does not correspond to a directory")
	} else {
		return args[0], nil
	}
}

func isDirectory(path string) bool {
	if info, err := os.Stat(path); err == nil { return info.IsDir() }
	return false
}
