package main

import (
	"fmt"
	"os/exec"
)

func main() {
	getDriveInfo()
}

func getDriveInfo() {
	out, err := exec.Command("lsblk -J -a").Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(out)
}
