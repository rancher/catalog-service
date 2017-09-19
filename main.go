package main

import (
	"os"
	"strconv"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/cmd"
	_ "github.com/rancher/catalog-service/signals"
)

func main() {
	cattleParentID := os.Getenv("CATTLE_PARENT_PID")
	if cattleParentID != "" {
		if pid, err := strconv.Atoi(cattleParentID); err == nil {
			go func() {
				for {
					process, err := os.FindProcess(pid)
					if err != nil {
						log.Fatalf("Failed to find process: %s\n", err)
					} else {
						err := process.Signal(syscall.Signal(0))
						if err != nil {
							log.Fatal("Parent process went away. Shutting down.")
						}
					}
					time.Sleep(time.Millisecond * 250)
				}
			}()
		}
	}
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
