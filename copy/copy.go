package main

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	color "github.com/fatih/color"
)

const (
	OS      string = runtime.GOOS
	VERSION string = "0.0.0"

	MIN_PORT       int  = 1025
	MAX_PORT       int  = 65536
	KEY_SIZE       int8 = 16
	RAND_MAX       int  = 999997
	MAX_to_present int  = 20

	ERR_low_port           string = " Port must be bigger than 1024"
	ERR_high_port          string = " Port must be smaller than 65536"
	ERR_invalid_port       string = " Port is not valid"
	ERR_dir_not_found      string = " Directory doesn't exist"
	ERR_file_not_found     string = " File not found"
	ERR_no_files           string = " Without files to copy"
	ERR_empty_dir          string = " Empty directory"
	ERR_file_error         string = " Can't upload "
	ERR_local_ip_not_found string = "Local IP not found"

	END_char_linux string = "/"
)

var (
	ip [4]byte

	// coloring system messages
	color_file string = color.New(color.Bold, color.FgCyan).SprintFunc()("[ FILE ] ")
	color_info string = color.New(color.Bold, color.FgWhite).SprintFunc()("[ INFO ] ")
	color_warn string = color.New(color.Bold, color.FgMagenta).SprintFunc()("[ WARN ] ")
	color_err  string = color.New(color.Bold, color.FgRed).SprintFunc()("[ ERR  ] ")
)

var Config struct {
	port_s    uint16
	timeout_s uint
	crypt_s   bool
	move_s    bool
}

func IsOK(err error, message string) error {
	if err != nil {
		log.Println(message)
		// if debug {
		// 	log.Fatal("[  ERROR ] ", err)
		// }
		return err
	}
	return nil
}

// Timeout is used to exit copy server.
func Timeout(expire int, debug bool) {
	time.Sleep(time.Duration(expire) * time.Second)
	if debug {
		log.Println(color_info + " Time is over")
		log.Println(color_info + " Exiting...")
	}
	os.Exit(0)
}

// GetLocalIP returns machine IP address.
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

// CreateFileList generates an index file where all downloadle files
// are declared. Also used in case of passing text instead of files.
func CreateFileList(dir string, files []string, debug bool) error {

	var all_files string

	for _, file := range files {
		all_files += file + "\n"
	}

	if debug {
		log.Println(color_info + "Generating " + dir + "/.info.txt")
	}
	d1 := []byte(all_files)
	err := ioutil.WriteFile(dir+"/.info.txt", d1, 0644)

	return err
}

// Encrypt original and compress files
func CopyFile(dstfile, srcfile string, key, iv []byte) error {
	r, err := os.Open(srcfile)
	if err != nil {
		return err
	}
	var w io.WriteCloser
	w, err = os.Create(dstfile)
	if err != nil {
		return err
	}

	w = gzip.NewWriter(w)

	defer w.Close()
	c, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	_, err = io.Copy(cipher.StreamWriter{S: cipher.NewOFB(c, iv), W: w}, r)
	return err
}

// Creates FileServer at port *port, serving directory dir.
func Copy(files []string, port int, debug bool) error {

	Bold := color.New(color.Bold).SprintFunc()
	// Create tmp dir
	rand.Seed(time.Now().UTC().UnixNano())
	tmp_dir := os.TempDir() + "/cp" + strconv.Itoa(rand.Intn(RAND_MAX))
	os.Mkdir(tmp_dir, 0755)

	defer RoomService(tmp_dir, debug)

	err := CreateFileList(tmp_dir, files, debug)
	if err != nil {
		log.Println(color_err + "Couldn't generate info file")
		log.Println(err)
	}

	if debug {
		if len(files) > MAX_to_present {
			log.Println(color_info + "Copying " + strconv.Itoa(len(files)) + " files")
		} else {
			log.Println(color_info + "Copying: ")
		}
	}
	for nl, file := range files {
		if debug && len(files) <= MAX_to_present {
			if nl < len(files)-1 {
				log.Println(color_file + " ├ " + file)
			} else {
				log.Println(color_file + " └ " + file)
			}
		}
		// move to temporary dir
		err := CopyFile(tmp_dir+"/"+file, file, make([]byte, 16), make([]byte, 16))
		if err != nil {
			log.Println(color_err + "  " + ERR_file_error + Bold(file))
			log.Println(err)
			continue
		}
	}
	if debug {
		log.Println(color_info + "Copy is ready!")
	} else {
		fmt.Println("Copy is ready\n")
	}
	panic(http.ListenAndServe(":"+strconv.Itoa(port), http.FileServer(http.Dir(tmp_dir))))
}

