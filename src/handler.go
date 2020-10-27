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
	str := fmt.Sprint(out)
	fmt.Println(str)
	fmt.Printf("%s", out)
	fmt.Println(reflect.TypeOf(out))

}
