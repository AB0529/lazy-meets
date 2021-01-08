package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AB0529/prompter"
)

var (
	// Weekdays a map of string to time.Weekday
	Weekdays = map[string]time.Weekday{
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}
)

// Info prints an info prompt
func Info(i ...interface{}) {
	fmt.Printf("[%s] - %s\n", prompter.Yellow.Sprint("INFO"), i)
}

// Error prints an error prompt
func Error(msg ...interface{}) {
	fmt.Printf("[%s] - %s\n", prompter.Red.Sprint("ERROR"), msg)
}

// ------------------------------------------------------
// ------------------- Prompts ---------------------
// ------------------------------------------------------

// ConfigQuestions the questions to be asked when creating a new config file
func ConfigQuestions() *Config {
	configQuestions := []interface{}{
		&prompter.Input{
			Message:    "The amount of people left before you leave",
			Name:       "leave",
			Validators: []prompter.Validator{prompter.Required, prompter.IsNumeric, IsOverZero},
		},
		&prompter.Input{
			Message:    "The window of time to join a class after the inital time (minutes)",
			Name:       "skip",
			Validators: []prompter.Validator{prompter.Required, prompter.IsNumeric, IsOverZero},
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

	// Ask the question and create a config
	answers := struct {
		Leave    int
		Skip     int
		Email    string
		Password string
	}{}
	prompter.Ask(&prompter.Prompt{Types: configQuestions}, &answers)

	return &Config{Leave: answers.Leave, Skip: answers.Skip, Email: answers.Email, Password: answers.Password}
}

// ClassQuestions the questions to be asked when creating a new class
func ClassQuestions() *Class {
	classQuestions := []interface{}{
		&prompter.Input{
			Message:    "A friendly name for the class",
			Name:       "name",
			Validators: []prompter.Validator{prompter.Required},
		},
		&prompter.Input{
			Message:    "The weekdays the class takes place (seperate by space)",
			Name:       "weekdays",
			Validators: []prompter.Validator{prompter.Required, IsWeekday},
		},
		&prompter.Input{
			Message:    "The time when the class starts (ex. 7:01am)",
			Name:       "time",
			Validators: []prompter.Validator{prompter.Required, IsValidTime},
		},
		&prompter.Input{
			Message:    "The Google Meets URL needed to join",
			Name:       "url",
			Validators: []prompter.Validator{prompter.Required, prompter.IsURL},
		},
	}
	// Ask question and get the answer
	answers := struct {
		Name     string
		Weekdays string
		URL      string
		Time     string
	}{}
	prompter.Ask(&prompter.Prompt{Types: classQuestions}, &answers)
	class := &Class{
		Name: answers.Name,
	}

	// Parse each answer to correct form for schedule
	weekdayStr := strings.Split(answers.Weekdays, " ")
	for _, w := range weekdayStr {
		day, _ := Weekdays[strings.ToLower(w[:3])]
		class.Weekdays = append(class.Weekdays, &day)
	}

	timeArr := strings.Split(answers.Time, ":")
	now := time.Now().Local()
	time, _ := time.Parse("01/02/2006 3:04pm", fmt.Sprintf("%s %s:%s", now.Format("01/02/2006"), timeArr[0], timeArr[1]))

	class.JoinTime = &time
	url, _ := url.ParseRequestURI(answers.URL)
	class.MeetURL = url

	return class
}

// ScheduleQuestions the question to be asked when creating a new schedule file
func ScheduleQuestions() *Schedule {
	ClearScreen()
	scheduleQuestions := []interface{}{
		&prompter.Input{
			Message:    "How many total classes do you want to add?",
			Name:       "classCount",
			Validators: []prompter.Validator{prompter.Required, prompter.IsNumeric, IsOverZero},
		},
	}
	answers := struct{ ClassCount int }{}
	prompter.Ask(&prompter.Prompt{Types: scheduleQuestions}, &answers)
	// Create the new classes
	schedule := Schedule{}
	for i := 0; i < answers.ClassCount; i++ {
		Info("Enter details for class " + prompter.Green.Sprint(i+1))
		class := ClassQuestions()
		schedule = append(schedule, class)
	}

	return &schedule
}

// SelectSchedule prompts for which schedule to use
// TODO: add delete schedule option
func SelectSchedule(overrwrite bool) string {
	files, _ := ioutil.ReadDir("./Schedules")

	if len(files) < 2 && !overrwrite {
		return "schedule_1.yml"
	}

	scheduleFiles := []string{"Create New Schedule"}
	for _, f := range files {
		scheduleFiles = append([]string{f.Name()}, scheduleFiles...)
	}

	selectSchedule := []interface{}{
		&prompter.Multiselect{
			Name:    "schedule",
			Message: "Which schedule do you want to use?",
			Options: scheduleFiles,
		},
	}
	answer := map[string]interface{}{}
	prompter.Ask(&prompter.Prompt{Types: selectSchedule}, &answer)

	return answer["schedule"].(string)
}

// InitalQuestions the question in which what happens with the program
func InitalQuestions() string {
	options := []string{
		prompter.Red.Sprint("Delete Class"),
		prompter.Green.Sprint("Add Class"),
		prompter.Cyan.Sprint("Edit Class"),
		"Select or Create Schedule",
		prompter.Purple.Sprint("Start Program"),
	}
	initalQuestions := []interface{}{
		&prompter.Multiselect{
			Message: "What do you want to do?",
			Name:    "Choice",
			Options: options,
		},
	}
	ans := struct{ Choice string }{}
	prompter.Ask(&prompter.Prompt{Types: initalQuestions}, &ans)

	return ans.Choice
}

// ------------------------------------------------------
// ------------------- Validators -------------------
// ------------------------------------------------------

// IsOverZero validator in which input must be over 0
func IsOverZero(inp interface{}) error {
	// Convert to int
	n, _ := strconv.Atoi(inp.(string))
	if n <= 0 {
		return errors.New("value must be over 0")
	}

	return nil
}

// IsWeekday validator in which input must be a correct time.Weekday type
func IsWeekday(inp interface{}) error {
	// Split the srting by spaces
	weekdayStr := strings.Split(inp.(string), " ")
	// Loop through each and determin the validity
	for _, w := range weekdayStr {
		_, ok := Weekdays[strings.ToLower(w[:3])]

		if !ok {
			return fmt.Errorf("invalid weekday %s", w)
		}
	}

	return nil
}

// IsValidTime validator in which the input must parse to a correct time format
func IsValidTime(inp interface{}) error {
	// Split input by ":"
	timeArr := strings.Split(inp.(string), ":")
	if len(timeArr) < 2 {
		return fmt.Errorf("%s is not a valid time", inp.(string))
	}
	// Attempt to parse the time
	now := time.Now().Local()
	_, err := time.Parse("01/02/2006 3:04pm", fmt.Sprintf("%s %s:%s", now.Format("01/02/2006"), timeArr[0], timeArr[1]))
	if err != nil {
		return fmt.Errorf("%s is not a valid time", inp.(string))
	}

	return nil
}
