package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
	"gopkg.in/yaml.v2"
)

var (
	_, b, _, _ = runtime.Caller(0)
	// Basepath the root directory
	Basepath = filepath.Dir(b)
	wg       sync.WaitGroup
)

// NewConfig will ask for promtps for the config
func NewConfig() Config {
	var cfg Config

	fmt.Println(Green("Enter details for the confiuration"))
	cfg = ConfigQuestions()

	return cfg
}

// NewSchedule will ask for prompts for the schedule
func NewSchedule() Schedule {
	var schedule Schedule

	fmt.Println(Green("Enter details for the schedule"))

	count := 0
	fmt.Println(Red("> ") + Teal("How many classes? "))
	fmt.Scanln(&count)

	if count <= 0 {
		Error("Count has to be above 0")
		os.Exit(1)
	}

	for i := 0; i < count; i++ {
		fmt.Println(Green(fmt.Sprintf("Enter details for %s", Yellow(fmt.Sprintf("class %d", i+1)))))
		class := ClassQuestions()
		schedule.Classes = append(schedule.Classes, class)
	}

	return schedule
}

// NewConfigFile will create a new configuration file
func NewConfigFile() {
	// Make sure config file exists, if not create it
	_, err := os.Stat(Basepath + "/config.yml")
	if err != nil {
		Info("No config found, creating new config...")
		cfg := NewConfig()
		file, _ := yaml.Marshal(cfg)
		ioutil.WriteFile("config.yml", file, 0644)
	}
}

// NewScheduleFile will create a new schedule file
func NewScheduleFile() {
	// Make sure schedule file exists, if not create it
	_, err := os.Stat(Basepath + "/schedule.yml")
	if err != nil {
		Info("No schedule found, creating new schedule...")
		schedule := NewSchedule()
		file, _ := yaml.Marshal(schedule)
		ioutil.WriteFile("schedule.yml", file, 0644)
	}
}

// InitialPrompt the inital prompt asked when starting the program
func InitialPrompt() {
	IniitalQuesitons()
}

// StartMeet will start the Google Meet
func StartMeet(class Class, config Config) {
	Info("Joining " + Yellow(class.Name))

	// Setup Seleniuim
	selenium.SetDebug(false)
	port := 4444
	service, err := selenium.NewGeckoDriverService(config.Geckodriver, port)

	if err != nil {
		Error("Invalid geckodriver path " + Yellow(config.Geckodriver))
		os.Exit(1)
	}
	defer service.Stop()

	// Firefox capabilites
	caps := selenium.Capabilities{}

	caps.AddFirefox(firefox.Capabilities{
		Binary: "/usr/bin/firefox",
		Args:   []string{"--log-level=3", "--disable-infobars"},
		Prefs:  map[string]interface{}{"permissions.default.microphone": 2, "permissions.default.camera": 2},
	})

	// Firefox web driver
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer wd.Quit()

	// Login to Google
	wd.Get("https://accounts.google.com/ServiceLogin?service=mail&continue=https://mail.google.com/mail/#identifier")
	// Fill in email
	email, _ := wd.FindElement(selenium.ByID, "identifierId")
	email.SendKeys(config.Email)
	btn, _ := wd.FindElement(selenium.ByID, "identifierNext")
	btn.Click()
	time.Sleep(3 * time.Second)
	// Fill in password
	password, _ := wd.FindElement(selenium.ByXPATH, "//input[@class='whsOnd zHQkBf']")
	password.SendKeys(config.Password)
	btn, err = wd.FindElement(selenium.ByID, "passwordNext")
	if err != nil {
		Error("Your email is invalid")
		wd.Quit()
		os.Exit(1)
	}
	btn.Click()
	time.Sleep(3 * time.Second)

	title, _ := wd.Title()
	// If the title is not the inbox, the password was wrong
	if !strings.Contains(strings.ToLower(title), "inbox") {
		Error("Your password is invalid")
		wd.Quit()
		os.Exit(1)
	}

	// Goto the Google meets
	wd.Get(class.URI.String())
	time.Sleep(5 * time.Second)
	// Find Join button
	btn, _ = wd.FindElement(selenium.ByXPATH, "//span[contains(text(), 'Join')]")
	btn.Click()
	time.Sleep(5 * time.Second)

	// Number of people in the call
	prevNum := 0
	numReg, _ := regexp.Compile("\\d+")

	wg.Add(2)
	go func() {
		for {
			// Number of people in call
			numElem, _ := wd.FindElement(selenium.ByXPATH, "//span[@class='wnPUne N0PJ8e']")
			numStr, _ := numElem.Text()

			// Look for other number element
			if numStr == "" {
				numElem, _ = wd.FindElement(selenium.ByXPATH, "//span[@class='rua5Nb']")
				numStr, _ = numElem.Text()
			}

			// Convert to number
			num, _ := strconv.Atoi(numReg.FindString(numStr))
			fmt.Println(num)

			if num != prevNum {
				Info("There are " + Yellow(num) + " people in the call")
				prevNum = num
			}
			if num < config.Leave {
				Info("Leaving class " + Yellow(class.Name))
				break
			}

			// Check every 10 seconds
			time.Sleep(time.Second * 5)
		}
		defer wg.Done()
	}()

	wg.Wait()
	defer wg.Done()
}

// CheckSchedule will check the schedule for right time
func CheckSchedule(now time.Time, config Config, schedule Schedule) bool {
	// Get the current weekday
	weekday := now.Weekday()
	// skipTime := time.Duration(config.Skip) * time.Minute

	// Make sure there's a calss for the weekday and the time is right
	for _, class := range schedule.Classes {
		for _, wd := range class.Weekdays {
			if weekday == wd {
				break
			}
		}

		// TODO: Use proper time compare instead of this
		jtH, jtM, _ := class.JoinTime.Clock()
		nH, nM, _ := now.Clock()

		if jtH == nH && nM-jtM <= config.Skip && nM-jtM >= 0 {
			wg.Add(1)
			go StartMeet(class, config)
			wg.Wait()
		}

	}
	return true
}

// StartProgram starts the main program
func StartProgram() {
	// Load the config and the schedule
	var config Config
	var schedule Schedule

	file, _ := ioutil.ReadFile(Basepath + "/config.yml")
	yaml.Unmarshal(file, &config)
	file, _ = ioutil.ReadFile(Basepath + "/schedule.yml")
	yaml.Unmarshal(file, &schedule)

	Info("Program started, will spring into action when class is ready!")

	// Main loop
	for {
		now := time.Now()
		CheckSchedule(now, config, schedule)
		// Wait a minute
		time.Sleep(5 * time.Second)
	}
}

func main() {
	NewConfigFile()
	NewScheduleFile()
	InitialPrompt()
}
