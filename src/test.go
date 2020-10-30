package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	getUUID("/dev/xvdf1")
}

func getUUID(label string) {
	out, err := exec.Command("blkid", label, "--output", "export").Output()
	if err != nil {
		log.Println(err)
	}
	newOut := string(out)
	fmt.Println(newOut)
}
