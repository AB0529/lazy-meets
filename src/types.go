package main

import (
	"net/url"
	"time"
)

// Config the repersentation of a config file
type Config struct {
	// Leave the number of people left which determins when to leave
	Leave int
	// Skip the window of time between the start of the class, and current time to join
	Skip int
	// Email Gmail login Email
	Email string
	// Password Gmail login password
	Password string
}

// Class the repersentation of a class
type Class struct {
	// Name a friendly name for the class
	Name string
	// Weekdays the weekdays the class takes place
	Weekdays []*time.Weekday
	// MeetURL the URL to join
	MeetURL *url.URL
	// JoinTime the time in which  to join the class
	JoinTime *time.Time
}

// Schedule the repersentation of a created schedule
type Schedule []*Class
