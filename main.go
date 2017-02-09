package main

import (
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/cmd"
)

func main() {
	if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_PDEATHSIG, uintptr(syscall.SIGTERM), 0); err != 0 {
		log.Fatalf("Failed to set parent death signal: %d", err)
	}

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
