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

- copy / server side
- paste / client side
- parse opts
    :copy
        - debug
        - timeout
        - file
        - dir
        - move
        - pass
        - network
        - internet

    :paste
        - debug
        - ip
        - port

 - copy cria um servidor de ficheiros com ssl numa porta especifica
 - paste precisa de nmap para a porta especifica para toda a gama de rede em q está
 - pensar em cifrar dados

 - tentar copiar os dados para uma pasta tmp e depois servir esse diretorio
 - pensar na parte do texto

 - só ipv4 por agora

