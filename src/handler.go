package main

import (
	"fmt"
	"log"
	"os/exec"
)

var excludedDrives struct {
	loopD string
}

func main() {
	getDriveInfo()
}

func getDriveInfo() {
	out, err := exec.Command("lsblk", "-J", "-a").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)
}
