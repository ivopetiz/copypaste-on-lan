
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	//"runtime"
	"io/ioutil"
	"net/http"

	color "github.com/fatih/color"
	"github.com/ivopetiz/portscanner"
)

const (
	DefaultPort string = "9876"
	//has_config_file 		bool   = false

	// Info
	INFOexiting string = "Exiting"

	// Errors
	ERRwrongPortOrIP    string = "Given IP or port are incorrect"
	ERRlocalIPnotFound  string = "Local IP not found"
	ERRpaste            string = "Something went wrong with the paste"
	ERRdownloadingFile  string = "Can't download file"
	ERRipportPairDown   string = "Given ip:port not working"
	ERRnoCopyMachines   string = "No Copy machines available"
)

var (
	cFile   string = color.New(color.Bold, color.FgCyan).SprintFunc()("[ FILE ] ")
	cInfo   string = color.New(color.Bold, color.FgWhite).SprintFunc()("[ INFO ] ")
	cDownld string = color.New(color.Bold, color.FgGreen).SprintFunc()("[DOWNLD] ")
	cErr    string = color.New(color.Bold, color.FgRed).SprintFunc()("[ ERR  ] ")
)

type IPv4 [4]int

type InfoFile struct {
	text       bool
	listFiles []string
}

// ToString is used to pass an IP from IPv4 type to string
func (ip *IPv4) ToString() string {
	ipStringed := strconv.Itoa(ip[0])
	for i := 1; i < 4; i++ {
		str_i := strconv.Itoa(ip[i])
		ipStringed += "." + str_i
	}
	return ipStringed
}

// ToIPv4 is used to pass an IP from string to IPv4 type
func ToIPv4(ip string) IPv4 {
	var newIP IPv4
	ipS := strings.Split(ip, ".")
	for i, v := range ipS {
		newIP[i], _ = strconv.Atoi(v)
	}
	return newIP
}

// IsOK is used to check errors
func IsOK(err error, message string, fatal bool, debug bool) {
	if err != nil {
		log.Println(message)
		if debug {
			log.Println(cErr + err.Error())
		}
		if fatal {
			os.Exit(1)
		}
	}
}

// GetLocalIP is used to obtain machine's local IP
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ERRlocalIPnotFound
}

func SelectServer(servList []string) string {

	var input string
	//var valid_entry bool = false

	for {
		log.Println("Select between these copy servers:")
		for i, ip := range servList {
			log.Println(i, " → ", ip)
		}
		fmt.Scanln(&input)
		entry, _ := strconv.Atoi(input)
		if entry > 0 && entry < len(servList) {
			return servList[entry]
		} else {
			log.Println("\nSelect a valid entry!")
			time.Sleep(2 * time.Second)
			// limpar consola
			fmt.Print("\033[2J")
		}
		return "0"
	}
}

func Init(debug bool) {
	if debug {
		log.Println(cInfo + "Paste On Lan")
		log.Println(cInfo + "Debug Mode\n")
	} else {
		fmt.Println("Paste On Lan")
	}
}

// Paste is used to Download file(s) from Copy server
func Paste(copy_ip, port string, debug bool) (string, error) {
	var linkServer string = "http://" +
		copy_ip + ":" +
		port + "/"

	file, err := DownloadFile(linkServer + ".info.txt")
	inf, err := ParseIndex(file)

	if inf.text {
		return "", nil
	} else {
		for _, remoteFile := range inf.listFiles {
			if len(remoteFile) == 0 {
				break
			}
			file, err := DownloadFile(linkServer + remoteFile)
			if err != nil {
				log.Fatal(err)
			}
			if debug {
				log.Println(cDownld + file)
			} else {
				fmt.Println(file)
			}

		}
		return "", err
	}
}

// DownloadFile is used to download a file from an URL
func DownloadFile(url string) (string, error) { // Thanks PabloK
	// Create the file
	out, err := os.Create(path.Base(url))
	if err != nil {
		return "", err
	}
	defer out.Close()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return path.Base(url), nil
}

// ParseIndex is used to get
func ParseIndex(file string) (InfoFile, error) {
	// so da para files agora -> corrigir
	var text bool = false
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	files := strings.Split(string(content), "\n")
	info := InfoFile{text, files}
	return info, nil
}

// PortIsOpen checks if a IP:Port is open.
func PortIsOpen(ipAddr, port string, debug bool) bool {

	var openPort []string

	var port_ []string

	openPort = portscanner.PortScanner(portscanner.ToIPv4(ipAddr), append(port_, port))

	if len(openPort) == 1 {
		return true
	} else {
		return false
	}

}

// IPScan returns all IP addresses with copy server available.
func ServersScan(ip, port string, debug bool) []string {

	var ipRange []string
	port_slc := []string{port}

	var servers []string

	ipRange = append(ipRange, ip[:strings.LastIndex(ip, ".")]+".1-254")

	servers_map := portscanner.IPScanner(ipRange, port_slc, false)

	for ip, _ := range servers_map {
		servers = append(servers, ip.ToString())
	}

	return servers
}

func main() {

	var serverIP, serverPort string
	var ipList []string

	// P A R S E R
	port := flag.String("port", DefaultPort, "Port to Copy's server")
	ipAddr := flag.String("ip", "", "Copy server IP address")
	debug := flag.Bool("debug", false, "Get all significant info")

	flag.Parse()

	Init(*debug)

	if *ipAddr == "" {
		ipList = ServersScan(GetLocalIP(), *port, *debug)
		if len(ipList) < 1 {
			log.Println(cErr + ERRnoCopyMachines)
			log.Println(cInfo + INFOexiting)
			os.Exit(1)
		} else if len(ipList) == 1 {
			serverIP = ipList[0]
		} else if len(ipList) > 1 {
			serverIP = SelectServer(ipList)
		}
	} else {
		if *ipAddr == "localhost" {
			serverIP = "127.0.0.1"
		} else {
			serverIP = *ipAddr
		}
		if !PortIsOpen(serverIP, *port, *debug) {
			log.Println(cErr + ERRipportPairDown)
			log.Println(cInfo + INFOexiting)
			os.Exit(1)
		}
	}

	if *port != DefaultPort {
		serverPort = *port
	} else {
		serverPort = DefaultPort
	}

	if *debug {
		log.Println(cInfo + "IP Address: " + serverIP) //I
		log.Println(cInfo + "Port: " + serverPort)
	}

	Paste(serverIP, serverPort, *debug)
}
