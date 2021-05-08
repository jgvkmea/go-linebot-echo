.PHONY: build
build:
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o ./build/linebot ./linebot/
	zip -j ./build/linebot ./build/linebot
