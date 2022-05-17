clean:
	rm -f create_titleid

dev:
	go build

prod:
	GOOS=linux GOARCH=amd64 go build -o ../build/create_titleid.linux_x86
	GOOS=linux GOARCH=arm64 go build -o ../build/create_titleid.linux_arm
	GOOS=darwin GOARCH=amd64 go build -o ../build/create_titleid.mac_x86
	GOOS=darwin GOARCH=arm64 go build -o ../build/create_titleid.mac_arm
	GOOS=windows GOARCH=amd64 go build -o ../build/create_titleid.exe

