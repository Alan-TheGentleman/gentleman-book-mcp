package book

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Parser handles parsing of MDX book files
type Parser struct {
	bookPath string
}

// NewParser creates a new parser with the book path
func NewParser(bookPath string) *Parser {
	return &Parser{bookPath: bookPath}
}

// frontmatter represents the YAML frontmatter from MDX
type frontmatter struct {
	ID        string    `json:"id"`
	Order     int       `json:"order"`
	Name      string    `json:"name"`
	TitleList []Section `json:"titleList"`
}

// ParseChapter parses an MDX file and returns a Chapter
func (p *Parser) ParseChapter(filePath string, locale string) (*Chapter, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	contentStr := string(content)

	// Separate frontmatter from content
	fm, body, err := p.parseFrontmatter(contentStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing frontmatter in %s: %w", filePath, err)
	}

	return &Chapter{
		ID:        fm.ID,
		Order:     fm.Order,
		Name:      fm.Name,
		Locale:    locale,
		TitleList: fm.TitleList,
		Content:   body,
		FilePath:  filePath,
	}, nil
}

// parseFrontmatter extracts the YAML frontmatter from MDX content
func (p *Parser) parseFrontmatter(content string) (*frontmatter, string, error) {
	// Frontmatter is between --- and ---
	if !strings.HasPrefix(content, "---") {
		return nil, content, fmt.Errorf("no frontmatter found")
	}

	// Find the second ---
	endIndex := strings.Index(content[3:], "---")
	if endIndex == -1 {
		return nil, content, fmt.Errorf("frontmatter not closed")
	}

	fmContent := content[3 : endIndex+3]
	body := strings.TrimSpace(content[endIndex+6:])

	// Parse frontmatter manually (it's YAML-like but with JSON arrays)
	fm := &frontmatter{}

	// Extract id (supports both quoted and unquoted values)
	idRegex := regexp.MustCompile(`id:\s*(?:['"]([^'"]+)['"]|([^\s'"]+))`)
	idMatch := idRegex.FindStringSubmatch(fmContent)
	if len(idMatch) > 1 {
		if idMatch[1] != "" {
			fm.ID = idMatch[1] // quoted value
		} else if len(idMatch) > 2 {
			fm.ID = idMatch[2] // unquoted value
		}
	}

	// Extract order
	orderMatch := regexp.MustCompile(`order:\s*(\d+)`).FindStringSubmatch(fmContent)
	if len(orderMatch) > 1 {
		fm.Order, _ = strconv.Atoi(orderMatch[1])
	}

	// Extract name (supports both quoted and unquoted values)
	// For unquoted, capture until end of line
	nameRegex := regexp.MustCompile(`(?m)^name:\s*(?:['"]([^'"]+)['"]|([^\n]+))`)
	nameMatch := nameRegex.FindStringSubmatch(fmContent)
	if len(nameMatch) > 1 {
		if nameMatch[1] != "" {
			fm.Name = nameMatch[1] // quoted value
		} else if len(nameMatch) > 2 {
			fm.Name = strings.TrimSpace(nameMatch[2]) // unquoted value
		}
	}

	// Extract titleList (it's a JSON-like array)
	titleListStart := strings.Index(fmContent, "titleList:")
	if titleListStart != -1 {
		// Find the complete array
		arrayStart := strings.Index(fmContent[titleListStart:], "[")
		if arrayStart != -1 {
			bracketCount := 0
			arrayEnd := -1
			startPos := titleListStart + arrayStart

			for i := startPos; i < len(fmContent); i++ {
				if fmContent[i] == '[' {
					bracketCount++
				} else if fmContent[i] == ']' {
					bracketCount--
					if bracketCount == 0 {
						arrayEnd = i + 1
						break
					}
				}
			}

			if arrayEnd != -1 {
				arrayContent := fmContent[startPos:arrayEnd]
				// Clean content to make it valid JSON
				arrayContent = p.cleanArrayToJSON(arrayContent)

				var sections []Section
				if err := json.Unmarshal([]byte(arrayContent), &sections); err == nil {
					fm.TitleList = sections
				}
			}
		}
	}

	return fm, body, nil
}

// cleanArrayToJSON cleans YAML-like array to valid JSON
func (p *Parser) cleanArrayToJSON(content string) string {
	// Replace single quotes with double quotes
	content = strings.ReplaceAll(content, "'", "\"")

	// Ensure keys are quoted
	content = regexp.MustCompile(`(\s)name:`).ReplaceAllString(content, `$1"name":`)
	content = regexp.MustCompile(`(\s)tagId:`).ReplaceAllString(content, `$1"tagId":`)
	content = regexp.MustCompile(`{\s*name:`).ReplaceAllString(content, `{"name":`)
	content = regexp.MustCompile(`{\s*tagId:`).ReplaceAllString(content, `{"tagId":`)

	// Clean extra spaces and newlines
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	return content
}

