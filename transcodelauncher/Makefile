clean:
	rm -f transcodelauncher

dev:
	go build

prod:
	GOOS=linux GOARCH=amd64 go build -o ../build/transcodelauncher.linux_x86
	GOOS=linux GOARCH=arm64 go build -o ../build/transcodelauncher.linux_arm
	GOOS=darwin GOARCH=amd64 go build -o ../build/transcodelauncher.mac_x86
	GOOS=darwin GOARCH=arm64 go build -o ../build/transcodelauncher.mac_arm
	GOOS=windows GOARCH=amd64 go build -o ../build/transcodelauncher.exe

