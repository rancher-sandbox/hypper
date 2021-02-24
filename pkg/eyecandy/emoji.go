package eyecandy

import (
	"fmt"
	"regexp"

	"github.com/kyokomi/emoji/v2"
)

func ESPrintf(emojisDisabled bool, format string, v ...interface{}) string {
	if emojisDisabled {
		return fmt.Sprintf(removeEmojiFromString(format), v...)
	}
	return emoji.Sprintf(format, v...)
}

func ESPrint(emojisDisabled bool, s string) string {
	if emojisDisabled {
		return fmt.Sprint(removeEmojiFromString(s))
	}
	return emoji.Sprint(s)
}

func removeEmojiFromString(s string) string {
	re := regexp.MustCompile(`:[a-zA-Z0-9-_+]+?:`)
	return re.ReplaceAllString(s, "")
}
