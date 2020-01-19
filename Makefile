.PHONY: binary

# TODO: Populate further builds for more platforms.

binary:
	GOARCH=amd64 GOOS=darwin go build -o ./hlstail ./cmd/hlstail/main.go
