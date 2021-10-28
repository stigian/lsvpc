#/bin/bash

mkdir -p packages/lsvpc-linux-amd64/
mkdir -p packages/lsvpc-linux-arm64/
mkdir -p packages/lsvpc-darwin-amd64/
mkdir -p packages/lsvpc-darwin-arm64/
mkdir -p packages/lsvpc-windows/

GOOS=linux GOARCH=amd64 go build  -o packages/lsvpc-linux-amd64/lsvpc .
GOOS=linux GOARCH=arm64 go build  -o packages/lsvpc-linux-arm64/lsvpc .
GOOS=darwin GOARCH=amd64 go build -o packages/lsvpc-darwin-amd64/lsvpc .
GOOS=darwin GOARCH=arm64 go build -o packages/lsvpc-darwin-arm64/lsvpc .
GOOS=windows go build  -o packages/lsvpc-windows/lsvpc.exe .

cd packages
tar cvzf lsvpc-linux-amd64.tgz lsvpc-linux-amd64/
tar cvzf lsvpc-linux-arm64.tgz lsvpc-linux-arm64/
tar cvzf lsvpc-darwin-amd64.tgz lsvpc-darwin-amd64/
tar cvzf lsvpc-darwin-arm64.tgz lsvpc-darwin-arm64/
zip -r lsvpc-windows.zip lsvpc-windows/
