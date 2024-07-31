package parsers

import "fmt"

type PostTable struct {
	postData map[int]Post
}

func NewPostTable() *PostTable {
	return &PostTable{
		postData: make(map[int]Post),
	}
}

func (pt *PostTable) AddPostTable(index int, data Post) {
	pt.postData[index] = data
}

func (pt *PostTable) GetPostTable(index int) Post {
	return pt.postData[index]
}

func (pt *PostTable) RemovePostTable(index int) {
	delete(pt.postData, index)
}

func (pt *PostTable) GetLengthTable() int {
	return len(pt.postData)
}

func (pt *PostTable) PrintAll() {
	for i := 0; i < len(pt.postData); i++ {
		fmt.Printf("%#v\n", pt.postData[i])
	}
}
