package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"time"

	"cg/internal"

	"github.com/g70245/win"
)

const (
	gameClientWidth  = 640
	gameClientHeight = 480
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "windows":
		return runWindows(args[1:])
	case "capture":
		return runCapture(args[1:])
	case "scratch":
		if len(args) != 1 {
			return fmt.Errorf("scratch does not accept arguments")
		}
		runScratch()
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q; use help to list commands", args[0])
	}
}

func runWindows(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("windows does not accept arguments")
	}

	games := internal.FindWindows()
	handles := make([]string, 0, len(games))
	for handle := range games {
		handles = append(handles, handle)
	}
	sort.Strings(handles)

	if len(handles) == 0 {
		fmt.Println("No compatible game windows found.")
		return nil
	}

	for _, handle := range handles {
		fmt.Println(handle)
	}
	return nil
}

func runCapture(args []string) error {
	flags := flag.NewFlagSet("capture", flag.ContinueOnError)
	handle := flags.Uint64("handle", 0, "decimal HWND of the game client")
	output := flags.String("output", filepath.Join(os.TempDir(), "cg-window-capture.png"), "PNG output path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("capture received unexpected arguments: %v", flags.Args())
	}
	if *handle == 0 {
		return fmt.Errorf("capture: -handle is required")
	}

	hWnd := win.HWND(*handle)
	if !isCompatibleGameWindow(hWnd) {
		return fmt.Errorf("capture: handle %d is not a discovered compatible game window", *handle)
	}

	startedAt := time.Now()
	capture, err := internal.CaptureClientArea(hWnd, 0, 0, gameClientWidth, gameClientHeight)
	if err != nil {
		return fmt.Errorf("capture game window %d: %w", *handle, err)
	}
	captureDuration := time.Since(startedAt)

	if err := writePNG(*output, capture); err != nil {
		return err
	}

	fmt.Printf("Captured HWND %d in %s\n", *handle, captureDuration)
	fmt.Printf("Output: %s\n", *output)
	return nil
}

func isCompatibleGameWindow(hWnd win.HWND) bool {
	for _, discovered := range internal.FindWindows() {
		if discovered == hWnd {
			return true
		}
	}
	return false
}

func writePNG(path string, capture image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create capture output: %w", err)
	}

	if err := png.Encode(file, capture); err != nil {
		file.Close()
		return fmt.Errorf("encode capture output: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close capture output: %w", err)
	}
	return nil
}

func printUsage() {
	fmt.Println(`Usage:
  go run ./cmd/cg-helper windows
  go run ./cmd/cg-helper capture -handle <HWND> [-output <path>]
  go run ./cmd/cg-helper scratch`)
}
