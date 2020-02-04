# hlstail
hlstail is a simple CLI tool for tailing a specific variant of an HLS playlist

# Usage
```
NAME:
   hlstail - Query an HLS playlist and then tail the new segments of a selected variant

USAGE:
   [playlist]

VERSION:
   1.0.3

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --count value     The number of segments to display (default: 5)
   --interval value  The number of seconds to wait between updates (default: 3)
   --variant value   The number of the variant you'd like to use (default: 0)
   --help, -h        show help
   --version, -v     print the version
```

## Install 
```
go get -u github.com/moore0n/hlstail
go install github.com/moore0n/hlstail/...
```

## Try
```
hlstail --count 10 --interval 3 http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/sl.m3u8
```

## Build
If you so choose you can build a binary locally using the supplied build command.
```
$make mac
$make linux
```
