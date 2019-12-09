package config

import (
	"fmt"
	"strings"
)

//FrontendExplorerType defines the type of explorer to be generated
type FrontendExplorerType uint8

const (
	frontendExplorerTypeVueTypescript FrontendExplorerType = iota
	frontendExplorerTypePlainJavascript
	frontendExplorerTypeNone
)

func (et FrontendExplorerType) String() string {
	switch et {
	case frontendExplorerTypePlainJavascript:
		return "plainjs"
	case frontendExplorerTypeNone:
		return ""
	default:
		return "vuets"
	}
}

//FromString tries to parse the passed string in to a  valid FrontendExplorerType and returns an error if it is not recognized
func (et *FrontendExplorerType) FromString(str string) error {
	str = strings.ToLower(str)
	switch str {
	case "vuets", "vue":
		*et = frontendExplorerTypeVueTypescript
		return nil
	case "plainjs", "plain":
		*et = frontendExplorerTypePlainJavascript
		return nil
	case "", "no", "none", "nil":
		*et = frontendExplorerTypeNone
		return nil
	default:
		return fmt.Errorf("%s is an invalid FrontendExplorerType in string format", str)
	}
}
