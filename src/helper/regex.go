// Helper functions will go here.
package helper

import "regexp"

type Regex struct{}

func (regex Regex) ParseTopic(uri string, serviceName string) string {

	compiledRegexp, _ := regexp.Compile("^/" + serviceName + "/([a-z]+)$")
	matches := compiledRegexp.FindStringSubmatch(uri)

	if len(matches) > 0 {
		return matches[1]
	}

	return ""
}
