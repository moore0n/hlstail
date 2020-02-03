.PHONY: binary-mac, binary-linux

binary-mac:
	GOARCH=amd64 GOOS=darwin go build -o ./hlstail ./cmd/hlstail/main.go

binary-linux:
	GOARCH=amd64 GOOS=linux go build -o ./hlstail ./cmd/hlstail/main.go

	// TODO: Add windows build?