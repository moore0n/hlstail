default: 
	@echo "Select a target of \"mac\" or \"linux\""
.PHONY: default

mac:
	@GOARCH=amd64 GOOS=darwin go build -o ./hlstail ./cmd/hlstail/main.go
linux:
	@GOARCH=amd64 GOOS=linux go build -o ./hlstail ./cmd/hlstail/main.go
