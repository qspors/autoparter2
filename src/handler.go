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
	"time"
)

type State struct {
	stop  string
	start string
}
type Drives struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}
type BlockDevice struct {
	Name     string        `json:"name"`
	Size     string        `json:"size"`
	Children []BlockDevice `json:"children,omitempty"`
}
type Suffixes struct {
	Blockdevices []SuffixDevice `json:"blockdevices"`
}
type SuffixDevice struct {
	Name       string         `json:"name"`
	MajMin     string         `json:"maj:min"`
	Rm         string         `json:"rm"`
	Size       string         `json:"size"`
	Ro         string         `json:"ro"`
	Type       string         `json:"type"`
	Mountpoint interface{}    `json:"mountpoint"`
	Children   []SuffixDevice `json:"children,omitempty"`
}

func UnmarshalSuffix(data []byte) (Suffixes, error) {
	var r Suffixes
	err := json.Unmarshal(data, &r)
	return r, err
}

const (
	xfs  string = "xfs"
	ext4 string = "ext4"
	ext3 string = "ext3"
)

func UnmarshalDrives(data []byte) (Drives, error) {
	var r Drives
	err := json.Unmarshal(data, &r)
	return r, err
}
func getDrives() map[string]int64 {
	driveMap := make(map[string]int64)
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
				if strings.Contains(itm.Size, "G") {

					splitString := strings.FieldsFunc(itm.Size, func(r rune) bool {
						return strings.ContainsRune("G", r)
					})[0]
					size, err := strconv.Atoi(splitString)
					if err != nil {
						log.Fatal(err)
					}
					driveMap[itm.Name] = int64(size)

				} else if strings.Contains(itm.Size, "T") {
					splitString := strings.FieldsFunc(itm.Size, func(r rune) bool {
						return strings.ContainsRune("T", r)
					})[0]
					size, err := strconv.ParseFloat(splitString, 64)
					newSize := size * 1000
					if err != nil {
						log.Fatal(err)
					}
					driveMap[itm.Name] = int64(newSize)
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
func dirsExist(volInfo map[string]int64) bool {
	for key, _ := range volInfo {
		if _, err := os.Stat(key); os.IsNotExist(err) {
			log.Println("Create dir: ", key)
			_, err := exec.Command("mkdir", "-p", key).Output()
			if err != nil {
				log.Fatal(err)
				return false
			}
		}
	}
	return true
}
func serviceStatus(command string, services []string) {
	for _, item := range services {

		cmd := exec.Command("systemctl", "check", item)
		out, err := cmd.CombinedOutput()
		outString := string(out)
		outString = strings.TrimSpace(outString)
		outString = strings.Trim(outString, "\t \n")
		if err != nil {
			if _, ok := err.(*exec.ExitError); ok {
			} else {
				fmt.Printf("failed to run systemctl: %v", err)
				os.Exit(1)
			}
		}

		if strings.EqualFold(outString, "active") && command == "stop" {
			fmt.Printf("Stop sertvice: %s\n", item)
			invokeStop := exec.Command("systemctl", command, item)
			_, err2 := invokeStop.CombinedOutput()
			if err2 != nil {
				if exitErr2, ok := err2.(*exec.ExitError); ok {
					fmt.Printf("systemctl finished with non-zero: %v\n", exitErr2)
				} else {
					fmt.Printf("failed to run systemctl: %v", err2)
					os.Exit(1)
				}
			}
		}
		if strings.EqualFold(outString, "inactive") && command == "start" {
			fmt.Printf("Start sertvice: %s\n", item)
			invokeStart := exec.Command("systemctl", command, item)
			_, err3 := invokeStart.CombinedOutput()
			if err3 != nil {
				if exitErr3, ok := err3.(*exec.ExitError); ok {
					fmt.Printf("systemctl finished with non-zero: %v\n", exitErr3)
				} else {
					fmt.Printf("failed to run systemctl: %v", err3)
					os.Exit(1)
				}
			}
		}
	}
}
func compareVolumeAndDrives(drives map[string]int64, volumes map[string]int64, filesystem string) {
	//fmt.Println("*******************************")
	//fmt.Printf("Volumes before deletion: %+v\n", volumes)
	//fmt.Println("*******************************")
	for driveLabel, driveSize := range drives {
		for dirName, dirSize := range volumes {
			if driveSize == dirSize {
				doMountingActions(driveLabel, dirName, filesystem)
				delete(volumes, dirName)
			}
		}
	}
	//fmt.Printf("Volumes after deletion: %+v\n", volumes)
	//fmt.Println("*******************************")
	//fmt.Printf("Drives: %+v\n", drives)
	//fmt.Println("*******************************")

}
func doMountingActions(label string, dir string, filesystem string) {
	tempDir := fmt.Sprintf("/temp%s", label)
	createDrive(label, filesystem)
	mountDrive(label, tempDir)
	copyData(dir, tempDir)
	unmountDrive(label)
	mountDrive(label, dir)
	fstabConfig(label, dir)
	removeTempDir(tempDir)

}
func createDrive(label string, filesystem string) {
	fmt.Printf("Create new drive for :%s", fmt.Sprintf("/dev/%s", label))
	if _, err1 := exec.Command("parted", "-s", fmt.Sprintf("/dev/%s", label), "mktable", "gpt").Output(); err1 != nil {
		fmt.Println(err1)
	}
	if _, err2 := exec.Command("parted", "-s", fmt.Sprintf("/dev/%s", label), "mkpart", "primary", "0%", "100%").Output(); err2 != nil {
		fmt.Println(err2)
	}
	time.Sleep(3 * time.Second)
	getSuffix(label)
	//if _, err3 := exec.Command(fmt.Sprintf("mkfs.%s", filesystem), fmt.Sprintf("/dev/%s/%s", label, driveSuffix)).Output(); err3 != nil {
	//	fmt.Println(err3)
	//}
	fmt.Println("exit")
	os.Exit(0)

}
func mountDrive(label string, directory string)  {}
func unmountDrive(label string)                  {}
func copyData(dir string, tempDir string)        {}
func fstabConfig(label string, directory string) {}
func removeTempDir(directory string)             {}
func getSuffix(label string) string {
	fullLabel := fmt.Sprintf("/dev/%s", label)
	out, err := exec.Command("lsblk", "-J", "-a", fullLabel).Output()
	if err != nil {
		log.Fatal(err)
	}
	r, err := UnmarshalSuffix(out)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range r.Blockdevices {
		for _, name := range item.Children {
			fmt.Println("Name is: ", name.Name)
		}
	}
	return ""
}
func main() {
	state := State{start: "start", stop: "stop"}

	services := []string{"lxcfs", "cron"}
	driveMap := getDrives()
	volInfo := getVolumeInfo(getInstanceId())
	dirsExist(volInfo)
	serviceStatus(state.start, services)
	compareVolumeAndDrives(driveMap, volInfo, xfs)
}
