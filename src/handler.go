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
	out, err := exec.Command("lsblk", "-P").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)
	fmt.Println(reflect.TypeOf(out))

}
