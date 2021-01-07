package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
	"gopkg.in/yaml.v2"
)

var wg sync.WaitGroup

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
	_, err := os.Stat("./config.yml")
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
	_, err := os.Stat("./schedule.yml")
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

// ElementIsLocated returns a condition that checks if the element is found on page
func ElementIsLocated(by, selector string) selenium.Condition {
	return func(wd selenium.WebDriver) (bool, error) {
		_, err := wd.FindElement(by, selector)
		return err == nil, nil
	}
}

// StartMeet will start the Google Meet
func StartMeet(class Class, config Config) {
	Info("Joining " + Yellow(class.Name))

	// Setup Seleniuim
	selenium.SetDebug(false)
	port := 4444
	service, err := selenium.NewGeckoDriverService(Geckodriver, port)

	if err != nil {
		Error("Invalid geckodriver path")
		os.Exit(1)
	}
	defer service.Stop()

	// Firefox capabilites
	caps := selenium.Capabilities{}

	caps.AddFirefox(firefox.Capabilities{
		Binary: Firefox,
		Args:   []string{"--log-level=3", "--disable-infobars"},
		Prefs:  map[string]interface{}{"permissions.default.microphone": 1, "permissions.default.camera": 1},
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

	// title, _ := wd.Title()
	// If the title is not the inbox, the password was wrong
	// if !strings.Contains(strings.ToLower(title), "inbox") {
	// 	Error("Your password is invalid")
	// 	wd.Quit()
	// 	os.Exit(1)
	// }

	// Goto the Google meets
	wd.Get(class.URI.String())
	time.Sleep(5 * time.Second)
	// Find and click the mic off button and camera button
	btns, err := wd.FindElements(selenium.ByXPATH, "//div[@data-is-muted='false']")
	if err == nil && len(btns) >= 4 {
		btns[0].Click()
		btns[3].Click()
	}

	// Find Join button
	btn, err = wd.FindElement(selenium.ByXPATH, "//span[contains(text(), 'Join')]")
	if err != nil {
		btn, err = wd.FindElement(selenium.ByXPATH, "//span[contains(text(), 'Ask')]")
		if err != nil {
			Error("Could not join meet, trying again...")
			// Restart
			wd.Quit()
			service.Stop()
			StartMeet(class, config)
		}
	}
	btn.Click()
	// Wait until class has been entered
	wd.Wait(ElementIsLocated(selenium.ByXPATH, "//span[@class='wnPUne N0PJ8e']"))

	// Number of people in the call
	prevNum := 0
	numReg, _ := regexp.Compile("\\d+")

	wg.Add(2)
	go func() {
		defer wg.Done()

		// Keep track of the old leave condition to restore later
		oldLeaveCondition := config.Leave

		// Join and leave breakout rooms
		go func() {
			for {
				curURL, _ := wd.CurrentURL()

				// Find "Join" breakout popup and join it
				joinBtn, err := wd.FindElement(selenium.ByXPATH, "/html/body/div[1]/div[3]/div/div[2]/div[3]/div[2]/span/span")
				// No error means it's appeared and we can click join
				if err == nil {
					joinBtn.Click()
				}
				// Same thing as above but for the leave breakout popup
				leaveBtn, err := wd.FindElement(selenium.ByXPATH, "/html/body/div[1]/div[3]/div/div[2]/div[3]/div/span/span")
				if err == nil {
					leaveBtn.Click()
				}

				// If the current url has "&born&hs" we know it ended
				if strings.Contains(curURL, "&born&hs") {
					config.Leave = oldLeaveCondition
					Info(Red("Ended ") + "breakout room")
					time.Sleep(time.Second * 2)
					continue
				}

				// If it has "&born=Breakout&" we know it started
				if strings.Contains(curURL, "&born=Breakout") {
					// Ignore leave condition until breakout room has ended
					config.Leave = -1
					Info(Green("Started ") + "breakout room")
					time.Sleep(time.Second * 2)
					continue
				}

				time.Sleep(time.Millisecond * 500)
			}
		}()

		for {
			// Get current URL, and make sure it's not still joining breakout room
			curURL, _ := wd.CurrentURL()
			if strings.Contains(curURL, "&born=Breakout") {
				time.Sleep(time.Second * 3)
				continue
			}

			// Number of people in call
			numElem, err := wd.FindElement(selenium.ByXPATH, "//span[@class='wnPUne N0PJ8e']")
			if err != nil {
				time.Sleep(time.Second * 3)
				continue
			}
			numStr, _ := numElem.Text()

			// Look for other number element
			if numStr == "" {
				numElem, err = wd.FindElement(selenium.ByXPATH, "//span[@class='rua5Nb']")
				if err != nil {
					time.Sleep(time.Second * 3)
					continue
				}
				numStr, _ = numElem.Text()
			}

			// Convert to int
			num, _ := strconv.Atoi(numReg.FindString(numStr))

			if num != prevNum {
				Info("There are " + Yellow(num) + " people in the call")
				prevNum = num
			}

			if config.Leave > num {
				Info("Leaving class " + Yellow(class.Name))
				wd.Quit()
				break
			}

			// Check every 3 seconds
			time.Sleep(time.Second * 3)
		}
	}()

	wg.Wait()
	defer wg.Done()
}

func contains(s []time.Weekday, e time.Weekday) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// CheckSchedule will check the schedule for right time
func CheckSchedule(now time.Time, config Config, schedule Schedule) bool {
	// Get the current weekday
	weekday := now.Weekday()

	// Make sure there's a calss for the weekday and the time is right
	for _, class := range schedule.Classes {
		if !contains(class.Weekdays, weekday) {
			continue
		}

		jtH, jtM, _ := class.JoinTime.Clock()
		nH, nM, _ := now.Clock()

		// TODO:  properly add skip time to a timestamp
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

	file, _ := ioutil.ReadFile("./config.yml")
	yaml.Unmarshal(file, &config)
	file, _ = ioutil.ReadFile("./schedule.yml")
	yaml.Unmarshal(file, &schedule)

	Info("Program started, will spring into action when class is ready!")

	// Main loop
	for {
		now := time.Now()
		CheckSchedule(now, config, schedule)
		// Wait 30 seconds
		time.Sleep(30 * time.Second)
	}
}

// Init initalizes and runs the program
func Init() {
	Update()
	NewConfigFile()
	NewScheduleFile()
	InitialPrompt()
}

func main() {
	Init()
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
