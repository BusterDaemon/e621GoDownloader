package parsers

type Post struct {
	Width   int
	Height  int
	FileExt string
	Hash    string
	FileUrl string
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
