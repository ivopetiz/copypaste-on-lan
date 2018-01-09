# Copypaste-on-lan

Copy/paste text and files between computers, along the network. Written in Golang.

## Installation

### Debian/Ubuntu

```bash
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" copy.go
sudo cp copy /usr/local/bin/
```

```bash
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" paste.go
sudo cp paste /usr/local/bin/
```

## COPY ON LAN

```bash
Usage of copy:
  -debug
    	Get all significant info
  -ip string
    	Destiny machine IP address
  -local
    	Intern transfer
  -port int
    	Port to Copy´s server (default 9876)
  -time int
    	Copy server window duration (in seconds) (default 300)
```


