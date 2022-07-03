package preview

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

func Start(_ *cobra.Command, args []string) error {
	path := args[0]
	parentDir := filepath.Dir(path)
	outputDir := filepath.Join(parentDir, "out")

	// Mkdir all returns nil if the folder already exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}

	// Initial compilation
	if err := compileTeX(parentDir, path); err != nil {
		return err
	}

	// Open PDF viewer
	cmd := exec.Command("open", "-a", "Skim", "out/main.pdf")
	cmd.Dir = parentDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not open output pdf: %w", err)
	}

	// Queue singaling a TeX compilation
	compileQueue := make(chan struct{})

	// Start watching directories
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	watchTeXChanges(watcher, compileQueue, parentDir)

	// Process compile trigger events
	for {
		<-compileQueue
		if err := compileTeX(parentDir, path); err != nil {
			log.Println(err)
		}
	}
}

func watchTeXChanges(watcher *fsnotify.Watcher, queuec chan struct{}, dirs ...string) error {
	// Process filesystem events without blocking
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Ensure the operation is an non-permission change operation on a .tex file
				if event.Op&fsnotify.Chmod == fsnotify.Chmod || strings.HasSuffix(event.Name, ".tex") {
					continue
				}

				log.Printf("event: %s", event)

				// If queue is not full, add the event to queue (i.e. try-send the fs event)
				select {
				case queuec <- struct{}{}:
				default:
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Watch the passed directories
	for _, dir := range dirs {
		err := watcher.Add(dir)
		if err != nil {
			return fmt.Errorf("unable to watch %s: %w", dir, err)
		}
	}
	return nil
}

func compileTeX(parentDir, filepath string) error {
	cmd := exec.Command("pdflatex", "-shell-escape", "-interaction=nonstopmode", "-output-directory=out", "-jobname=main", filepath)
	cmd.Dir = parentDir

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to compile TeX: %w", err)
	}
	return nil
}
