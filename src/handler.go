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
	fmt.Println(str)
	//
	type Type string
	const (
		Disk Type = "disk"
		Loop Type = "loop"
		Part Type = "part"
	)
	type Blockdevice struct {
		Name       string        `json:"name"`
		MajMin     string        `json:"maj:min"`
		Rm         string        `json:"rm"`
		Size       *string       `json:"size"`
		Ro         string        `json:"ro"`
		Type       Type          `json:"type"`
		Mountpoint *string       `json:"mountpoint"`
		Children   []Blockdevice `json:"children,omitempty"`
	}
	type Welcome struct {
		Blockdevices []Blockdevice `json:"blockdevices"`
	}

	//
	bytes, err := json.Marshal(str)
	if err != nil {
		fmt.Println(err)
	}
	var d Welcome

	err = json.Unmarshal(bytes, &d)

	fmt.Printf("%+v\n", d)
}
