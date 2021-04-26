package simple

import (
	"testing"
)

func TestCasualFormat(t *testing.T) {
	p := Person{
		Name: "Jeff",
	}
	want := "Sup, Jeff"

	if want != FormatCasual(p) {
		t.Errorf("Oops")
	}
}

func TestCasualFormatUsesNickname(t *testing.T) {
	p := Person{
		Name:     "Jeff",
		Nickname: "J-Dawg",
	}
	want := "Sup, J-Dawg"

	if want != FormatCasual(p) {
		t.Errorf("Oops")
	}
}

func TestFormatProfessional(t *testing.T) {
	p := Person{
		Name: "Jeff",
	}
	want := "Greetings, Jeff"

	if want != FormatProfessional(p) {
		t.Errorf("Oops")
	}
}

func TestFormatProfessionalUsesTitle(t *testing.T) {
	p := Person{
		Name:  "Jeff",
		Title: "Dr.",
	}
	want := "Greetings, Dr. Jeff"

	if want != FormatProfessional(p) {
		t.Errorf("Oops")
	}
}

func TestFormatProfessionalYellsAtAstronauts(t *testing.T) {
	p := Person{
		Name:  "Jeff",
		Title: "Astronaut",
	}
	want := "GREETINGS, ASTRONAUT JEFF"

	if want != FormatProfessional(p) {
		t.Errorf("Oops")
	}
}
