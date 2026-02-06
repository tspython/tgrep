package domain

type SearchResult struct {
	File    string
	Line    int
	Column  int
	Content string
}
