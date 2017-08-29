package utils

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"regexp"

	log "github.com/Sirupsen/logrus"
)

var (
	// Matches literal string: ${Local::IncludeEnv <key>}
	// example:   ${Local::IncludeEnv HOME}
	literalIncludeEnvRe = regexp.MustCompile(`\${[ ]*Local::IncludeEnv[ ]+([[:ascii:]]+)}`)
	// Matches tag  !Local::IncludeEnv <key>
	// for example:   !Local::IncludeEnv HOME
	valueIncludeEnvTagRe = regexp.MustCompile(`!Local::IncludeEnv[ ]+([[:ascii:]]+)`)
)

func ApplyIncludeEnvDirective(reader io.Reader) []byte {
	output := bytes.NewBuffer([]byte{})
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Bytes()

		loc := literalIncludeEnvRe.FindIndex(line)
		tagLoc := valueIncludeEnvTagRe.FindIndex(line)

		if len(loc) == 2 {
			envName := string(literalIncludeEnvRe.FindSubmatch(line)[1])
			log.Debugf("loading include env: %s", envName)

			output.Write(line[0:loc[0]])
			envValue := os.Getenv(envName)
			output.WriteString(envValue)
			output.Write(line[loc[1]:])

		} else if len(tagLoc) == 2 {
			envName := string(valueIncludeEnvTagRe.FindSubmatch(line)[1])
			log.Debugf("loading include env: %s", envName)

			output.Write(line[0:tagLoc[0]])
			envValue := os.Getenv(envName)
			output.WriteString(envValue)
			output.Write(line[tagLoc[1]:])

		} else {
			output.Write(line)
		}

		output.WriteByte('\n')
	}
	return output.Bytes()
}
