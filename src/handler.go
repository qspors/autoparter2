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
	type Devices struct {
		Blockdevices string `json:"blockdevices"`
	}
	out, err := exec.Command("lsblk", "-J", "-a").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)
	bytes, err := json.Marshal(out)
	if err != nil {
		log.Fatal(err)
	}
	var d Devices
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", d)

}
