#!/bin/sh

# empty `/dist` folder by deleting and re-creating it
rm -rf ./dist
mkdir -p ./dist

# create empty `.gitkeep`
touch ./dist/.gitkeep

# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o ./dist/HelperScripts_win-amd64.exe ./main.go

# Windows 32-bit
GOOS=windows GOARCH=386 go build -o ./dist/HelperScripts_win-386.exe ./main.go

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o ./dist/HelperScripts_linux-amd64 ./main.go

# Linux 32-bit
GOOS=linux GOARCH=386 go build -o ./dist/HelperScripts_linux-386 ./main.go

# # macOS 64-bit
# GOOS=darwin GOARCH=amd64 go build -o ./dist/HelperScripts-amd64-darwin ./main.go

# # macOS 32-bit
# GOOS=darwin GOARCH=386 go build -o ./dist/HelperScripts-386-darwin ./main.go
