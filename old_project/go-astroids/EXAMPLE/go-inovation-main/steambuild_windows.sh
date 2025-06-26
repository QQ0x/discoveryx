set -e

name=innovation2007
mkdir -p bin

# TODO: Should '-ldflags=-H=windowsgui' be added?
GOOS=windows GOARCH=386 go build -tags=steam -o bin/${name}_windows_386.exe .
GOOS=windows GOARCH=amd64 go build -tags=steam -o bin/${name}_windows_amd64.exe .
