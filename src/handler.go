package main

import (
	"fmt"
	"log"
	"os/exec"
	"reflect"
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
	fmt.Println(reflect.TypeOf(str))
}
