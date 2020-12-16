package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config the structure for the config
type Config struct {
	Leave       int
	Skip        int
	Email       string
	Password    string
	Geckodriver string
	Firefox     string
}

// Class the structure for a class
type Class struct {
	Weekdays []time.Weekday
	Name     string
	URI      *url.URL
	JoinTime time.Time
}

// Schedule a slice of classes
type Schedule struct {
	Classes []Class
}

var (
	// Black color black
	Black = Color("\033[1;30m%s\033[0m")
	// Red color red
	Red = Color("\033[1;31m%s\033[0m")
	// Green color green
	Green = Color("\033[1;32m%s\033[0m")
	// Yellow color yellow
	Yellow = Color("\033[1;33m%s\033[0m")
	// Purple color purple
	Purple = Color("\033[1;34m%s\033[0m")
	// Magenta color magenta
	Magenta = Color("\033[1;35m%s\033[0m")
	// Teal color teal
	Teal = Color("\033[1;36m%s\033[0m")
	// White color white
	White = Color("\033[1;37m%s\033[0m")
)

// Color colorizes strings
func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

// Info info print
func Info(msg ...interface{}) {
	fmt.Printf("[%s] - %s\n", Yellow("INFO"), msg)
}

// Error error print
func Error(msg ...interface{}) {
	fmt.Printf("[%s] - %s\n", Red("ERROR"), msg)
}

// ParseWeekday makes sure a weekday is valid
func ParseWeekday(s string) (time.Weekday, error) {
	var weekdays = map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}

	if d, ok := weekdays[strings.ToLower(s)]; ok {
		return d, nil
	}

	return time.Sunday, fmt.Errorf("invalid weekday '%s'", s)
}

// IniitalQuesitons prompts questions used for inital prompt
func IniitalQuesitons() {
	var schedule Schedule
	questions := []string{
		Magenta("What do you want to do?\n") + Green("1. ") + Purple("Start Program\n") + Green("2. ") + Teal("Edit Classes\n") + Green("3. ") + Red("Add Classes\n") + Green("4. ") + Yellow("Delete Classes"),
	}
	scheduleQuestions := []string{
		Magenta("Which class do you want to edit?"),
	}
	scanner := bufio.NewScanner(os.Stdin)
	// Get the schedule
	file, _ := ioutil.ReadFile(Basepath + "/schedule.yml")
	yaml.Unmarshal(file, &schedule)

	for _, q := range questions {
		fmt.Println(Yellow("> ") + q)
		scanner.Scan()
		t := scanner.Text()

		switch t {
		case "1":
			StartProgram()
			break
		// Edit
		case "2":
			// The class to edit
			for _, q2 := range scheduleQuestions {
				fmt.Println(Yellow("> ") + q2)
				for i, classes := range schedule.Classes {
					fmt.Printf("%s. %s\n", Green(i+1), Purple(classes.Name))
				}

				scanner.Scan()
				t, err := strconv.Atoi(scanner.Text())
				if err != nil {
					Error(err.Error())
					os.Exit(1)
				}
				// Update the class
				schedule.Classes[t-1] = ClassQuestions()
			}
			// Rewrite file
			file, _ = yaml.Marshal(schedule)
			ioutil.WriteFile("schedule.yml", file, 0644)
			break
		// Add
		case "3":
			newClass := ClassQuestions()
			schedule.Classes = append(schedule.Classes, newClass)
			// Rewrite file
			file, _ = yaml.Marshal(schedule)
			ioutil.WriteFile("schedule.yml", file, 0644)
			break
		// Delete
		case "4":
			// The class to delete
			for _, q2 := range scheduleQuestions {
				fmt.Println(Yellow("> ") + q2)
				for i, classes := range schedule.Classes {
					fmt.Printf("%s. %s\n", Green(i+1), Purple(classes.Name))
				}

				scanner.Scan()
				t, err := strconv.Atoi(scanner.Text())
				if err != nil {
					Error(err.Error())
					os.Exit(1)
				}
				// Delete the class
				l := len(schedule.Classes)
				copy(schedule.Classes[t:], schedule.Classes[t:])
				schedule.Classes[l-1] = Class{}
				schedule.Classes = schedule.Classes[:l-1]
			}

			// Delete file if no classes are present
			if len(schedule.Classes) <= 0 {
				os.Remove(Basepath + "/schedule.yml")
				break
			}

			// Rewrite file
			file, _ = yaml.Marshal(schedule)
			ioutil.WriteFile("schedule.yml", file, 0644)
			break
		default:
			StartProgram()
		}

	}
}

// ConfigQuestions prompts questions used to fill in the config
func ConfigQuestions() Config {
	questions := []string{
		Purple("The amount of people left before you leave"),
		Purple("The window of time to join a class after the inital time (minutes)"),
		Purple("The email used to login"),
		Purple("The password for that email"),
	}
	answers := make([]string, 0, 4)
	scanner := bufio.NewScanner(os.Stdin)

	for i, q := range questions {
		fmt.Println(Yellow("> ") + q)

		// Hide pasword input
		if i == len(questions)-1 {
			fmt.Print("\033[8m")   // Hide text
			fmt.Print("\033[?25l") // Hide cursor

			scanner.Scan()
			t := scanner.Text()
			answers = append(answers, t)

			fmt.Println("\033[28m") // Show text
			fmt.Print("\033[?25h")  // Show cursor

			break
		}

		scanner.Scan()
		t := scanner.Text()
		answers = append(answers, t)
	}
	l, _ := strconv.Atoi(answers[0])
	s, _ := strconv.Atoi(answers[1])

	if l <= 0 || s <= 0 {
		Error("Leave count or Skip count cannot be 0")
		os.Exit(1)
	}

	return Config{Leave: l, Skip: s, Email: answers[2], Password: answers[3]}
}

// ClassQuestions prompts questions used to fill in class details
func ClassQuestions() Class {
	questions := []string{
		Teal("A friendly name for the class"),
		Teal("The weekdays the class takes place (seperate by space)"),
		Teal("The time when the class starts (ex. 7:01am)"),
		Teal("The Google Meets URL needed to join"),
	}
	answers := make([]string, 0, 4)
	scanner := bufio.NewScanner(os.Stdin)

	for _, q := range questions {
		fmt.Println(Red("> ") + q)
		scanner.Scan()
		t := scanner.Text()
		answers = append(answers, t)
	}

	name := answers[0]
	if len(name) <= 0 {
		Error("Name of class cannot be empty")
		os.Exit(1)
	}

	// Weekday validation
	weekdaysStr := strings.Split(answers[1], " ")
	weekdays := make([]time.Weekday, 0, len(weekdaysStr))
	for _, w := range weekdaysStr {
		wd, err := ParseWeekday(w)
		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}

		weekdays = append(weekdays, wd)
	}
	// Time validation
	now := time.Now().Local()
	timeStr := strings.Split(answers[2], ":")
	t, err := time.Parse("01/02/2006 3:04pm", fmt.Sprintf("%d/%d/%d %s:%s", now.Month(), now.Day(), now.Year(), timeStr[0], timeStr[1]))

	if err != nil {
		Error("Could not parse time,", err.Error())
		os.Exit(1)
	}

	// Url Validation
	url, err := url.ParseRequestURI(answers[3])
	if err != nil {
		Error(err.Error())
		os.Exit(1)
	}

	return Class{Name: name, Weekdays: weekdays, JoinTime: t, URI: url}
}
