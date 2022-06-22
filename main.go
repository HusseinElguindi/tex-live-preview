package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	if err := os.MkdirAll("out", os.ModePerm); err != nil {
		log.Fatal("could not create output directory")
	}

	if err := recompileTeX(); err != nil {
		log.Fatal("failed to recompile TeX")
	}

	if err := exec.Command("open", "-a", "Skim", "out/main.pdf").Run(); err != nil {
		log.Fatal("could not open output pdf")
	}

	compileQueue := make(chan struct{})
	go func() {
		for {
			<-compileQueue
			if err := recompileTeX(); err != nil {
				log.Println("failed to recompile TeX")
			}
		}
	}()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}
				if strings.HasSuffix(event.Name, ".tex") {
					log.Printf("event: %s", event)
					select {
					case compileQueue <- struct{}{}:
					default:
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("./")
	if err != nil {
		log.Fatal(err)
	}

	<-make(chan struct{})
}

func recompileTeX() error {
	return exec.Command("pdflatex", "-shell-escape", "-interaction=nonstopmode", "-output-directory=out", "main.tex").Run()
}
