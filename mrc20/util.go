package mrc20

import (
	"regexp"
)

func PathParse(pathStr string) (path, query, key, operator, value string) {
	regex := regexp.MustCompile(`^([^[]+)(\['.*'\])$`)
	matches := regex.FindStringSubmatch(pathStr)
	if len(matches) != 3 {
		return
	}
	path = matches[1]
	query = matches[2]
	queryRegex := regexp.MustCompile(`\['([^']+)'(?:\s*(=|#=)\s*)'([^']+)'\]`)
	queryMatches := queryRegex.FindStringSubmatch(query)
	if len(queryMatches) != 4 {
		return
	}
	key = queryMatches[1]
	operator = queryMatches[2]
	value = queryMatches[3]
	return
}
