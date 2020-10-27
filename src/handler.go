package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

func UnmarshalDrives(data []byte) (Drives, error) {
	var r Drives
	err := json.Unmarshal(data, &r)
	return r, err
}

type Drives struct {
	Blockdevices []Blockdevice `json:"blockdevices"`
}

type Blockdevice struct {
	Name     string        `json:"name"`
	Size     string        `json:"size"`
	Type     string        `json:"type"`
	Children []Blockdevice `json:"children,omitempty"`
}

func lsblkUtil() {
	//driveMap := make(map[string]string)
	//fmt.Println(driveMap)
	out, err := exec.Command("lsblk", "-J", "-a").Output()
	if err != nil {
		log.Fatal(err)
	}
	r, err := UnmarshalDrives(out)
	if err != nil {
		log.Fatal(err)
	}
	for _, itm := range r.Blockdevices {
		switch itm.Name {
		case "loop0":
			fmt.Println("loop0")
		case "loop1":
			fmt.Println("loop1")
		case "loop2":
			fmt.Println("loop2")
		}
	}
}

func main() {
	lsblkUtil()
}
