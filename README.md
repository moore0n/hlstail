# hlstail
hlstail is a simple CLI tool for tailing a specific variant of an HLS playlist

# Usage
```
NAME:
   hlstail - Query an HLS playlist and then tail the new segments of a variant

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --playlist value  The url of the master playlist
   --count value     The number of segments to display (default: 5)
   --help, -h        show help
   --version, -v     print the version
```

## Install 
```
go get -u github.com/moore0n/hlstail/...
go install github.com/moore0n/hlstail/...
```

## Build
If you so choose you can build a binary locally using the supplied build command.
```
make binary
```
