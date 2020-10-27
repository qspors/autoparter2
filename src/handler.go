package main

import (
	"encoding/json"
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
	str := string(out)

	type Devices struct {
		Blockdevices map[string]string `json:"blockdevices"`
	}

	bytes, err := json.Marshal(str)
	if err != nil {
		fmt.Println(err)
	}
	var d Devices
	err = json.Unmarshal(bytes, &d)

	fmt.Printf("%+v", d)
}
