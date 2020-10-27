package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

func UnmarshalWelcome(data []byte) (Welcome, error) {
	var r Welcome
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Welcome) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Welcome struct {
	Blockdevices []Blockdevice `json:"blockdevices"`
}

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

type Type string

const (
	Disk Type = "disk"
	Loop Type = "loop"
	Part Type = "part"
)

func main() {
	out, err := exec.Command("lsblk", "-J", "-a").Output()
	if err != nil {
		log.Fatal(err)
	}
	r, err := UnmarshalWelcome(out)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
}
