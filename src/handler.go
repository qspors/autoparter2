package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"io/ioutil"
	"log"
	"net/http"
)

//type State struct {
//	stop  string
//	start string
//}
//type Drives struct {
//	BlockDevices []BlockDevice `json:"blockdevices"`
//}
//type BlockDevice struct {
//	Name     string        `json:"name"`
//	Size     json.Number   `json:"size"`
//	Children []BlockDevice `json:"children,omitempty"`
//}
//
//func pvCreate(label string) {
//	label = fmt.Sprintf("/dev/%s", label)
//	_, err := exec.Command("pvcreate", label).Output()
//	if err != nil {
//		log.Fatal(err)
//
//	}
//}
//func pvGroupCreate(label string) {
//	label = fmt.Sprintf("/dev/%s", label)
//	_, err := exec.Command("vgcreate", "group1", label).Output()
//	if err != nil {
//		log.Fatal(err)
//	}
//}
//func getFs(fs string) string {
//	fileSystems := []string{"xfs", "ext3", "ext4"}
//	if _, found := findInSlice(fileSystems, fs); !found {
//		log.Println("Filesystem is not correct")
//		log.Println("Correct is:")
//		for _, item := range fileSystems {
//			log.Printf("> %s\n", item)
//		}
//		os.Exit(1)
//	}
//	return fs
//}
//func findInSlice(slice []string, val string) (int, bool) {
//	for i, item := range slice {
//		if item == val {
//			return i, true
//		}
//	}
//	return -1, false
//}
//func prepareService(services string) []string {
//	stringSlice := strings.Split(services, ",")
//	return stringSlice
//}
//func getDrives() map[string]int64 {
//	driveMap := make(map[string]int64)
//	out, err := exec.Command("lsblk", "-J", "-b").Output()
//	if err != nil {
//		log.Fatal(err)
//	}
//	r, err := unmarshalDrives(out)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for idx, itm := range r.BlockDevices {
//		switch itm.Name {
//		case fmt.Sprintf("loop%d", idx):
//		default:
//			if len(itm.Children) == 0 {
//				size, err := itm.Size.Int64()
//				if err != nil {
//					log.Println(err)
//				}
//				driveMap[itm.Name] = size / 1024 / 1024 / 1024
//			}
//		}
//	}
//	return driveMap
//}
//func unmarshalDrives(data []byte) (Drives, error) {
//	var r Drives
//	err := json.Unmarshal(data, &r)
//	return r, err
//}
//func getVolumeInfo() map[string]int64 {
//	driveMap := make(map[string]int64)
//	ses, err := session.NewSession(&aws.Config{
//		Region: aws.String("us-east-1")},
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	svc := ssm.New(ses)
//	parameter := getParameter()
//	input := &ssm.GetParameterInput{
//		Name: aws.String(parameter),
//	}
//	response, err := svc.GetParameter(input)
//	if err != nil {
//		log.Println(err)
//	}
//	notEncodedParam := *response.Parameter.Value
//	decoded, err := base64.StdEncoding.DecodeString(notEncodedParam)
//	if err != nil {
//		log.Println(err)
//	}
//	scanner := bufio.NewScanner(strings.NewReader(string(decoded)))
//	for scanner.Scan() {
//		tmp := scanner.Text()
//		split := strings.Split(tmp, "=")
//		tmpSize, err := strconv.ParseInt(split[1], 10, 64)
//		if err != nil {
//			log.Println(err)
//		}
//		driveMap[split[0]] = tmpSize
//	}
//	return driveMap
//}
func getParameter() string {

	ses, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		log.Fatal(err)
	}
	svc := ec2.New(ses)

	input := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("resource-id"),
			Values: []*string{aws.String(getInstanceId())},
		}},
	}
	result, err := svc.DescribeTags(input)
	if err != nil {
		fmt.Println(err)
	}

	return result.String()
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