// ListChapters lists all chapters for a locale
func (p *Parser) ListChapters(locale string) ([]Chapter, error) {
	localePath := filepath.Join(p.bookPath, locale)

	entries, err := os.ReadDir(localePath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", localePath, err)
	}

	var chapters []Chapter
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mdx") {
			continue
		}

		filePath := filepath.Join(localePath, entry.Name())
		chapter, err := p.ParseChapter(filePath, locale)
		if err != nil {
			// Log error but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: could not parse %s: %v\n", filePath, err)
			continue
		}
		chapters = append(chapters, *chapter)
	}

	// Sort by order
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Order < chapters[j].Order
	})

	return chapters, nil
}

// GetChapter gets a specific chapter by ID
func (p *Parser) GetChapter(chapterID string, locale string) (*Chapter, error) {
	chapters, err := p.ListChapters(locale)
	if err != nil {
		return nil, err
	}

	for _, ch := range chapters {
		if ch.ID == chapterID {
			return &ch, nil
		}
	}

	return nil, fmt.Errorf("chapter not found: %s", chapterID)
}

// GetSection gets a specific section from a chapter
func (p *Parser) GetSection(chapterID string, sectionTagID string, locale string) (string, error) {
	chapter, err := p.GetChapter(chapterID, locale)
	if err != nil {
		return "", err
	}

	// Search for the section in content
	lines := strings.Split(chapter.Content, "\n")

	// Find the header that matches the tagId
	inSection := false
	var sectionContent strings.Builder
	headerPattern := regexp.MustCompile(`^#{1,6}\s+(.+)$`)

	for _, line := range lines {
		if matches := headerPattern.FindStringSubmatch(line); len(matches) > 1 {
			headerText := matches[1]
			currentTagID := p.generateTagID(headerText)

			if currentTagID == sectionTagID {
				inSection = true
				sectionContent.WriteString(line)
				sectionContent.WriteString("\n")
				continue
			} else if inSection {
				// Reached another section, stop
				break
			}
		}

		if inSection {
			sectionContent.WriteString(line)
			sectionContent.WriteString("\n")
		}
	}

	if sectionContent.Len() == 0 {
		return "", fmt.Errorf("section not found: %s", sectionTagID)
	}

	return strings.TrimSpace(sectionContent.String()), nil
}

// generateTagID generates a tagId from a title
func (p *Parser) generateTagID(title string) string {
	// Convert to lowercase
	tagID := strings.ToLower(title)

	// Replace spaces with hyphens
	tagID = strings.ReplaceAll(tagID, " ", "-")

	// Remove special characters except hyphens and accented letters
	tagID = regexp.MustCompile(`[^\p{L}\p{N}-]`).ReplaceAllString(tagID, "")

	// Remove multiple hyphens
	tagID = regexp.MustCompile(`-+`).ReplaceAllString(tagID, "-")

	// Remove leading and trailing hyphens
	tagID = strings.Trim(tagID, "-")

	return tagID
}

// Search searches content in the book
func (p *Parser) Search(query string, locale string) ([]SearchResult, error) {
	chapters, err := p.ListChapters(locale)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	for _, chapter := range chapters {
		scanner := bufio.NewScanner(strings.NewReader(chapter.Content))
		lineNum := 0
		currentSection := ""
		headerPattern := regexp.MustCompile(`^#{1,6}\s+(.+)$`)

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			lineLower := strings.ToLower(line)

			// Update current section
			if matches := headerPattern.FindStringSubmatch(line); len(matches) > 1 {
				currentSection = matches[1]
			}

			// Search for matches
			matchCount := 0
			for _, word := range queryWords {
				if strings.Contains(lineLower, word) {
					matchCount++
				}
			}

			if matchCount > 0 {
				relevance := float64(matchCount) / float64(len(queryWords))

				// Create snippet with context
				snippet := line
				if len(snippet) > 200 {
					snippet = snippet[:200] + "..."
				}

				results = append(results, SearchResult{
					ChapterID:   chapter.ID,
					ChapterName: chapter.Name,
					Section:     currentSection,
					Snippet:     snippet,
					LineNumber:  lineNum,
					Relevance:   relevance,
					Locale:      locale,
				})
			}
		}
	}

	// Sort by relevance
	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance > results[j].Relevance
	})

	// Limit results
	if len(results) > 20 {
		results = results[:20]
	}

	return results, nil
}

// GetBookIndex gets the complete book index
func (p *Parser) GetBookIndex(locale string) (*BookIndex, error) {
	chapters, err := p.ListChapters(locale)
	if err != nil {
		return nil, err
	}

	// Clear content for index (metadata only)
	for i := range chapters {
		chapters[i].Content = "" // Don't include full content in index
	}

	return &BookIndex{
		Locale:        locale,
		TotalChapters: len(chapters),
		Chapters:      chapters,
	}, nil
}

// GetAvailableLocales returns available locales
func (p *Parser) GetAvailableLocales() ([]string, error) {
	entries, err := os.ReadDir(p.bookPath)
	if err != nil {
		return nil, fmt.Errorf("error reading book path: %w", err)
	}

	var locales []string
	for _, entry := range entries {
		if entry.IsDir() && (entry.Name() == "en" || entry.Name() == "es") {
			locales = append(locales, entry.Name())
		}
	}

	return locales, nil
}
