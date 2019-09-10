package config

import (
	"fmt"
	"strings"
)

type FrontendExplorerType uint8

const (
	FrontendExplorerTypeVueTypescript FrontendExplorerType = iota
	FrontendExplorerTypePlainJavascript
	FrontendExplorerTypeNone
)

func (et FrontendExplorerType) String() string {
	switch et {
	case FrontendExplorerTypePlainJavascript:
		return "plainjs"
	case FrontendExplorerTypeNone:
		return ""
	default:
		return "vuets"
	}
}

func (et *FrontendExplorerType) FromString(str string) error {
	str = strings.ToLower(str)
	switch str {
	case "vuets", "vue":
		*et = FrontendExplorerTypeVueTypescript
		return nil
	case "plainjs", "plain":
		*et = FrontendExplorerTypePlainJavascript
		return nil
	case "", "no", "none", "nil":
		*et = FrontendExplorerTypeNone
		return nil
	default:
		return fmt.Errorf("%s is an invalid FrontendExplorerType in string format", str)
	}
}
