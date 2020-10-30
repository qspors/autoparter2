package main

import (
	"bufio"
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
	for key := range volInfo {
		if _, err := os.Stat(key); os.IsNotExist(err) {
			if _, err := exec.Command("mkdir", "-p", key).Output(); err != nil {
				log.Println(err)
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
	for driveLabel, driveSize := range drives {
		for dirName, dirSize := range volumes {
			if driveSize == dirSize {
				log.Printf("Action for drive: %s, dir: %s\n", driveLabel, dirName)
				log.Println("####################################################")
				doMountingActions(driveLabel, dirName, filesystem)
				delete(volumes, dirName)
				log.Println("####################################################")
			}
		}
	}
}
func doMountingActions(label string, dir string, filesystem string) {
	log.Printf("Start doMountingActions for: %s\n", label)
	tempDir := fmt.Sprintf("/temp%s", label)
	fullLabel := createDrive(label, filesystem)
	old := fmt.Sprintf("%s.old", dir)
	createTempDir(tempDir)
	mountDrive(fullLabel, tempDir)
	// Data move
	copyData(dir, tempDir)
	moveData(dir, old)
	unmountDrive(fullLabel)
	moveData(tempDir, dir)
	mountDrive(fullLabel, dir)
	removeTempDir(old)
	fstabConfig(fullLabel, dir, filesystem)

}
func createDrive(label string, filesystem string) string {
	log.Printf("Start createDrive for: %s\n", label)
	labelPath := fmt.Sprintf("/dev/%s", label)
	formatCommand := fmt.Sprintf("mkfs.%s", filesystem)
	log.Printf("mktable for drive: %s\n", labelPath)
	if _, err1 := exec.Command("parted", "-s", labelPath, "mktable", "gpt").Output(); err1 != nil {
		log.Println(err1)
	}
	log.Printf("mkpart for labelpath: %s\n", labelPath)
	if _, err2 := exec.Command("parted", "-s", labelPath, "mkpart", "primary", "0%", "100%").Output(); err2 != nil {
		log.Println(err2)
	}
	driveSuffix := getSuffix(label)
	fullPartPath := fmt.Sprintf("/dev/%s", driveSuffix)
	time.Sleep(5 * time.Second)
	log.Printf("format for fullpath: %s\n", fullPartPath)
	if _, err3 := exec.Command(formatCommand, "-f", fullPartPath).Output(); err3 != nil {
		log.Println(err3)
	}
	return fullPartPath
}
func createTempDir(tempDir string) {
	log.Printf("Create tempdir : %s\n", tempDir)
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err := os.MkdirAll(tempDir, 0700)
		if err != nil {
			log.Println(err)
		}
	}
}
func mountDrive(label string, directory string) {
	log.Printf("Mount drive: %s to dir: %s\n", label, directory)
	_, err := exec.Command("mount", label, directory).Output()
	if err != nil {
		log.Println(err)
	}
}
func unmountDrive(label string) {
	log.Printf("Umount drive: %s\n", label)
	_, err := exec.Command("umount", label).Output()
	if err != nil {
		log.Println(err)
	}
}
func copyData(src string, dst string) {
	log.Printf("Copy data, source: %s, destination: %s\n", src, dst)
	if _, err1 := exec.Command("rsync", "-raX", src+"/", dst+"/").Output(); err1 != nil {
		log.Println(err1)
	}
}
func moveData(src string, dst string) {
	log.Printf("Move data, source: %s, destination: %s\n", src, dst)
	if _, err1 := exec.Command("mv", src, dst).Output(); err1 != nil {
		log.Println(err1)
	}
}
func removeTempDir(directory string) {
	log.Printf("Remove old, dir: %s\n", directory)
	if _, err1 := exec.Command("/bin/rm -Rf", directory).Output(); err1 != nil {
		log.Println(err1)
	}
}
func fstabConfig(label string, directory string, fsType string) {
	uuid := getUUID(label)
	var uuidString string
	log.Printf("UUID for: %s is: %s for directory: %s\n", label, uuid, directory)
	file, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if directory == "/tmp" {
		uuidString = fmt.Sprintf("\nUUID=%s %s %s nodev,nosuid,noexec 0 0", uuid, directory, fsType)
	} else {
		uuidString = fmt.Sprintf("\nUUID=%s %s %s defaults 0 0", uuid, directory, fsType)
	}

	if _, err := file.WriteString(uuidString); err != nil {
		log.Println(err)
	}

}
func getSuffix(label string) string {
	log.Printf("Get suffix for: %s\n", label)
	var childName string
	fullLabel := fmt.Sprintf("/dev/%s", label)
	out, err := exec.Command("lsblk", "-J", "-a", fullLabel).Output()
	if err != nil {
		log.Println(err)
	}
	r, err := UnmarshalSuffix(out)
	if err != nil {
		log.Println(err)
	}
	for _, item := range r.Blockdevices {

		for _, name := range item.Children {
			childName = name.Name

		}
	}
	return childName
}
func getUUID(label string) string {
	var uuid string
	out, err := exec.Command("blkid", label, "--output", "export").Output()
	if err != nil {
		log.Println(err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		UUID := strings.Split(scanner.Text(), "=")
		if UUID[0] == "UUID" {
			uuid = UUID[1]
		}
	}
	return uuid
}
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
func main() {
	if len(os.Args) != 2 {
		log.Println("Filesystem is not specified or to many args")
		os.Exit(1)
	}
	fileSystems := []string{"xfs", "ext3", "ext4"}
	if _, found := Find(fileSystems, os.Args[1]); !found {
		log.Println("Filesystem is not correct")
		log.Println("Correct is:")
		for _, item := range fileSystems {
			log.Printf("> %s\n", item)
		}
		os.Exit(1)
	}
	FileSystemType := os.Args[1]

	state := State{start: "start", stop: "stop"}
	services := []string{"lxcfs", "cron"}
	driveMap := getDrives()
	volInfo := getVolumeInfo(getInstanceId())
	dirsExist(volInfo)
	serviceStatus(state.start, services)
	compareVolumeAndDrives(driveMap, volInfo, FileSystemType)
}
