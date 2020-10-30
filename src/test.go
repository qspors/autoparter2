package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	waitPartition2("/dev/xvdf1")
}
func waitPartition2(filePath string) {
	for {
		ok := func() bool {
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				fmt.Println("Not exit")
				return false
			}
			fmt.Println("Is exit")
			return true
		}
		if ok() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
