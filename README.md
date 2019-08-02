# Copypaste-on-lan

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/4f04e44763804dc7b833948b3d59feda)](https://app.codacy.com/app/ivopetiz/copypaste-on-lan?utm_source=github.com&utm_medium=referral&utm_content=ivopetiz/copypaste-on-lan&utm_campaign=Badge_Grade_Settings)
[![Build Status](https://travis-ci.com/ivopetiz/copypaste-on-lan.svg?branch=master)](https://travis-ci.com/ivopetiz/copypaste-on-lan)

Copy/paste text and files between computers, along the network. Written in Golang.

## Installation

### Debian/Ubuntu

```bash
git clone https://github.com/ivopetiz/copypaste-on-lan.git
cd copypaste-on-lan/copy/
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" copy.go
sudo cp copy /usr/local/bin/gocopy
cd ../paste
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" paste.go
sudo cp paste /usr/local/bin/gopaste
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

## PASTE ON LAN
```bash
Usage of cpaste:
  -debug
    	Get all significant info
  -ip string
    	Copy server IP address
  -port string
    	Port to Copy's server (default "9876")
```
