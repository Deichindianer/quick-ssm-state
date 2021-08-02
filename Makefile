build: clean fmt
	mkdir -p ./bin
	cd cmd/qss; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags -static" -o ../../bin/qss_linux_x64; \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags -static" -o ../../bin/qss_darwin_x64; \
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags -static" -o ../../bin/qss_windows_x64.exe

fmt:
	go fmt ./...

clean:
	go clean
	rm -rf ./bin/
