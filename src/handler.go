package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Drives struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}

type BlockDevice struct {
	Name     string        `json:"name"`
	Size     string        `json:"size"`
	Children []BlockDevice `json:"children,omitempty"`
}

func UnmarshalDrives(data []byte) (Drives, error) {
	var r Drives
	err := json.Unmarshal(data, &r)
	return r, err
}

func getDrives() map[string]int {
	driveMap := make(map[string]int)
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
				fmt.Println(itm.Size)
				if strings.Contains(itm.Size, "G") {

					splitString := strings.FieldsFunc(itm.Size, func(r rune) bool {
						return strings.ContainsRune("G", r)
					})[0]
					size, err := strconv.Atoi(splitString)
					if err != nil {
						log.Fatal(err)
					}
					driveMap[itm.Name] = size

				} else if strings.Contains(itm.Size, "T") {
					splitString := strings.FieldsFunc(itm.Size, func(r rune) bool {
						return strings.ContainsRune("T", r)
					})[0]
					size, err := strconv.ParseFloat(splitString, 64)
					newSize := size * 1000
					if err != nil {
						log.Fatal(err)
					}
					driveMap[itm.Name] = int(newSize)
				}
			}
		}
	}
	return driveMap
}

func getInstanceId() string {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

func getVolumeInfo(instanceId string) map[string]int64 {
	driveMap := make(map[string]int64)
	ses, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		log.Fatal(err)
	}
	svc := ec2.New(ses)
	input := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{{
			Name: aws.String("attachment.instance-id"),
			Values: []*string{
				aws.String(instanceId),
			},
		},
		},
	}
	response, err := svc.DescribeVolumes(input)
	if err != nil {
		log.Fatal(err)
	}
	for _, vol := range response.Volumes {
		for _, tags := range vol.Tags {
			if *tags.Key == "mount" {
				if *tags.Value != "none" {
					driveMap[*tags.Value] = *vol.Size
				}

			}
		}
	}
	return driveMap
}

func prepareDirs(volInfo map[string]int64) bool {
	for key, _ := range volInfo {
		_, err := os.Stat(key)
		if err != nil {
			log.Fatal(err)
		}
	}
	return true
}

func main() {
	//driveMap := getDrives()
	volInfo := getVolumeInfo(getInstanceId())
	dirsIsReady := prepareDirs(volInfo)
	fmt.Println(dirsIsReady)
}
