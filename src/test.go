package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
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
	scanner := bufio.NewScanner(strings.NewReader(newOut))
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
