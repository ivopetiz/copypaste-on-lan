
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
	DEFAULT_PORT string = "9876"
	//has_config_file 		bool   = false

	// Info
	INFO_exiting string = "Exiting"

	// Errors
	ERR_wrong_port_or_ip   string = "Given IP or port are incorrect"
	ERR_local_ip_not_found string = "Local IP not found"
	ERR_paste              string = "Something went wrong with the paste"
	ERR_downloading_file   string = "Can't download file"
	ERR_ipport_pair_down   string = "Given ip:port not working"
	ERR_no_copy_machines   string = "No Copy machines available"
)

var (
	c_file   string = color.New(color.Bold, color.FgCyan).SprintFunc()("[ FILE ] ")
	c_info   string = color.New(color.Bold, color.FgWhite).SprintFunc()("[ INFO ] ")
	c_downld string = color.New(color.Bold, color.FgGreen).SprintFunc()("[DOWNLD] ")
	c_err    string = color.New(color.Bold, color.FgRed).SprintFunc()("[ ERR  ] ")
)

type IPv4 [4]int

type Info_File struct {
	text       bool
	list_files []string
}

// ToString is used to pass an IP from IPv4 type to string
func (ip *IPv4) ToString() string {
	ip_stringed := strconv.Itoa(ip[0])
	for i := 1; i < 4; i++ {
		str_i := strconv.Itoa(ip[i])
		ip_stringed += "." + str_i
	}
	return ip_stringed
}

// ToIPv4 is used to pass an IP from string to IPv4 type
func ToIPv4(ip string) IPv4 {
	var new_ip IPv4
	ip_s := strings.Split(ip, ".")
	for i, v := range ip_s {
		new_ip[i], _ = strconv.Atoi(v)
	}
	return new_ip
}

// IsOK is used to check errors
func IsOK(err error, message string, fatal bool, debug bool) {
	if err != nil {
		log.Println(message)
		if debug {
			log.Println(c_err + err.Error())
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
	return ERR_local_ip_not_found
}

func SelectServer(serv_list []string) string {

	var input string
	//var valid_entry bool = false

	for {
		log.Println("Select between these copy servers:")
		for i, ip := range serv_list {
			log.Println(i, " → ", ip)
		}
		fmt.Scanln(&input)
		entry, _ := strconv.Atoi(input)
		if entry > 0 && entry < len(serv_list) {
			return serv_list[entry]
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
		log.Println(c_info + "Paste On Lan")
		log.Println(c_info + "Debug Mode\n")
	} else {
		fmt.Println("Paste On Lan\n")
	}
}

// Paste is used to Download file(s) from Copy server
func Paste(copy_ip, port string, debug bool) (string, error) {
	var link_server string = "http://" +
		copy_ip + ":" +
		port + "/"

	file, err := DownloadFile(link_server + ".info.txt")
	inf, err := ParseIndex(file)

	if inf.text {
		return "", nil
	} else {
		for _, remote_file := range inf.list_files {
			if len(remote_file) == 0 {
				break
			}
			file, err := DownloadFile(link_server + remote_file)
			if err != nil {
				log.Fatal(err)
			}
			if debug {
				log.Println(c_downld + file)
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
func ParseIndex(file string) (Info_File, error) {
	// so da para files agora -> corrigir
	var text bool = false
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	files := strings.Split(string(content), "\n")
	info := Info_File{text, files}
	return info, nil
}

// PortIsOpen checks if a IP:Port is open.
func PortIsOpen(ip_addr, port string, debug bool) bool {

	var open_port []string

	var port_ []string

	open_port = portscanner.PortScanner(portscanner.ToIPv4(ip_addr), append(port_, port))

	if len(open_port) == 1 {
		return true
	} else {
		return false
	}

}

// IPScan returns all IP addresses with copy server available.
func ServersScan(ip, port string, debug bool) []string {

	var ip_range []string
	port_slc := []string{port}

	var servers []string

	ip_range = append(ip_range, ip[:strings.LastIndex(ip, ".")]+".1-254")

	servers_map := portscanner.IPScanner(ip_range, port_slc, false)

	for ip, _ := range servers_map {
		servers = append(servers, ip.ToString())
	}

	return servers
}

func main() {

	var server_ip, server_port string
	var ip_list []string

	// P A R S E R
	port := flag.String("port", DEFAULT_PORT, "Port to Copy's server")
	ip_addr := flag.String("ip", "", "Copy server IP address")
	debug := flag.Bool("debug", false, "Get all significant info")

	flag.Parse()

	Init(*debug)

	if *ip_addr == "" {
		ip_list = ServersScan(GetLocalIP(), *port, *debug)
		if len(ip_list) < 1 {
			log.Println(c_err + ERR_no_copy_machines)
			log.Println(c_info + INFO_exiting)
			os.Exit(1)
		} else if len(ip_list) == 1 {
			server_ip = ip_list[0]
		} else if len(ip_list) > 1 {
			server_ip = SelectServer(ip_list)
		}
	} else {
		if *ip_addr == "localhost" {
			server_ip = "127.0.0.1"
		} else {
			server_ip = *ip_addr
		}
		if !PortIsOpen(server_ip, *port, *debug) {
			log.Println(c_err + ERR_ipport_pair_down)
			log.Println(c_info + INFO_exiting)
			os.Exit(1)
		}
	}

	if *port != DEFAULT_PORT {
		server_port = *port
	} else {
		server_port = DEFAULT_PORT
	}

	if *debug {
		log.Println(c_info + "IP Address: " + server_ip) //I
		log.Println(c_info + "Port: " + server_port)
	}

	Paste(server_ip, server_port, *debug)
}
