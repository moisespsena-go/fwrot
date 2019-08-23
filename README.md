# glogrotate

Advanged Go! Log Rotation library and CLI implementation.

## Installation

### Library
```bash
go get -u github.com/moisespsena-go/glogrotate
```

### CLI


```bash
go get -u github.com/moisespsena-go/glogrotate/glogrotate
```

Executable installed on $GOPATH/bin/glogrotate

#### Build from source

```bash
cd $GOPATH/src/github.com/moisespsena-go/glogrotate/glogrotate
```

##### Using Makefile

requires [goreleaser](https://goreleaser.com/).

```bash
make spt
```

See `./dist` directory to show all executables.

##### Default build

```bash
go build main.go
```

## Usage

### Library

See to [cli root source code](https://github.com/moisespsena-go/glogrotate/blob/master/glogrotate/cmd/root.go#L60) for example.

### CLI
    
```bash
glogrotate -h
```

    Starts file writer rotation reads IN and writes to OUT.
    
    EXAMPLES:
        NOTE: duration as minutely
    
        A. Basic example
            $ my_program | glogrotate -d m -o program.log
            $ my_program 2>&1 | glogrotate -d m -o program.log
        
        B. Input is STDIN, UDP, TCP and HTTP server
            main terminal:
                $ echo message from stdin | 
                    glogrotate -d m -o program.log -i +udp:localhost:5678+tcp:localhost:5679
    
            secondary terminal:
                a. send message from UDP client:
                    $ echo "message from UDP client" >/dev/udp/localhost/5678
    
                b. send message from TCP client:
                    $ echo "message from TCP client " >/dev/udp/localhost/5679
    
                c. send message from HTTP client:
                    $ curl -X POST -d "message from HTTP client" http://localhost:5680
    
        C. Input is STDIN and UDP server
            main terminal:
                $ (while true; do date; sleep 3; done) | 
                    glogrotate -d m -o program.log -i +udp:localhost:5678
    
            secondary terminal - send message from UDP client:
                $ echo "date from UDP client: "$(date) >/dev/udp/localhost/5678
    
    IN:
        Accept multiple inputs of STDIN, UDP and TCP servers.
        NOTE: Use plus char to join multiple values.
              The first plus char, combines with STDIN.
    
        SERVERS:
            UDP: udp:ADDR, udp4:ADDR, udp6:ADDR ('udp:' is alias of 'udp4:')
                Max message size is 1024 bytes.
    
                Example:
                    udp:localhost:5678
                    udp4:localhost:5678
                    udp:[::1]:5678
                    udp6:[::1]:5678
    
            TCP: tcp:ADDR ('tcp:' is alias of 'tcp4:')
                Example:
                    tcp:localhost:5679
                    tcp4:localhost:5679
                    tcp:[::1]:5679
                    tcp6:[::1]:5679
    
            HTTP: http:ADDR ('http:' is alias of 'http4:')
                - Accept HTTP POST method and copy all request body.
                - Accept WebSocket INPUT on "/" and copy all message body.
    
                Example:
                    http:localhost:5680
                    http4:localhost:5680
                    http:[::1]:5680
                    http6:[::1]:5680
        
        Examples:
            1. Multiple servers
                udp:localhost:5678+tcp:localhost:5679+http:localhost:5680
            2. Multiple servers with STDIN
                +udp:localhost:5678+tcp:localhost:5679+http:localhost:5680
    
    ENV VARIABLES:
        GLOGROTATE_OUT, GLOGROTATE_IN
        GLOGROTATE_HISTORY_DIR, GLOGROTATE_HISTORY_PATH, GLOGROTATE_HISTORY_COUNT 
        GLOGROTATE_DURATION, GLOGROTATE_MAX_SIZE  
        GLOGROTATE_DIR_MODE, GLOGROTATE_FILE_MODE
        GLOGROTATE_SILENT
    
        SET ENV variables to set default flag values.
    
        Usage example:
            Set duration as minutely and enable silent mode:
            $ export GLOGROTATE_DURATION=m
            $ export GLOGROTATE_SILENT=true
            
            run first program as background:
            $ my_first_program | glogrotate -d m -o first_program.log &
    
            run second program:
            $ my_second_program | glogrotate -d m -o second_program.log		
        
    TIME FORMAT:
        %Y - Year. (example: 2006)
        %M - Month with left zero pad. (examples: 01, 12)
        %D - Day with left zero pad. (examples: 01, 31)
        %h - Hour with left zero pad. (examples: 00, 05, 23)
        %m - Minute with left zero pad. (examples: 00, 05, 59)
        %s - Second with left zero pad. (examples: 00, 05, 59)
        %Z - Time Zone. If not set, uses UTC time. (examples: +0700, -0330)
    
    Usage:
      glogrotate [flags]
      glogrotate [command]
    
    Available Commands:
      follower    tail with follower OUT file
      help        Help about any command
      version     Show binary version
    
    Flags:
          --config string         config file
      -M, --dir-mode int          directory perms (default 0750)
      -d, --duration string       rotates every DURATION. Accepted values: Y - yearly, M - monthly, W - weekly, D - daily, h - hourly, m - minutely (default "M")
      -m, --file-mode int         file perms (default 0640)
      -h, --help                  help for glogrotate
      -C, --history-count int     Max history log count
      -c, --history-dir string    history root directory (default "OUT.history")
      -p, --history-path string   dynamic direcotry path inside ROOT DIR using TIME FORMAT (default "%Y/%M")
      -i, --in -                  the INPUT file. - (hyphen char) is STDIN. See INPUT section for details (default "-")
      -S, --max-size string       Forces rotation if current log size is greather then MAX_SIZE. Values in bytes. Examples: 100, 100K, 50M, 1G, 1T (default "50M")
      -o, --out string            the OUTPUT file
          --print                 print current config
          --silent                disable tee to STDOUT
    
    Use "glogrotate [command] --help" for more information about a command.

# Author
[Moises P. Sena](https://github.com/moisespsena)