// RoomService erases the folder created to serve files with all files.
func RoomService(dir string, debug bool) error {
	if debug {
		log.Println(color_info + "Cleaning tmp folder")
	}
	err := os.RemoveAll(dir)
	fmt.Println(err)
	return err
}

func Init(debug bool) {
	if debug {
		log.Println(color_info + "Copy On Lan")
		log.Println(color_info + "Debug Mode\n")
	} else {
		fmt.Println("Copy On Lan\n")
	}
}

func main() {

	var files_to_copy []string
	var dir string
	var err error

	Bold := color.New(color.Bold).SprintFunc()

	// P A R S E R
	port := flag.Int("port", 9876, "Port to Copy's server")
	timeout := flag.Int("time", 300, "Copy server window duration (in seconds)")
	ip_dst := flag.String("ip", "", "Destiny machine IP address")
	debug := flag.Bool("debug", false, "Get all significant info")
	local := flag.Bool("local", false, "Intern transfer")
	flag.Parse()

	// Check if flags are valid.
	if *port < MIN_PORT {
		log.Println(color_err + ERR_low_port)
		os.Exit(1)
	}
	if *port > MAX_PORT {
		log.Println(color_err + ERR_high_port)
		os.Exit(1)
	}

	Init(*debug)

	// DEBUG
	if *debug {
		if *local {
			log.Println(color_info + "Intern Copy")
		} else {
			local_ip, err := GetLocalIP()
			if err != nil {
				log.Println(color_err + ERR_local_ip_not_found)
			} else {
				log.Println(color_info + "IP Address: " + local_ip)
			}
			log.Println(color_info + "Port: " + strconv.Itoa(*port))
			// IP dst defined
			if *ip_dst != "" {
				log.Println(color_info + "IP Address destination: " + Bold(*ip_dst))
			}
		}
	}

	// Check what is the working dir
	if len(flag.Args()) == 0 || (len(flag.Args()) == 1 && flag.Args()[0] == "*") {
		dir, err = os.Getwd()
		if err != nil {
			log.Fatal(color_err + ERR_dir_not_found)
			os.Exit(1)
		}
		// Check if first arg is a dir
	} else if fi, err := os.Stat(flag.Args()[0]); err != nil || fi.IsDir() {
		dir = flag.Args()[0]
		// Copy all files selected
	} else {
		for _, file := range flag.Args() {
			if _, err := os.Stat(file); err == nil {
				// path/to/whatever exists
				files_to_copy = append(files_to_copy, file)
			} else {
				log.Fatal(color_err + file + " -" + ERR_file_not_found)
			}
		}
		// Exit if there are no files to copy
	}

	// In case of a dir, get all files
	if dir != "" {
		wrk_dir, err := os.Open(dir)
		// Exit if dir is empty or other error
		files_to_copy, err = wrk_dir.Readdirnames(0)
		if err != nil {
			log.Fatal(color_err + ERR_no_files)
			os.Exit(1)
		}
		//_ = files_to_copy
	}

	if len(files_to_copy) == 0 {
		log.Fatal(color_err + ERR_no_files)
		os.Exit(1)
	}
	// call Timeout trigger

	if *timeout != 0 {
		if *debug {
			log.Println(color_info + "Expires at " + Bold(time.Now().Add(time.Duration(*timeout)*time.Second).Format("Jan 2 15:04:05")))
		}
		go Timeout(*timeout, *debug)
	} else {
		if *debug {
			log.Println(color_info + "Undefined timeout")
			log.Println(color_warn + "Need to press Ctrl+C to quit!")
		} else {
			fmt.Println("Press Ctrl+C to quit")
		}
	}

	Copy(files_to_copy, *port, *debug)
}
