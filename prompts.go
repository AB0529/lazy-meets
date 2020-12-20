package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/AB0529/prompter"

	"gopkg.in/yaml.v2"
)

// Config the structure for the config
type Config struct {
	Leave    int
	Skip     int
	Email    string
	Password string
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

func isOverZero(val interface{}) error {
	n, err := strconv.Atoi(val.(string))
	if err != nil {
		return errors.New("Value has to be numeric")
	}
	if n <= 0 {
		return errors.New("Value has to be above 0")
	}

	return nil
}

// Weekday validation
func isWeekday(val interface{}) error {
	weekdaysStr := strings.Split(val.(string), " ")
	for _, w := range weekdaysStr {
		_, err := ParseWeekday(w)
		if err != nil {
			return errors.New(w + " is not a valid day")
		}
	}

	return nil
}

// Time validation
func isCorrectTime(val interface{}) error {
	// Make sure input has ':'
	if !strings.Contains(val.(string), ":") {
		return errors.New(val.(string) + " is not a valid time")
	}

	timeStr := strings.Split(val.(string), ":")
	now := time.Now().Local()
	_, err := time.Parse("01/02/2006 3:04pm", fmt.Sprintf("%d/%d/%d %s:%s", now.Month(), now.Day(), now.Year(), timeStr[0], timeStr[1]))

	if err != nil {
		return errors.New(val.(string) + " is not a valid time")
	}

	return nil
}

// IniitalQuesitons prompts questions used for inital prompt
func IniitalQuesitons() {
	qs := []interface{}{
		&prompter.Multiselect{
			Message:    "What do you want to do?",
			Name:       "initialPrompt",
			Options:    []string{Red("Delete Class"), Yellow("Add Class"), Teal("Edit Class"), Purple("Start Program")},
			Validators: []prompter.Validator{prompter.Required},
		},
	}

	var schedule Schedule
	ans := map[string]interface{}{}

	prompter.Ask(&prompter.Prompt{Types: qs}, &ans)

	// Get the schedule
	file, _ := ioutil.ReadFile("./schedule.yml")
	yaml.Unmarshal(file, &schedule)
	// Find all classes
	class := []string{}

	for _, classes := range schedule.Classes {
		class = append(class, classes.Name)
	}
	// Form class edit questions
	editQS := []interface{}{
		&prompter.Multiselect{
			Message: "Select a class",
			Name:    "class",
			Options: class,
		},
	}

	switch ans["initialPrompt"] {
	case Purple("Start Program"):
		ClearScreen()
		StartProgram()
		break
	// Edit
	case Teal("Edit Class"):
		ClearScreen()
		// The class to edit
		prompter.Ask(&prompter.Prompt{Types: editQS}, &ans)
		// Update the class
		for i, sc := range schedule.Classes {
			if sc.Name == ans["class"] {
				schedule.Classes[i] = ClassQuestions()
			}
		}
		// Rewrite file
		file, _ = yaml.Marshal(schedule)
		ioutil.WriteFile("schedule.yml", file, 0644)
		break
	// Add
	case Yellow("Add Class"):
		ClearScreen()
		newClass := ClassQuestions()
		schedule.Classes = append(schedule.Classes, newClass)
		// Rewrite file
		file, _ = yaml.Marshal(schedule)
		ioutil.WriteFile("schedule.yml", file, 0644)
		break
	// Delete
	case Red("Delete Class"):
		ClearScreen()
		// The class to delete
		prompter.Ask(&prompter.Prompt{Types: editQS}, &ans)

		// Delete the class
		l := len(schedule.Classes)
		for i, sc := range schedule.Classes {
			if sc.Name == ans["class"] {
				copy(schedule.Classes[i:], schedule.Classes[i:])
				schedule.Classes[l-1] = Class{}
				schedule.Classes = schedule.Classes[:l-1]
			}
		}

		// Delete file if no classes are present
		if len(schedule.Classes) <= 0 {
			os.Remove("./schedule.yml")
			break
		}

		// Rewrite file
		file, _ = yaml.Marshal(schedule)
		ioutil.WriteFile("schedule.yml", file, 0644)
		break
	default:
		StartProgram()
	}

	// Clear screen and init
	ClearScreen()
	Init()
}

// ConfigQuestions prompts questions used to fill in the config
func ConfigQuestions() Config {
	qs := []interface{}{
		&prompter.Input{
			Message:    "The amount of people left before you leave",
			Name:       "leave",
			Validators: []prompter.Validator{prompter.Required, isOverZero},
		},
		&prompter.Input{
			Message:    "The window of time to join a class after the inital time (minutes)",
			Name:       "skip",
			Validators: []prompter.Validator{prompter.Required, isOverZero},
		},
		&prompter.Input{
			Message:    "The email used to login",
			Name:       "email",
			Validators: []prompter.Validator{prompter.Required},
		},
		&prompter.Password{
			Message:    "The password used to login",
			Name:       "password",
			Validators: []prompter.Validator{prompter.Required},
		},
	}
	answers := struct {
		Leave    int
		Skip     int
		Email    string
		Password string
	}{}
	prompter.Ask(&prompter.Prompt{Types: qs}, &answers)

	if answers.Leave <= 0 || answers.Skip <= 0 {
		Error("Leave count or Skip count cannot be 0")
		os.Exit(1)
	}

	return Config{Leave: answers.Leave, Skip: answers.Skip, Email: answers.Email, Password: answers.Password}
}

// ClassQuestions prompts questions used to fill in class details
func ClassQuestions() Class {
	qs := []interface{}{
		&prompter.Input{
			Message:    "A friendly name for the class",
			Name:       "name",
			Validators: []prompter.Validator{prompter.Required},
		},
		&prompter.Input{
			Message:    "The weekdays the class takes place (seperate by space)",
			Name:       "weekdays",
			Validators: []prompter.Validator{prompter.Required, isWeekday},
		},
		&prompter.Input{
			Message:    "The time when the class starts (ex. 7:01am)",
			Name:       "time",
			Validators: []prompter.Validator{prompter.Required, isCorrectTime},
		},
		&prompter.Input{
			Message:    "The Google Meets URL needed to join",
			Name:       "url",
			Validators: []prompter.Validator{prompter.Required, prompter.IsURL},
		},
	}
	answers := struct {
		Name     string
		Weekdays string
		Time     string
		URL      string
	}{}
	// Ask the questions
	prompter.Ask(&prompter.Prompt{Types: qs}, &answers)

	// Weekday validation
	weekdayStr := strings.Split(answers.Weekdays, " ")
	weekdays := []time.Weekday{}
	for _, wd := range weekdayStr {
		w, _ := ParseWeekday(wd)
		weekdays = append(weekdays, w)
	}

	// Time validation
	now := time.Now().Local()
	timeStr := strings.Split(answers.Time, ":")
	t, _ := time.Parse("01/02/2006 3:04pm", fmt.Sprintf("%d/%d/%d %s:%s", now.Month(), now.Day(), now.Year(), timeStr[0], timeStr[1]))
	// Url Validation
	url, _ := url.ParseRequestURI(answers.URL)

	return Class{Name: answers.Name, Weekdays: weekdays, JoinTime: t, URI: url}
}
