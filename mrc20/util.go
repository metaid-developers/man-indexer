package mrc20

import (
	"fmt"
	"regexp"
)

func PathParse(pathStr string) (path, query, key, operator, value string) {
	regex := regexp.MustCompile(`^([^[]+)(\['.*'\])$`)
	matches := regex.FindStringSubmatch(path)

	if len(matches) != 3 {
		return
	}
	path = matches[1]
	query = matches[2]
	queryRegex := regexp.MustCompile(`\['([^']+)'(?:\s*(=|#=)\s*)'([^']+)'\]`)
	queryMatches := queryRegex.FindStringSubmatch(query)
	fmt.Println(queryMatches)
	if len(queryMatches) != 4 {
		return
	}
	key = queryMatches[1]
	operator = queryMatches[2]
	value = queryMatches[3]
	return
}
