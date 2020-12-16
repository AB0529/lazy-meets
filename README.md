# Lazy-Meets

Are you so lazy you can't even go to online classes? Well fret no more!

This uses Selenium with Firefox to automate login and going to your online classes on a fixed schedule.

You can leave this running in the background forever and it will automaticlly start a Firefox instance and join a class when the time is right.

For leaving a class, it will check for a threshold of users are left, once the amount of users are below this threshold, it will leave the meet by quitting the browser.

# Usage
Just run the binary files you need for your platform.

For Windows Users
-
Just Run the `lazy-meets.exe` file

For Linux and MacOS Users
- 
```sh
chmod +x lazy-meets-linux
./lazy-meets-linux
``` 
or
```sh
chmod +x lazy-meets-macos
./lazy-meets-macos
``` 

Or if you prefer running the files directly
```sh
go run main.go prompts.go update.go
``` 

# Requirements

* Go 1.15 or higher ([Install](https://golang.org/dl/))
* Selenium with Firefox Driver ([Install Guide](https://selenium-python.readthedocs.io/installation.html))

# Installation

## Install dependencies via go get
```sh
go get github.com/tebeka/selenium 
go get gopkg.in/yaml.v2 
```
