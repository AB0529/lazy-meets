# Lazy-Meets

Are you so lazy you can't even go to online classes? Well fret no more!

This uses Selenium with Firefox to automate login and going to your online classes on a fixed schedule.

You can leave this running in the background forever and it will automaticlly start a Firefox instance and join a class when the time is right.

For leaving a class, it will check for a threshold of users are left, once the amount of users are below this threshold, it will leave the meet by quitting the browser.

# Table of Contents
- [Releases](https://github.com/AB0529/lazy-meets/releases/)
- [Usage](#Usage)
    - [Windows](#Windows)
    - [Mac/Linux](#Mac/Linux)
- [Build](#Build)
    - [Requirements](#Requirements)
    - [Dependencies](#Dependencies)
    - [Running](#Running)

# Usage
To use this, it's fairly simple. Just follow the guide for your platform.
## Windows
For Windows user's all you need to do is run the `lazy-meets.exe` file and you're all set!
## Mac/Linux
For Unix users, you need to first make sure the file is exectuable, then you can run it via the command line, like so:
```sh
# Make it exectuable
chmod +x ./lazy-meets-mac
# Run it
./lazy-meets-mac
```

# Build from Source
There are binary releases available, but if you want to build from source, here it is.
## Requirements
* Go version 1.15 or higher ([Download](https://golang.org/dl/))
## Dependencies
This only has a two dependencies `selenium` and `yaml`, to install them:
```sh
go get github.com/tebeka/selenium 
go get gopkg.in/yaml.v2 
```
## Running
To run the program directly with Go:
```sh
go run main.go prompts.go update.go
```