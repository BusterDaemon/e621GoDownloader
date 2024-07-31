package parsers

type Post struct {
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	FileExt string `json:"ext"`
	Hash    string `json:"md5"`
	FileUrl string `json:"file_url"`
	Score   int    `json:"score"`
	Rating  string `json:"rating"`
	Tags    string `json:"tags"`
	Sources string `json:"sources"`
}

type Parserer interface {
	Scrap() *PostTable
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
