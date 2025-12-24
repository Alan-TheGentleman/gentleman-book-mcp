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

// Parser maneja el parsing de archivos MDX del libro
type Parser struct {
	bookPath string
}

// NewParser crea un nuevo parser con la ruta al libro
func NewParser(bookPath string) *Parser {
	return &Parser{bookPath: bookPath}
}

// frontmatter representa el YAML frontmatter del MDX
type frontmatter struct {
	ID        string    `json:"id"`
	Order     int       `json:"order"`
	Name      string    `json:"name"`
	TitleList []Section `json:"titleList"`
}

// ParseChapter parsea un archivo MDX y retorna un Chapter
func (p *Parser) ParseChapter(filePath string, locale string) (*Chapter, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	contentStr := string(content)

	// Separar frontmatter del contenido
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

// parseFrontmatter extrae el frontmatter YAML del contenido MDX
func (p *Parser) parseFrontmatter(content string) (*frontmatter, string, error) {
	// El frontmatter está entre --- y ---
	if !strings.HasPrefix(content, "---") {
		return nil, content, fmt.Errorf("no frontmatter found")
	}

	// Encontrar el segundo ---
	endIndex := strings.Index(content[3:], "---")
	if endIndex == -1 {
		return nil, content, fmt.Errorf("frontmatter not closed")
	}

	fmContent := content[3 : endIndex+3]
	body := strings.TrimSpace(content[endIndex+6:])

	// Parsear el frontmatter manualmente (es YAML-like pero con JSON arrays)
	fm := &frontmatter{}

	// Extraer id
	idMatch := regexp.MustCompile(`id:\s*['"]([^'"]+)['"]`).FindStringSubmatch(fmContent)
	if len(idMatch) > 1 {
		fm.ID = idMatch[1]
	}

	// Extraer order
	orderMatch := regexp.MustCompile(`order:\s*(\d+)`).FindStringSubmatch(fmContent)
	if len(orderMatch) > 1 {
		fm.Order, _ = strconv.Atoi(orderMatch[1])
	}

	// Extraer name
	nameMatch := regexp.MustCompile(`name:\s*['"]([^'"]+)['"]`).FindStringSubmatch(fmContent)
	if len(nameMatch) > 1 {
		fm.Name = nameMatch[1]
	}

	// Extraer titleList (es un array JSON-like)
	titleListStart := strings.Index(fmContent, "titleList:")
	if titleListStart != -1 {
		// Encontrar el array completo
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
				// Limpiar el contenido para que sea JSON válido
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

// cleanArrayToJSON limpia el array YAML-like para que sea JSON válido
func (p *Parser) cleanArrayToJSON(content string) string {
	// Reemplazar comillas simples por dobles
	content = strings.ReplaceAll(content, "'", "\"")

	// Asegurar que las keys estén entre comillas
	content = regexp.MustCompile(`(\s)name:`).ReplaceAllString(content, `$1"name":`)
	content = regexp.MustCompile(`(\s)tagId:`).ReplaceAllString(content, `$1"tagId":`)
	content = regexp.MustCompile(`{\s*name:`).ReplaceAllString(content, `{"name":`)
	content = regexp.MustCompile(`{\s*tagId:`).ReplaceAllString(content, `{"tagId":`)

	// Limpiar espacios y saltos de línea extra
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	return content
}

// ListChapters lista todos los capítulos de un locale
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
			// Log error pero continuar con otros archivos
			fmt.Fprintf(os.Stderr, "Warning: could not parse %s: %v\n", filePath, err)
			continue
		}
		chapters = append(chapters, *chapter)
	}

	// Ordenar por order
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Order < chapters[j].Order
	})

	return chapters, nil
}

// GetChapter obtiene un capítulo específico por ID
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

// GetSection obtiene una sección específica de un capítulo
func (p *Parser) GetSection(chapterID string, sectionTagID string, locale string) (string, error) {
	chapter, err := p.GetChapter(chapterID, locale)
	if err != nil {
		return "", err
	}

	// Buscar la sección en el contenido
	lines := strings.Split(chapter.Content, "\n")

	// Encontrar el header que corresponde al tagId
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
				// Llegamos a otra sección, terminar
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

// generateTagID genera un tagId a partir de un título
func (p *Parser) generateTagID(title string) string {
	// Convertir a minúsculas
	tagID := strings.ToLower(title)

	// Reemplazar espacios por guiones
	tagID = strings.ReplaceAll(tagID, " ", "-")

	// Remover caracteres especiales excepto guiones y letras con acentos
	tagID = regexp.MustCompile(`[^\p{L}\p{N}-]`).ReplaceAllString(tagID, "")

	// Remover guiones múltiples
	tagID = regexp.MustCompile(`-+`).ReplaceAllString(tagID, "-")

	// Remover guiones al inicio y final
	tagID = strings.Trim(tagID, "-")

	return tagID
}

// Search busca contenido en el libro
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

			// Actualizar sección actual
			if matches := headerPattern.FindStringSubmatch(line); len(matches) > 1 {
				currentSection = matches[1]
			}

			// Buscar coincidencias
			matchCount := 0
			for _, word := range queryWords {
				if strings.Contains(lineLower, word) {
					matchCount++
				}
			}

			if matchCount > 0 {
				relevance := float64(matchCount) / float64(len(queryWords))

				// Crear snippet con contexto
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

	// Ordenar por relevancia
	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance > results[j].Relevance
	})

	// Limitar resultados
	if len(results) > 20 {
		results = results[:20]
	}

	return results, nil
}

// GetBookIndex obtiene el índice completo del libro
func (p *Parser) GetBookIndex(locale string) (*BookIndex, error) {
	chapters, err := p.ListChapters(locale)
	if err != nil {
		return nil, err
	}

	// Limpiar el contenido para el índice (solo metadata)
	for i := range chapters {
		chapters[i].Content = "" // No incluir contenido completo en el índice
	}

	return &BookIndex{
		Locale:        locale,
		TotalChapters: len(chapters),
		Chapters:      chapters,
	}, nil
}

// GetAvailableLocales retorna los locales disponibles
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
