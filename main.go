package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		fmt.Printf("usage: %s <src> <dest> [<dest>]\n", os.Args[0])
		os.Exit(1)
	}

	if err := copyAll(args[0], args[1:]); err != nil {
		panic(err)
	}
}

func copyAll(src string, dest []string) error {
	kill, err := runContainerd()
	if err != nil {
		return err
	}
	defer kill()

	if err := fetch(src); err != nil {
		return err
	}
	for _, d := range dest {
		if err := copy(src, d); err != nil {
			return err
		}
	}
	return nil
}

func fetch(src string) error {
	return run([]string{"content", "fetch", "--all-platforms", src})
}

func copy(src string, dest string) error {
	return run([]string{"images", "push", dest, src})
}

func runContainerd() (func(), error) {
	cmd := exec.Command("containerd")
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	i := 0
	for {
		if i > 10 {
			return nil, fmt.Errorf("containerd did not start")
		}
		if testContainerd() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return func() {
		cmd.Process.Kill()
	}, nil
}

func testContainerd() bool {
	cmd := exec.Command("ctr", "image", "ls")
	err := cmd.Run()
	return err == nil
}

func run(args []string) error {
	cmdArgs := append([]string{}, args[:2]...)
	args = append(append(cmdArgs, []string{"-u", "file:/root/.docker/config.json"}...), args[2:]...)
	fmt.Println("ctr", args)
	cmd := exec.Command("ctr", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
