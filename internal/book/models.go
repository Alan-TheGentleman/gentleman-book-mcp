package book

// Chapter representa un capítulo del libro
type Chapter struct {
	ID        string    `json:"id"`
	Order     int       `json:"order"`
	Name      string    `json:"name"`
	Locale    string    `json:"locale"`
	TitleList []Section `json:"titleList"`
	Content   string    `json:"content"`
	FilePath  string    `json:"filePath"`
}

// Section representa una sección dentro de un capítulo
type Section struct {
	Name  string `json:"name"`
	TagID string `json:"tagId"`
}

// SearchResult representa un resultado de búsqueda
type SearchResult struct {
	ChapterID   string  `json:"chapterId"`
	ChapterName string  `json:"chapterName"`
	Section     string  `json:"section"`
	Snippet     string  `json:"snippet"`
	LineNumber  int     `json:"lineNumber"`
	Relevance   float64 `json:"relevance"`
	Locale      string  `json:"locale"`
}

// BookIndex representa el índice completo del libro
type BookIndex struct {
	Locale        string    `json:"locale"`
	TotalChapters int       `json:"totalChapters"`
	Chapters      []Chapter `json:"chapters"`
}
