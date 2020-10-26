package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type Drives struct {
	driveName string `json:"name"`
	driveSize string `json:"size"`
	driveType string `json:"type"`
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
	in := out
	bytes, err := json.Marshal(in)
	if err != nil {
		fmt.Println(err)
	}
	var d Drives
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", d)
}
