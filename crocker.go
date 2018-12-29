package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("nsInitialization", nsInitialization)
	if reexec.Init() {
		os.Exit(0)
	}
}

func nsInitialization() {
	newRootPath := os.Args[1]
	if err := mountProc(newRootPath); err != nil {
		log.Fatalf("error mounting proc - %s", err)
	}
	if err := pivotRoot(newRootPath); err != nil {
		log.Fatalf("error executing pivot_root - %s", err)
	}
	nsRun()
}

func nsRun() {
	cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{"PS1=-[ns-process]- # "}

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running the /bin/sh command - %s", err)
	}
}

func main() {
	var rootFsPath string
	flag.StringVar(&rootFsPath, "rootfs", "/tmp/ns-process/rootfs", "Path to the new root file system")
	flag.Parse()

	cmd := reexec.Command("nsInitialization", rootFsPath)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			syscall.SysProcIDMap{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			syscall.SysProcIDMap{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}
