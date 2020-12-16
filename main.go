package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	wg         sync.WaitGroup
)

// GetBasePath gets the Basepath
func GetBasePath() string {
	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fmt.Printf(fmt.Sprintf(`Error on getting the go-kit base path: %s - %s`, err.Error(), string(cmdOut)))
		os.Exit(1)
	}
	return strings.TrimSpace(string(cmdOut))
}

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
	_, err := os.Stat(GetBasePath() + "/config.yml")
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
	_, err := os.Stat(GetBasePath() + "/schedule.yml")
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
	btn, err = wd.FindElement(selenium.ByXPATH, "//span[contains(text(), 'Join')]")
	if err != nil {
		btn, _ = wd.FindElement(selenium.ByXPATH, "//span[contains(text(), 'Ask')]")
	}
	btn.Click()
	// Wait until class has been entered
	wd.WaitWithTimeout(ElementIsLocated(selenium.ByID, "wnPUne N0PJ8e"), time.Second*10)

	// Number of people in the call
	prevNum := 0
	numReg, _ := regexp.Compile("\\d+")

	wg.Add(2)
	go func() {
		defer wg.Done()

		for {
			// Number of people in call
			numElem, _ := wd.FindElement(selenium.ByXPATH, "//span[@class='wnPUne N0PJ8e']")
			numStr, _ := numElem.Text()

			// Look for other number element
			if numStr == "" {
				numElem, _ = wd.FindElement(selenium.ByXPATH, "//span[@class='rua5Nb']")
				numStr, _ = numElem.Text()
			}

			// Convert to int
			num, _ := strconv.Atoi(numReg.FindString(numStr))

			if num != prevNum {
				Info("There are " + Yellow(num) + " people in the call")
				prevNum = num
			}
			if num < config.Leave {
				Info("Leaving class " + Yellow(class.Name))
				wd.Quit()
				break
			}

			// Check every 5 seconds
			time.Sleep(time.Second * 5)
		}
	}()

	wg.Wait()
	defer wg.Done()
}

// CheckSchedule will check the schedule for right time
func CheckSchedule(now time.Time, config Config, schedule Schedule) bool {
	// Get the current weekday
	weekday := now.Weekday()

	// Make sure there's a calss for the weekday and the time is right
	for _, class := range schedule.Classes {
		for _, wd := range class.Weekdays {
			if weekday != wd {
				continue
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

	file, _ := ioutil.ReadFile(GetBasePath() + "/config.yml")
	yaml.Unmarshal(file, &config)
	file, _ = ioutil.ReadFile(GetBasePath() + "/schedule.yml")
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

func main() {
	Update()
	NewConfigFile()
	NewScheduleFile()
	InitialPrompt()
}
