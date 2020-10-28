package utility

import (
	"config"
	"strings"
)

func HideName(name string) string {
	if *config.EuhtHideNameStategy == false {
		return name
	}
	runeSlice := []rune(name)
	length := len(runeSlice)
	if length > 0 {
		return string(runeSlice[0]) + strings.Repeat("*", length-1)
	} else {
		return name
	}

}

func HidePhone(phone string) string {
	if *config.EuhtHidePhoneStategy == false {
		return phone
	}
	if len(phone) == 11 {
		return phone[:3] + "****" + phone[7:]
	} else {
		return phone
	}
}
