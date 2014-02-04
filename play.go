package main

import (
  "fmt"
  "errors"
  "path/filepath"
  "os"
  "os/exec"
  "strings"
  "syscall"
  "bufio"
)

type track struct {
  fullPath string
  info os.FileInfo
}

type serverAction struct {
  action string
  track  track
}

func main() {
  basePath, err := getBasePath(os.Args[1:])
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  play(gatherTracks(basePath))
  fmt.Println("Bye.")
}

func play(tracks []track) {
  fmt.Println(len(tracks), "tracks found")

  playNext := make(chan track)
  playInfo := make(chan PlayInfo)
  go playTrack(playNext, playInfo)

  userInput := make(chan string)
  go console(userInput)

  var pInfo PlayInfo
	var trackName string

  for _, track := range tracks {
    playNext <- track
    for {
      select {
      case pInfo = <- playInfo:
				trackName = pInfo.t.info.Name()
      case userCommand := <- userInput:
        switch userCommand {
        case "p":
          pInfo.cmd.Process.Signal(syscall.SIGSTOP)
          fmt.Println("pausing...", trackName)
        case "r":
          pInfo.cmd.Process.Signal(syscall.SIGCONT)
          fmt.Println("resuming...", trackName)
        case "s":
          pInfo.cmd.Process.Signal(syscall.SIGTERM)
          fmt.Println("skipping...", trackName)
				default:
					fmt.Println("I can't understand \"", userCommand, "\" ...")
        }
      }
      if pInfo.action == "done" { break }
    }
  }
}

type PlayInfo struct {
  action string
  t track
  cmd *exec.Cmd
}

func playTrack(playNext <-chan track, infoChannel chan<- PlayInfo) {
  for {
    toPlay := <- playNext
    cmd := exec.Command("afplay", toPlay.fullPath)
    cmd.Start()
    fmt.Println("playing", toPlay.fullPath)
    infoChannel <- PlayInfo{action: "command", t: toPlay, cmd: cmd}
    cmd.Wait()
    infoChannel <- PlayInfo{action: "done", t: toPlay, cmd: cmd}
  }
}

func console(actions chan string) {
  reader := bufio.NewReader(os.Stdin)
  for {
    input, err := reader.ReadString('\n')
    if err == nil { actions <- strings.TrimSpace(input) }
  }
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
