package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type Drives struct {
	blockdevices string `json:"blockdevices"`
	//driveName  string `json:"name"`
	//majMin     string `json:"maj:min"`
	//rM         string `json:"rm"`
	//driveSize  string `json:"size"`
	//rO         string `json:"ro"`
	//driveType  string `json:"type"`
	//tyPe       string `json:"type"`
	//mountPoint string `json:"mountpoint"`
	//children   string `json:"children"`
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
