package collector

import "strings"

func (c Collector) ParseTags() string {
	var (
		result   string = "posts?tags="
		splitted []string
	)
	postTags := strings.ToLower(*c.PostTags)
	splitted = strings.Split(postTags, ",")
	for i, j := range splitted {
		j = strings.ReplaceAll(j, " ", "_")
		result += j
		if i != len(splitted)-1 {
			result += "+"
		}
	}

	return result
}
