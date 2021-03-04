package main

import (
	"errors"
	"fmt"
	"github.com/AB0529/prompter"
	"github.com/AlecAivazis/survey/v2"
	"github.com/gookit/color"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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
	// Red color red
	Red = color.Red
	// Blue color blue
	Blue = color.Blue
	// Cyan color cyan
	Cyan = color.Cyan
	// Purple color purple
	Purple = color.Magenta
	// Green color green
	Green = color.LightGreen
	// Yellow color yellow
	Yellow = color.Yellow
	// White color white
	White = color.White

)

// Info prints an info prompt
func Info(i ...interface{}) {
	fmt.Printf("[%s] - %s\n", Yellow.Sprint("INFO"), i)
}

// Error prints an error prompt
func Error(msg ...interface{}) {
	fmt.Printf("[%s] - %s\n", Red.Sprint("ERROR"), msg)
}

// ------------------------------------------------------
// ------------------- Prompts ---------------------
// ------------------------------------------------------

// UpdateQuestion asks whether user wants to update or not
func UpdateQuestion() bool {
	var answer bool

	updateQuestion := &survey.Confirm{
		Message:  "Update found, would you like to update?",
		Default:  false,
		Help:     "",
	}

	survey.AskOne(updateQuestion, &answer)
	return answer
}

// ConfigQuestions the questions to be asked when creating a new config file
func ConfigQuestions() *Config {
	configQuestions := []*survey.Question{
		{
			Name: "leave",
			Prompt: &survey.Input{Message: "The amount of people left before you leave"},
			Validate: IsOverZero,
		},
		{
			Name: "skip",
			Prompt: &survey.Input{Message: "The window of time to join a class after the initial time (minutes)"},
		},
		{
			Name: "email",
			Prompt: &survey.Input{Message: "The email used to login"},
		},
		{
			Name: "password",
			Prompt: &survey.Password{Message: "The password for that email"},
		},
	}

	// Ask the question and create a config
	var answers *Config
	survey.Ask(configQuestions, &answers, survey.WithValidator(survey.Required))

	return answers
}

// ClassQuestions the questions to be asked when creating a new class
func ClassQuestions() *Class {
	classQuestions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{Message: "A friendly name for the class"},
		},
		{
			Name: "weekdays",
			Prompt: &survey.Input{Message: "The weekdays the class takes place (separated by space)"},
			Validate: IsWeekday,
		},
		{
			Name: "time",
			Prompt: &survey.Input{Message: "The time the class starts (ex. 7:01am)"},
			Validate: IsValidTime,
		},
		{
			Name: "url",
			Prompt: &survey.Input{Message: "The Google Meets URL needed to join"},
			Validate: IsURL,
		},
	}
	
	// Ask question and get the answer
	answers := struct {
		Name     string
		Weekdays string
		URL      string
		Time     string
	}{}
	survey.Ask(classQuestions, &answers, survey.WithValidator(survey.Required))
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
	parsedTime, _ := time.Parse("01/02/2006 3:04pm", fmt.Sprintf("%s %s:%s", now.Format("01/02/2006"), timeArr[0], timeArr[1]))

	class.JoinTime = &parsedTime
	parsedURL, _ := url.ParseRequestURI(answers.URL)
	class.MeetURL = parsedURL

	return class
}

// ScheduleQuestions the question to be asked when creating a new schedule file
func ScheduleQuestions() *Schedule {
	ClearScreen()
	scheduleQuestion := []*survey.Question{
		{
			Name: "classCount",
			Prompt: &survey.Input{Message: "How many total classes do you want to add?"},
			Validate: IsOverZero,
		},
	}

	var classCount int
	survey.Ask(scheduleQuestion, &classCount, survey.WithValidator(survey.Required))

	// Create the new classes
	schedule := Schedule{}
	for i := 0; i < classCount; i++ {
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

	selectSchedule := []*survey.Question{
		{
			Name: "schedule",
			Prompt:  &survey.Select{
				Message: "Which schedule do you want to use?",
				Options: scheduleFiles,
			},
			Validate: survey.Required,
		},
	}
	var schedule string
	err := survey.Ask(selectSchedule, &schedule)
	if err != nil {
		os.Exit(0)
	}

	return schedule
}

// InitalQuestions the question in which what happens with the program
func InitalQuestions() string {
	options := []string{
		prompter.Purple.Sprint("Start Program"),
		"Select or Create Schedule",
		prompter.Cyan.Sprint("Edit Class"),
		prompter.Green.Sprint("Add Class"),
		prompter.Red.Sprint("Delete Class"),
	}

	initalQuestions :=[]*survey.Question{
		{
			Name: "choice",
			Prompt: &survey.Select{
				Message: "What do you want to do?",
				Options: options,
			},
		},
	}

	var choice string
	survey.Ask(initalQuestions, &choice, survey.WithValidator(survey.Required))

	return choice
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

// IsNumeric makes sure the value is numeric
func IsNumeric(val interface{}) error {
	_, err := strconv.Atoi(val.(string))

	// Handle non numeric
	if err != nil {
		return errors.New("Value is not numeric")
	}

	return nil
}

// IsURL makes sure a value is a valid URL
func IsURL(val interface{}) error {
	_, err := url.ParseRequestURI(val.(string))

	if err != nil {
		return errors.New("Value is invalid URL")
	}

	return nil
}
