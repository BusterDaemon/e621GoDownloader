package parsers

type Post struct {
	Width   int
	Height  int
	Ext     string
	Hash    string
	Url     string
	Score   int
	Rating  string
	Tags    string
	Sources string
}

func convertArray(arr []string) string {
	var total string
	for i, j := range arr {
		total += j
		if i < len(arr)-1 {
			total += ","
		}
	}
	return total
}
