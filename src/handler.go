package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func UnmarshalDrives(data []byte) (Drives, error) {
	var r Drives
	err := json.Unmarshal(data, &r)
	return r, err
}

type Drives struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}

type BlockDevice struct {
	Name     string        `json:"name"`
	Size     string        `json:"size"`
	Children []BlockDevice `json:"children,omitempty"`
}

func getDrives() map[string]string {
	driveMap := make(map[string]string)
	out, err := exec.Command("lsblk", "-J", "-a").Output()
	if err != nil {
		log.Fatal(err)
	}
	r, err := UnmarshalDrives(out)
	if err != nil {
		log.Fatal(err)
	}
	for idx, itm := range r.BlockDevices {
		switch itm.Name {
		case fmt.Sprintf("loop%d", idx):
		default:
			if len(itm.Children) == 0 {
				driveMap[itm.Name] = itm.Size
			}
		}
	}
	return driveMap
}

func getInstanceId() string {
	reader := strings.NewReader("")
	request, err := http.NewRequest("GET", " http://169.254.169.254/latest/meta-data/instance-id", reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v", request)
	return ""
}

func getTags() {}

func main() {
	getInstanceId()
}
