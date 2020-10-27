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
	out, err := exec.Command("lsblk", "-P").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)

}
