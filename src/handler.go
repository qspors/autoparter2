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

	bytes, err := json.Marshal(str)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(bytes)

}
