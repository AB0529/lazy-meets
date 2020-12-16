package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
	service, err := selenium.NewGeckoDriverService("/vendors/geckodriver", port)

	if err != nil {
		Error("Invalid geckodriver path")
		os.Exit(1)
	}
	defer service.Stop()

	// Firefox capabilites
	caps := selenium.Capabilities{}

	caps.AddFirefox(firefox.Capabilities{
		Binary: "/vendors/firefox",
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

	file, _ := ioutil.ReadFile(Basepath + "/config.yml")
	yaml.Unmarshal(file, &config)
	file, _ = ioutil.ReadFile(Basepath + "/schedule.yml")
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

// IsDirEmpty determins if a directory is empty
func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.Readdir(1)

	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// DownloadFile will download a file from a url
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Untar will untar a tarball
func Untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		fmt.Println("A")
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {

		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					fmt.Println("A")
					return err
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				fmt.Println("AA")
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				fmt.Println("AAA")
				return err
			}

			f.Close()
		}
	}
}

// Unzip unzip a zip file
func Unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// Update will check for updates and download latest updates
// TODO: Make it actually update instead of just downloading a single version
// TODO: Move this to a seperate file
func Update() {
	// TODO: None of this shit
	// FIREFOX_LINUX := ""
	// FIREFOX_WINDOWS := ""
	// FIREFOX_DARWIN := ""
	GeckoLinux := "https://github.com/mozilla/geckodriver/releases/download/v0.28.0/geckodriver-v0.28.0-linux64.tar.gz"
	GeckoWindows := "https://github.com/mozilla/geckodriver/releases/download/v0.28.0/geckodriver-v0.28.0-win64.zip"
	GeckoDarwin := "https://github.com/mozilla/geckodriver/releases/download/v0.28.0/geckodriver-v0.28.0-macos.tar.gz"

	// Check if directory is empty
	isEmpty, err := IsDirEmpty("vendors")
	if err != nil {
		panic(err)
	}
	// Download the geckodrivers and firefox for the platform
	if isEmpty {
		// Windows
		if runtime.GOOS == "windows" {
			Info("Downloading for platform " + Yellow("Windows"))

			if err := DownloadFile(Basepath+"/vendors/gecko.zip", GeckoWindows); err != nil {
				panic(err)
			}

			// Unzip the file
			Info("Unzipping...")
			_, err = Unzip(Basepath+"/vendors/gecko.zip", Basepath+"/vendors")

			if err != nil {
				panic(err)
			}
		}
		// Darwin
		if runtime.GOOS == "darwin" {
			Info("Downloading for platform " + Yellow("Darwin"))

			if err := DownloadFile("vendors/gecko.tar.gz", GeckoDarwin); err != nil {
				panic(err)
			}

			// Untar the tarball
			file, err := os.Open(Basepath + "/vendors/gecko.tar.gz")
			defer file.Close()
			Info("Unzipping...")
			err = Untar(Basepath+"/vendors", file)

			if err != nil {
				panic(err)
			}
		}
		// Linux
		if runtime.GOOS == "linux" {
			Info("Downloading for platform " + Yellow("Linux"))

			if err := DownloadFile(Basepath+"vendors/gecko.tar.gz", GeckoLinux); err != nil {
				panic(err)
			}

			// Untar the tarball
			file, err := os.Open(Basepath + "/vendors/gecko.tar.gz")
			defer file.Close()
			Info("Unzipping...")
			err = Untar(Basepath+"/vendors", file)

			if err != nil {
				panic(err)
			}
		}
	}

}

func main() {
	Update()
	NewConfigFile()
	NewScheduleFile()
	InitialPrompt()
}
