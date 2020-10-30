package main

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
)

func main() {
	sgetUUID("/dev/xvdf1")
}

func sgetUUID(label string) string {
	var uuid string
	out, err := exec.Command("blkid", label, "--output", "export").Output()
	if err != nil {
		log.Println(err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		UUID := strings.Split(scanner.Text(), "=")
		if UUID[0] == "UUID" {
			uuid = UUID[1]
		}
	}
	return uuid
}
