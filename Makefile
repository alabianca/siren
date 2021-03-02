build:
	go build -o bin/darwin/siren main.go tarball.go

compile:
	echo "Compiling..."
	echo "Darwing - arm64 done"
	GOOS=linux GOARCH=arm GOARM=7 go build -o bin/linux_armv7/siren main.go tarball.go
	echo "Linux armv7 done"