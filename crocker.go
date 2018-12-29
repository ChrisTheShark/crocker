package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strconv"
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
		log.Fatalf("error mounting proc - %s\n", err)
	}
	if err := pivotRoot(newRootPath); err != nil {
		log.Fatalf("error executing pivot_root - %s\n", err)
	}
	if err := waitForNetwork(); err != nil {
		log.Fatalf("error waiting for network - %s\n", err)
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
	var netSetGoPath string

	flag.StringVar(&rootFsPath, "rootfs", "/tmp/ns-process/rootfs", "Path to the new root file system")
	flag.StringVar(&netSetGoPath, "netsetgo", "/usr/local/bin/netsetgo", "Path to the netsetgo binary")

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

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting the /bin/sh command - %s\n", err)
	}

	pid := strconv.Itoa(cmd.Process.Pid)
	netSetGoCmd := exec.Command(netSetGoPath, "-pid", pid)
	if err := netSetGoCmd.Run(); err != nil {
		log.Fatalf("Error running netsetgo binary - %s\n", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("Error starting the /bin/sh command - %s\n", err)
	}
}
