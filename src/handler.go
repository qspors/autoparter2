package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type Drives struct {
	Blockdevices []map[string]string `json:"blockdevices"`
}

func main() {
	getDriveInfo()
}

func getDriveInfo() {
	out, err := exec.Command("lsblk", "-J", "-a").Output()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := json.Marshal(out)
	if err != nil {
		log.Fatal(err)
	}
	var d Drives
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", d)

}
