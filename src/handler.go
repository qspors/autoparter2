package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	getDriveInfo()
}

func getDriveInfo() {
	out, err := exec.Command("lsblk", "-J", "-a").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Output is %s", out)
}