//func dirIsExist(volInfo map[string]int64) bool {
//	for key := range volInfo {
//		if _, err := os.Stat(key); os.IsNotExist(err) {
//			if _, err := exec.Command("mkdir", "-p", key).Output(); err != nil {
//				log.Println(err)
//				return false
//			}
//		}
//	}
//	return true
//}
//func serviceStatus(command string, services []string) {
//	for _, item := range services {
//
//		cmd := exec.Command("systemctl", "check", item)
//		out, err := cmd.CombinedOutput()
//		outString := string(out)
//		outString = strings.TrimSpace(outString)
//		outString = strings.Trim(outString, "\t \n")
//		if err != nil {
//			if _, ok := err.(*exec.ExitError); ok {
//			} else {
//				log.Printf("failed to run systemctl: %+v\n", err)
//				os.Exit(1)
//			}
//		}
//
//		if strings.EqualFold(outString, "active") && command == "stop" {
//			invokeStop := exec.Command("systemctl", command, item)
//			_, err2 := invokeStop.CombinedOutput()
//			if err2 != nil {
//				if exitErr2, ok := err2.(*exec.ExitError); ok {
//					log.Printf("systemctl finished with non-zero: %+v\n", exitErr2)
//				} else {
//					log.Printf("failed to run systemctl: %+v\n", err2)
//					os.Exit(1)
//				}
//			}
//		}
//		if strings.EqualFold(outString, "inactive") && command == "start" {
//			invokeStart := exec.Command("systemctl", command, item)
//			_, err3 := invokeStart.CombinedOutput()
//			if err3 != nil {
//				if exitErr3, ok := err3.(*exec.ExitError); ok {
//					if exitErr3.ExitCode() == 5 {
//						log.Printf("No such service in system: %+v\n", item)
//					} else {
//						log.Printf("Error type: %+v\n", exitErr3.ExitCode())
//					}
//				} else {
//					log.Printf("failed to run systemctl: %+v\n", err3)
//					os.Exit(1)
//				}
//			}
//		}
//	}
//}
//func lvcCreate(mPoint string, size int64) string {
//	points := strings.Split(mPoint, "/")
//	point := fmt.Sprintf("mountpoint_%s", points[len(points)-1])
//	newSize := fmt.Sprintf("%sG", strconv.FormatInt(size, 10))
//	_, err := exec.Command("lvcreate", "-n", point, "-L", newSize, "group1").Output()
//	if err != nil {
//		log.Fatal(err)
//	}
//	return point
//}
//func createFS(mPoint string) {
//	fullPoint := fmt.Sprintf("/dev/group1/%s", mPoint)
//	_, err := exec.Command("mkfs.xfs", fullPoint).Output()
//	if err != nil {
//		log.Fatal(err)
//	}
//}
//func createTempDir(tempDir string) {
//	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
//		err := os.MkdirAll(tempDir, 0700)
//		if err != nil {
//			log.Println(err)
//		}
//	}
//}
//func mountDrive(label string, directory string) {
//	_, err := exec.Command("mount", label, directory).Output()
//	if err != nil {
//		log.Println(err)
//	}
//}
//func copyData(src string, dst string) {
//	if _, err1 := exec.Command("rsync", "-raX", src+"/", dst+"/").Output(); err1 != nil {
//		log.Println(err1)
//	}
//}
//func moveData(src string, dst string) {
//	if _, err1 := exec.Command("mv", src, dst).Output(); err1 != nil {
//		log.Println(err1)
//	}
//}
//func unmountDrive(label string) {
//	_, err := exec.Command("umount", label).Output()
//	if err != nil {
//		log.Println(err)
//	}
//}
//func removeOldDir(directory string) {
//	if _, err1 := exec.Command("rm", "-Rf", directory).Output(); err1 != nil {
//		log.Println(err1)
//	}
//}
//func fstabConfig(label string, directory string, fsType string) {
//	uuid := getUUID(label)
//	var uuidString string
//	file, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0644)
//	if err != nil {
//		log.Println(err)
//	}
//	defer file.Close()
//	if directory == "/tmp" {
//		uuidString = fmt.Sprintf("\nUUID=%s %s %s nodev,nosuid,noexec 0 0", uuid, directory, fsType)
//	} else {
//		uuidString = fmt.Sprintf("\nUUID=%s %s %s defaults 0 0", uuid, directory, fsType)
//	}
//
//	if _, err := file.WriteString(uuidString); err != nil {
//		log.Println(err)
//	}
//
//}
//func getUUID(label string) string {
//	var uuid string
//	out, err := exec.Command("blkid", label, "--output", "export").Output()
//	if err != nil {
//		log.Println(err)
//	}
//	scanner := bufio.NewScanner(strings.NewReader(string(out)))
//	for scanner.Scan() {
//		UUID := strings.Split(scanner.Text(), "=")
//		if UUID[0] == "UUID" {
//			uuid = UUID[1]
//		}
//	}
//	return uuid
//}
//func compareVolumeAndDrives(drives map[string]int64, volumes map[string]int64, FileSystemType string) {
//	var resultSize int64
//	for _, val := range volumes {
//		resultSize = resultSize + val
//	}
//	for drv, size := range drives {
//		if size == resultSize {
//			pvCreate(drv)
//			pvGroupCreate(drv)
//		}
//	}
//	for mPoint, size := range volumes {
//
//		tempPointDirName := fmt.Sprintf("/temp%s", strings.Split(mPoint, "/")[len(strings.Split(mPoint, "/"))-1])
//
//		point := lvcCreate(mPoint, size-1)
//		createFS(point)
//		createTempDir(tempPointDirName)
//		fullLabel := fmt.Sprintf("/dev/mapper/group1-%s", point)
//		oldDir := fmt.Sprintf("%s.old", mPoint)
//		mountDrive(fullLabel, tempPointDirName)
//		copyData(mPoint, tempPointDirName)
//		moveData(mPoint, oldDir)
//		unmountDrive(fullLabel)
//		moveData(tempPointDirName, mPoint)
//		mountDrive(fullLabel, mPoint)
//		removeOldDir(oldDir)
//		fstabConfig(fullLabel, mPoint, FileSystemType)
//	}
//}
//func main() {
//	fsPtr := flag.String("f", "xfs", "File system type")
//	svcPtr := flag.String("s", "", "List of services for stop/start, enter inside quotes thru commas: \"ServiceName1,ServiceName2\"")
//	flag.Parse()
//	state := State{start: "start", stop: "stop"}
//	FileSystemType := getFs(*fsPtr)
//	services := prepareService(*svcPtr)
//	driveMap := getDrives()
//	volInfo := getVolumeInfo()
//	dirIsExist(volInfo)
//	serviceStatus(state.stop, services)
//	compareVolumeAndDrives(driveMap, volInfo, FileSystemType)
//	serviceStatus(state.start, services)
//}

func main() {
	asd := getParameter()
	fmt.Println(asd)
}
