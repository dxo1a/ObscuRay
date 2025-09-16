//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func run(cmd *exec.Cmd) {
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

var tags = "with_utls with_clash_api"

func main() {
	repo := "https://github.com/SagerNet/sing-box.git"
	dir := "build/sing-box"
	bin := "../../assets/sing-box.exe"

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Cloning sing-box...")
		run(exec.Command("git", "clone", "--depth=1", repo, dir))
	} else {
		fmt.Println("Updating sing-box...")
		run(exec.Command("git", "-C", dir, "pull"))
	}

	fmt.Printf("Building sing-box with tags [%s]...\n", tags)
	cmd := exec.Command("go", "build", "-tags", tags, "-o", bin, "./cmd/sing-box")
	cmd.Dir = dir
	run(cmd)

	fmt.Println("sing-box built at:", bin)
}
