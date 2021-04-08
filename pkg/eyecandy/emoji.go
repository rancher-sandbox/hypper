/*
Copyright SUSE LLC.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*Package eyecandy provides common methods to print messages with emojis
 */
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
