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

	"github.com/AB0529/prompter"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"

	"gopkg.in/yaml.v2"
)

var (
	// ConfigPath the path to the config file
	ConfigPath = "./config.yml"
	// Sched the filled in the schedule type
	Sched *Schedule
	wg    sync.WaitGroup
)

// NewConfig responsible for creating a new config file with the details filled in
func NewConfig() *Config {
	// Make sure config file exists, if not create it
	if _, err := os.Stat(ConfigPath); err != nil {
		cfg := ConfigQuestions()
		cfgFileData, _ := yaml.Marshal(&cfg)

		ioutil.WriteFile(ConfigPath, cfgFileData, 0666)
		return cfg
	}
	// If it already exists, load it in
	var cfg *Config
	cfgFileData, _ := ioutil.ReadFile(ConfigPath)
	yaml.Unmarshal(cfgFileData, &cfg)

	return cfg
}

// NewSchedule responsible for creating a new schedule file with the details filled in
func NewSchedule(schedulePath string) *Schedule {
	// Make sure schedules directory exists, if not create it
	if _, err := os.Stat("./Schedules"); os.IsNotExist(err) {
		os.Mkdir("./Schedules", 0700)
	}

	// Make sure config file exists, if not create it
	if _, err := os.Stat(schedulePath); os.IsNotExist(err) {
		files, _ := ioutil.ReadDir("./Schedules")
		schedulePath := fmt.Sprintf("./Schedules/schedule_%d.yml", len(files)+1)
		Info("Creating Shedule" + prompter.Green.Sprint(" #"+fmt.Sprint(len(files)+1)))

		sc := ScheduleQuestions()
		scFileData, _ := yaml.Marshal(&sc)

		ioutil.WriteFile(schedulePath, scFileData, 0666)
		return sc
	}
	// If it already exists, load it in
	var sc *Schedule
	scFileData, _ := ioutil.ReadFile(schedulePath)
	yaml.Unmarshal(scFileData, &sc)

	return sc
}

// Init initializes the program
func Init() {
	currentSched := SelectSchedule(false)
	Sched = NewSchedule("./Schedules/" + currentSched)
	ClearScreen()
	fmt.Println("You have a total of " + prompter.Green.Sprint(len((*Sched))) + " classes")
	for _, c := range *Sched {
		// This is awful, but it works
		wds := make([]string, len(c.Weekdays))
		for _, w := range c.Weekdays {
			wds = append(wds, prompter.Yellow.Sprint(w.String()))
		}
		fmt.Printf("- %s at %s (%s)\n", prompter.Red.Sprint(c.Name), prompter.Cyan.Sprint(strings.Split(c.JoinTime.Format("2006-01-02 3:4pm"), " ")[1]), strings.TrimSpace(strings.Join(wds, " ")))
	}

	fmt.Println()
	Info("Using schedule " + prompter.Green.Sprint(currentSched))

	// Ask iniital question
	choice := InitalQuestions()
	switch choice {
	// ----------------------------------------
	// Handle schedule stuff
	case "Select or Create Schedule":
		ClearScreen()
		Sched = NewSchedule("./Schedules/" + SelectSchedule(true))
		Init()
		break
	// ----------------------------------------
	// Handle deleting a class
	case prompter.Red.Sprint("Delete Class"):
		ClearScreen()
		q := []interface{}{
			&prompter.Multiselect{
				Name:    "classToDel",
				Message: "Which class do you want to delete?",
				Options: GetAllClassNames(Sched),
			},
		}
		a := struct{ ClassToDel string }{}
		l := len((*Sched))
		prompter.Ask(&prompter.Prompt{Types: q}, &a)

		for i, class := range *Sched {
			if class.Name == a.ClassToDel {
				copy((*Sched)[i:], (*Sched)[i:])
				(*Sched)[l-1] = &Class{}
				(*Sched) = (*Sched)[:l-1]
				break
			}
		}

		// Delete file if no classes are present
		if len((*Sched)) <= 0 {
			os.Remove("./Schedules/" + currentSched)
			Init()
			break
		}

		// Write new schedule
		file, _ := yaml.Marshal(Sched)
		ioutil.WriteFile("./Schedules/"+currentSched, file, 0666)
		Init()
		break
	// ----------------------------------------
	// Handle editing a class
	case prompter.Cyan.Sprint("Edit Class"):
		ClearScreen()
		q := []interface{}{
			&prompter.Multiselect{
				Name:    "classToEdit",
				Message: "Which class do you want to delete?",
				Options: GetAllClassNames(Sched),
			},
		}
		a := struct{ ClassToEdit string }{}
		prompter.Ask(&prompter.Prompt{Types: q}, &a)

		for i, class := range *Sched {
			if class.Name == a.ClassToEdit {
				(*Sched)[i] = ClassQuestions()
				break
			}
		}
		// Write new schedule
		file, _ := yaml.Marshal(Sched)
		ioutil.WriteFile("./Schedules/"+currentSched, file, 0666)
		Init()
		break
	// ----------------------------------------
	// Handle adding a new class
	case prompter.Green.Sprint("Add Class"):
		ClearScreen()
		newClass := ClassQuestions()
		(*Sched) = append((*Sched), newClass)
		// Rewrite file
		file, _ := yaml.Marshal(Sched)
		ioutil.WriteFile("./Schedules/"+currentSched, file, 0666)
		Init()
		break
	// ----------------------------------------
	// Handle starting program
	case prompter.Purple.Sprint("Start Program"):
		StartProgram()
		break
	// ----------------------------------------
	default:
		StartProgram()
	}
}

