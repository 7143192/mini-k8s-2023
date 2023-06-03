package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"mini-k8s/pkg/config"
	defines2 "mini-k8s/pkg/defines"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ConvertContainerStateToStr(status defines2.ContainerState) string {
	state := status.State
	return ConvertIntToState(state)
}

func ConvertIntToState(state int) string {
	if state == defines2.Pending {
		return "Pending"
	} else {
		if state == defines2.Running {
			return "Running"
		} else {
			if state == defines2.Succeed {
				return "Succeed"
			} else {
				if state == defines2.Failed {
					return "Failed"
				} else {
					return "Unknown"
				}
			}
		}
	}
}

func GetReadyNumber(status defines2.ContainerState) int {
	if status.State != defines2.Failed && status.State != defines2.Unknown {
		return 1
	}
	return 0
}

func GetReadyNumberFromInt(state int) int {
	if state != defines2.Failed && state != defines2.Unknown {
		return 1
	}
	return 0
}

func GetTime(cur time.Time) uint64 {
	return uint64(cur.Second() + cur.Minute()*60 + (cur.Day()*24+cur.Hour())*3600)
}

func ConvertMemStrToInt(mem string) int {
	memType := mem[len(mem)-2:]
	memVal := mem[0 : len(mem)-2]
	val, _ := strconv.Atoi(memVal)
	switch memType {
	case "KB":
		return val << 10
	case "MB":
		return val << 20
	case "GB":
		return val << 30
	}
	return -1
}

func ConvertMemTotalToStr(total int) string {
	if total>>10 == 0 {
		return strconv.Itoa(total) + "B"
	} else {
		if total>>20 == 0 {
			return strconv.Itoa(total>>10) + "KB"
		} else {
			if total>>30 == 0 {
				return strconv.Itoa(total>>20) + "MB"
			} else {
				return strconv.Itoa(total>>30) + "GB"
			}
		}
	}
}

func GetTotalMem(pod *defines2.Pod) (string, string) {
	request := ""
	limit := ""
	requestTotal := 0
	limitTotal := 0
	for _, con := range pod.YamlPod.Spec.Containers {
		req := ConvertMemStrToInt(con.Resource.ResourceRequest.Memory)
		lim := ConvertMemStrToInt(con.Resource.ResourceLimit.Memory)
		requestTotal += req
		limitTotal += lim
	}
	request = ConvertMemTotalToStr(requestTotal)
	limit = ConvertMemTotalToStr(limitTotal)
	return request, limit
}

func GetTotalCPU(pod *defines2.Pod) (float64, float64) {
	request := 0
	limit := 0
	for _, con := range pod.YamlPod.Spec.Containers {
		reqCpu := strings.Replace(con.Resource.ResourceRequest.Cpu, ".", "", -1)
		req, _ := strconv.Atoi(reqCpu)
		request = req + request
		limCpu := strings.Replace(con.Resource.ResourceLimit.Cpu, ".", "", -1)
		lim, _ := strconv.Atoi(limCpu)
		limit = lim + limit
		// fmt.Printf("req cpu = %d, lim cpu = %d\n", request, limit)
	}
	reqF := float64(request) / 100.0
	limF := float64(limit) / 100.0
	return reqF, limF
}

func ParseCPUInfo(cpu string) int {
	cpuLen := len(cpu)
	cpuLimit := 0
	if cpu[cpuLen-1] == 'm' {
		// micro-CPU
		val := cpu[0 : cpuLen-1]
		gotVal, _ := strconv.Atoi(val)
		cpuLimit = gotVal * 1e6
	} else {
		// 0.XX
		if cpu[0] == '0' {
			idx := strings.Index(cpu, ".")
			intCpu := strings.Replace(cpu, ".", "", -1)
			gotIntCpu, _ := strconv.Atoi(intCpu)
			cpuLimit = gotIntCpu * 10e9
			backLen := cpuLen - idx - 1
			for i := 0; i < backLen; i++ {
				cpuLimit = cpuLimit / 10
			}
		} else {
			// a.XX (a != 0)
			newCPU := strings.Replace(cpu, ".", "", -1)
			gotCpu, _ := strconv.Atoi(newCPU)
			cpuLimit = gotCpu * 1e7
		}
	}
	return cpuLimit
}

func GetDescribeCPU(pod *defines2.Pod) (float64, float64) {
	request := 0
	limit := 0
	for _, con := range pod.YamlPod.Spec.Containers {
		limitCPU := con.Resource.ResourceLimit.Cpu
		reqCPU := con.Resource.ResourceRequest.Cpu
		limit = limit + ParseCPUInfo(limitCPU)
		request = request + ParseCPUInfo(reqCPU)
	}
	return float64(limit), float64(request)
}

func CheckStrInStrList(target string, list []string) bool {
	if len(list) == 0 {
		return false
	} else {
		for _, str := range list {
			if str == target {
				return true
			}
		}
	}
	return false
}

func ParseFlannelIP() (int, int) {
	ip := config.FlannelIP
	idx := strings.Index(ip, "/")
	ip = ip[0:idx]
	fmt.Printf("Flannel ip mask = %v\n", ip)
	parts := make([]int, 0)
	for {
		if strings.Contains(ip, ".") == false {
			break
		}
		idx := strings.Index(ip, ".")
		num := ip[0:idx]
		ip = ip[idx+1:]
		got, _ := strconv.Atoi(num)
		parts = append(parts, got)
	}
	return parts[0], parts[1]
}

func AllocateNewSubIp(a int, b int, nodeInfo *defines2.NodeInfo) string {
	got, _ := strconv.Atoi(nodeInfo.NodeData.NodeId)
	newIP := net.IPv4(byte(a), byte(b), byte(got), 0)
	res := newIP.String()
	fmt.Printf("new IP for node %v is = %v\n", nodeInfo.NodeData.NodeSpec.Metadata.Name, res)
	return res
}

func UnzipFile(zipFile, targetDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func(r *zip.ReadCloser) {
		_ = r.Close()
	}(r)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		defer func(rc io.ReadCloser) {
			_ = rc.Close()
		}(rc)

		targetPath := filepath.Join(targetDir, f.Name)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(targetPath, f.Mode())
			if err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
				return err
			}

			file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			defer func(file *os.File) {
				_ = file.Close()
			}(file)

			_, err = io.Copy(file, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
