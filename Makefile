.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o ./build/echo ./linebot/main.go
	zip -j ./build/echo ./build/echo
