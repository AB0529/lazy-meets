# Lazy-Meets
![Go](https://github.com/AB0529/lazy-meets/workflows/Go/badge.svg)

Are you so lazy you can't even go to online classes? Well fret no more!

This uses Selenium with Firefox to automate a login and going to your online classes on a fixed schedule.

You can leave this running in the background forever, and it will automatically start a Firefox instance and join a class when the time is right.

For leaving a class, it will check for a threshold of users is left, once the amount of users is below this threshold, it will leave the meet by quitting the browser.

# Table of Contents
1. [Releases (Download)](https://github.com/AB0529/lazy-meets/releases/)
1. [Features](#Features)
1. [Usage](#Usage)
    - [Windows](#Windows)
    - [Mac/Linux](#Mac/Linux)
1. [Build](#build-from-source)
    - [Requirements](#Requirements)
    - [Dependencies](#Dependencies)
    - [Running](#Running)
    - [Compile](#Compile)

---
![showcase-img](https://raw.githubusercontent.com/AB0529/lazy-meets/master/Showcase-Image.png)

# Features
- Multiple schedule support.
- Automatically join a Google Meets session based on selected schedule. 
- Leave when less than *n* people are in the class.
- Joins Breakout rooms when prompted.
- Exits Breakout rooms when prompted.
- Can still manually talk in chat, unmute, etc.
# Usage
Download the correct file for your platform from the [releases page here](https://github.com/AB0529/lazy-meets/releases/).

To use this, it's fairly simple. Just follow the guide below for your platform.
## Windows
For Windows user's all you need to do is run the `lazy-meets.exe` file, and you're all set! 
## Mac/Linux
For Unix users, you need to first make sure the file is executable, then you can run it via the command line, like so:
```sh
# Make it executable
chmod +x ./lazy-meets-mac
# Run for Mac
./lazy-meets-mac
# Run for Linux 
./lazy-meets-linux
```

# Build from Source
There are binary releases available, but if you want to build from a source, here it is.
## Requirements
* Go version 1.15 or higher ([Download](https://golang.org/dl/))
## Dependencies
```sh
git clone https://github.com/AB0529/lazy-meets
cd lazy-meets/src
go get .
```
## Running
To run the program directly with Go:
```sh
go run main.go prompts.go types.go update.go util.go
```

# Compile
To compile into a binary:
```sh
# Releases for your current platform
go build
# Releases for Windows
GOOS=windows GOARCH=amd64 go build
# Releases for Mac
GOOS=darwin GOARCH=amd64 go build
# Releases for Linux
GOOS=linux GOARCH=amd64 go build
```
