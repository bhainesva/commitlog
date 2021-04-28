package simple

import "strings"

type Person struct {
	Name     string // This is their normal name
	Nickname string // You can call them this
	Title    string // It's their title
	Unused   string
}

// FormatCasual formats a person's name like they're your friend
func FormatCasual(p Person) string {
	nameToUse := p.Name

	if p.Nickname != "" {
		nameToUse = p.Nickname
	}

	return "Sup, " + nameToUse
}

// FormatProfessional formats a person's name very officially
func FormatProfessional(p Person) string {
	suffix := p.Name
	if p.Title != "" {
		suffix = p.Title + " " + suffix
	}

	greeting := "Greetings, " + suffix

	// Need to yell, they're far away
	if p.Title == "Astronaut" {
		greeting = strings.ToUpper(greeting)
	}

	return greeting
}
