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

	MinPort       int  = 1025
	MaxPort       int  = 65536
	KeySize       int8 = 16
	RandMax       int  = 999997
	MaxToPresent  int  = 20

	ERRlowPort           string = " Port must be bigger than 1024"
	ERRhighPort          string = " Port must be smaller than 65536"
	ERRinvalidPort       string = " Port is not valid"
	ERRdirNotFound       string = " Directory doesn't exist"
	ERRfileNotFound      string = " File not found"
	ERRnoFile            string = " Without files to copy"
	ERRemptyDir          string = " Empty directory"
	ERRfileError         string = " Can't upload "
	ERRlocalIPnotFound   string = "Local IP not found"

	ENDcharLinux string = "/"
)

var (
	ip [4]byte

	// coloring system messages
	colorFile string = color.New(color.Bold, color.FgCyan).SprintFunc()("[ FILE ] ")
	colorInfo string = color.New(color.Bold, color.FgWhite).SprintFunc()("[ INFO ] ")
	colorWarn string = color.New(color.Bold, color.FgMagenta).SprintFunc()("[ WARN ] ")
	colorErr  string = color.New(color.Bold, color.FgRed).SprintFunc()("[ ERR  ] ")
)

var Config struct {
	portS    uint16
	timeoutS uint
	cryptS   bool
	moveS    bool
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
		log.Println(colorInfo + " Time is over")
		log.Println(colorInfo + " Exiting...")
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

	var allFiles string

	for _, file := range files {
		allFiles += file + "\n"
	}

	if debug {
		log.Println(colorInfo + "Generating " + dir + "/.info.txt")
	}
	d1 := []byte(allFiles)
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
	tmpDir := os.TempDir() + "/cp" + strconv.Itoa(rand.Intn(RandMax))
	os.Mkdir(tmpDir, 0755)

	defer RoomService(tmpDir, debug)

	err := CreateFileList(tmpDir, files, debug)
	if err != nil {
		log.Println(colorErr + "Couldn't generate info file")
		log.Println(err)
	}

	if debug {
		if len(files) > MaxToPresent  {
			log.Println(colorInfo + "Copying " + strconv.Itoa(len(files)) + " files")
		} else {
			log.Println(colorInfo + "Copying: ")
		}
	}
	for nl, file := range files {
		if debug && len(files) <= MaxToPresent  {
			if nl < len(files)-1 {
				log.Println(colorFile + " ├ " + file)
			} else {
				log.Println(colorFile + " └ " + file)
			}
		}
		// move to temporary dir
		err := CopyFile(tmpDir+"/"+file, file, make([]byte, 16), make([]byte, 16))
		if err != nil {
			log.Println(colorErr + "  " + ERRfileError + Bold(file))
			log.Println(err)
			continue
		}
	}
	if debug {
		log.Println(colorInfo + "Copy is ready!")
	} else {
		fmt.Println("Copy is ready\n")
	}
	panic(http.ListenAndServe(":"+strconv.Itoa(port), http.FileServer(http.Dir(tmpDir))))
}

// RoomService erases the folder created to serve files with all files.
func RoomService(dir string, debug bool) error {
	if debug {
		log.Println(colorInfo + "Cleaning tmp folder")
	}
	err := os.RemoveAll(dir)
	fmt.Println(err)
	return err
}

func Init(debug bool) {
	if debug {
		log.Println(colorInfo + "Copy On Lan")
		log.Println(colorInfo + "Debug Mode\n")
	} else {
		fmt.Println("Copy On Lan\n")
	}
}

func main() {

	var filesToCopy []string
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
	if *port < MinPort {
		log.Println(colorErr + ERRlowPort)
		os.Exit(1)
	}
	if *port > MaxPort {
		log.Println(colorErr + ERRhighPort)
		os.Exit(1)
	}

	Init(*debug)

	// DEBUG
	if *debug {
		if *local {
			log.Println(colorInfo + "Intern Copy")
		} else {
			local_ip, err := GetLocalIP()
			if err != nil {
				log.Println(colorErr + ERRlocalIPnotFound)
			} else {
				log.Println(colorInfo + "IP Address: " + local_ip)
			}
			log.Println(colorInfo + "Port: " + strconv.Itoa(*port))
			// IP dst defined
			if *ip_dst != "" {
				log.Println(colorInfo + "IP Address destination: " + Bold(*ip_dst))
			}
		}
	}

	// Check what is the working dir
	if len(flag.Args()) == 0 || (len(flag.Args()) == 1 && flag.Args()[0] == "*") {
		dir, err = os.Getwd()
		if err != nil {
			log.Fatal(colorErr + ERRdirNotFound)
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
				filesToCopy = append(filesToCopy, file)
			} else {
		 		log.Fatal(colorErr + file + " -" + ERRfileNotFound)
		 	}
		}
		// Exit if there are no files to copy
	}

	// In case of a dir, get all files
	if dir != "" {
		wrk_dir, err := os.Open(dir)
		// Exit if dir is empty or other error
		filesToCopy, err = wrk_dir.Readdirnames(0)
		if err != nil {
			log.Fatal(colorErr + ERRnoFile )
			os.Exit(1)
		}
		//_ = filesToCopy
	}

	if len(filesToCopy) == 0 {
		log.Fatal(colorErr + ERRnoFile )
		os.Exit(1)
	}
	// call Timeout trigger

	if *timeout != 0 {
		if *debug {
			log.Println(colorInfo + "Expires at " + \
				Bold(time.Now().Add(time.Duration(*timeout)*time.Second).Format("Jan 2 15:04:05")))
		}
		go Timeout(*timeout, *debug)
	} else {
		if *debug {
			log.Println(colorInfo + "Undefined timeout")
			log.Println(colorWarn + "Need to press Ctrl+C to quit!")
		} else {
			fmt.Println("Press Ctrl+C to quit")
		}
	}

	Copy(filesToCopy, *port, *debug)
}