// GetAllClassNames will return a slice of all class names in current schedule
func GetAllClassNames(schedule *Schedule) []string {
	names := []string{}

	for _, class := range *schedule {
		names = append(names, class.Name)
	}

	return names
}

// ClearScreen will clear the screen
func ClearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// ElementIsLocated returns a condition that checks if the element is found on page
func ElementIsLocated(by, selector string) selenium.Condition {
	return func(wd selenium.WebDriver) (bool, error) {
		_, err := wd.FindElement(by, selector)
		return err == nil, nil
	}
}

func contains(s []*time.Weekday, e time.Weekday) bool {
	for _, a := range s {
		if *a == e {
			return true
		}
	}
	return false
}

// StartMeet will start the Google Meet
func StartMeet(class *Class, config *Config) {
	Info("Joining " + prompter.Yellow.Sprint(class.Name))

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

	// Goto the Google meets
	wd.Get(class.MeetURL.String())
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
			Error("Could not join meet")
			return
		}
	}
	btn.Click()
	// Wait until class has been entered
	wd.Wait(ElementIsLocated(selenium.ByXPATH, "//span[@class='wnPUne N0PJ8e']"))

	// Number of people in the call
	prevNum := 0
	numReg, _ := regexp.Compile("\\d+")

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
				Info(prompter.Red.Sprint("Ended ") + "breakout room")
				time.Sleep(time.Minute)
				continue
			}

			// If it has "&born=Breakout&" we know it started
			if strings.Contains(curURL, "&born=Breakout") {
				// Ignore leave condition until breakout room has ended
				config.Leave = -1
				Info(prompter.Green.Sprint("Started ") + "breakout room")
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
        // Handle 'removed from meet'
        _, rmErr := wd.FindElement(selenium.ByXPATH, "//div[contains(text(), \"You've been removed from the meeting\")]")
        if rmErr == nil {
            // End the meet if been removed
            Info("Removed, leaving class " + prompter.Yellow.Sprint(class.Name))
            break
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
			Info("There are " + prompter.Yellow.Sprint(num) + " people in the call")
			prevNum = num
		}

		if config.Leave > num {
			Info("Leaving class " + prompter.Yellow.Sprint(class.Name))
			break
		}

		// Check every 3 seconds
		time.Sleep(time.Second * 3)
	}
	wg.Done()
}

// CheckSchedule will check the schedule for right time
func CheckSchedule(now time.Time, config *Config, schedule *Schedule) bool {
	// Get the current weekday
	weekday := now.Weekday()

	// Make sure there's a calss for the weekday and the time is right
	for _, class := range *schedule {
		if !contains(class.Weekdays, weekday) {
			continue
		}

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
	var config *Config

	file, _ := ioutil.ReadFile("./config.yml")
	yaml.Unmarshal(file, &config)

	ClearScreen()
	Info("Program started, will spring into action when class is ready!")

	weekday := time.Now().Weekday()

	// Make sure there's a calss for the weekday and the time is right
	for _, class := range *Sched {
		if !contains(class.Weekdays, weekday) {
			continue
		}

		fmt.Printf("- %s at %s\n", prompter.Red.Sprint(class.Name), prompter.Cyan.Sprint(strings.Split(class.JoinTime.Format("2006-01-02 3:4pm"), " ")[1]))
	}

	// Main loop
	for {
		now := time.Now()
		CheckSchedule(
			time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location()),
			config, Sched)
		// Wait 30 seconds
		time.Sleep(30 * time.Second)
	}
}
