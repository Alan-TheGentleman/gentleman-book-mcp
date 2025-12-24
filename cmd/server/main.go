package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Alan-TheGentleman/gentleman-book-mcp/internal/book"
	"github.com/Alan-TheGentleman/gentleman-book-mcp/internal/embeddings"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var parser *book.Parser
var semanticEngine *embeddings.SemanticEngine

func main() {
	// Obtener la ruta del libro desde variable de entorno o usar default
	bookPath := os.Getenv("BOOK_PATH")
	if bookPath == "" {
		// Default path relativo al proyecto gentleman-programming-book
		homeDir, _ := os.UserHomeDir()
		bookPath = homeDir + "/work/gentleman-programming-book/src/data/book"
	}

	// Verificar que el path existe
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		log.Fatalf("Book path does not exist: %s", bookPath)
	}

	parser = book.NewParser(bookPath)

	// Inicializar semantic engine si hay API key de OpenAI o Ollama
	initSemanticEngine()

	// Crear servidor MCP
	s := server.NewMCPServer(
		"Gentleman Programming Book",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
	)

	// ============================================
	// NIVEL 1: TOOLS BÁSICOS
	// ============================================

	// Tool: list_chapters
	s.AddTool(
		mcp.NewTool("list_chapters",
			mcp.WithDescription("List all chapters in the Gentleman Programming Book. Returns chapter metadata including ID, name, order, and sections."),
			mcp.WithString("locale",
				mcp.Description("Language locale: 'es' for Spanish, 'en' for English"),
				mcp.DefaultString("es"),
			),
		),
		handleListChapters,
	)

	// Tool: read_chapter
	s.AddTool(
		mcp.NewTool("read_chapter",
			mcp.WithDescription("Read a specific chapter from the book. Can read the entire chapter or a specific section."),
			mcp.WithString("chapter_id",
				mcp.Required(),
				mcp.Description("The chapter ID (e.g., 'clean-agile', 'hexagonal-architecture')"),
			),
			mcp.WithString("section_id",
				mcp.Description("Optional section tag ID to read only that section"),
			),
			mcp.WithString("locale",
				mcp.Description("Language locale: 'es' for Spanish, 'en' for English"),
				mcp.DefaultString("es"),
			),
		),
		handleReadChapter,
	)

	// Tool: search_book
	s.AddTool(
		mcp.NewTool("search_book",
			mcp.WithDescription("Search for content in the book using keywords. Returns relevant snippets with chapter and section information."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query (keywords to find in the book)"),
			),
			mcp.WithString("locale",
				mcp.Description("Language locale: 'es' for Spanish, 'en' for English"),
				mcp.DefaultString("es"),
			),
		),
		handleSearchBook,
	)

	// Tool: get_book_index
	s.AddTool(
		mcp.NewTool("get_book_index",
			mcp.WithDescription("Get the complete table of contents for the book, including all chapters and their sections."),
			mcp.WithString("locale",
				mcp.Description("Language locale: 'es' for Spanish, 'en' for English"),
				mcp.DefaultString("es"),
			),
		),
		handleGetBookIndex,
	)

	// ============================================
	// NIVEL 3: SEMANTIC SEARCH
	// ============================================

	// Tool: semantic_search (solo disponible si hay embeddings)
	s.AddTool(
		mcp.NewTool("semantic_search",
			mcp.WithDescription("Search the book using semantic similarity (AI-powered). More accurate than keyword search. Requires OPENAI_API_KEY or Ollama running locally."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Natural language query to search for"),
			),
			mcp.WithString("locale",
				mcp.Description("Language locale: 'es' for Spanish, 'en' for English"),
				mcp.DefaultString("es"),
			),
			mcp.WithNumber("top_k",
				mcp.Description("Number of results to return (default: 5)"),
			),
		),
		handleSemanticSearch,
	)

	// Tool: build_semantic_index
	s.AddTool(
		mcp.NewTool("build_semantic_index",
			mcp.WithDescription("Build or rebuild the semantic search index. Required before using semantic_search. Takes a few minutes."),
			mcp.WithString("locale",
				mcp.Description("Language locale to index: 'es', 'en', or 'all'"),
				mcp.DefaultString("all"),
			),
		),
		handleBuildSemanticIndex,
	)

	// Tool: semantic_status
	s.AddTool(
		mcp.NewTool("semantic_status",
			mcp.WithDescription("Check the status of the semantic search engine (availability, index status, chunk count)."),
		),
		handleSemanticStatus,
	)

	// ============================================
	// NIVEL 2: RESOURCES DINÁMICOS
	// ============================================

	// Resource: Book index
	s.AddResource(
		mcp.NewResource(
			"book://index/es",
			"Book Index (Spanish)",
			mcp.WithResourceDescription("Complete table of contents for the Spanish version"),
			mcp.WithMIMEType("application/json"),
		),
		handleBookIndexResource,
	)

	s.AddResource(
		mcp.NewResource(
			"book://index/en",
			"Book Index (English)",
			mcp.WithResourceDescription("Complete table of contents for the English version"),
			mcp.WithMIMEType("application/json"),
		),
		handleBookIndexResource,
	)

	// ============================================
	// NIVEL 2: PROMPTS PREDEFINIDOS
	// ============================================

	// Prompt: explain_concept
	s.AddPrompt(
		mcp.NewPrompt("explain_concept",
			mcp.WithPromptDescription("Ask the AI to explain a concept from the Gentleman Programming Book"),
			mcp.WithArgument("concept",
				mcp.ArgumentDescription("The concept to explain (e.g., 'hexagonal architecture', 'clean architecture', 'TDD')"),
			),
			mcp.WithArgument("locale",
				mcp.ArgumentDescription("Language: 'es' or 'en'"),
			),
		),
		handleExplainConceptPrompt,
	)

	// Prompt: compare_patterns
	s.AddPrompt(
		mcp.NewPrompt("compare_patterns",
			mcp.WithPromptDescription("Compare two architectural patterns or concepts from the book"),
			mcp.WithArgument("pattern_a",
				mcp.ArgumentDescription("First pattern to compare"),
			),
			mcp.WithArgument("pattern_b",
				mcp.ArgumentDescription("Second pattern to compare"),
			),
		),
		handleComparePatternsPrompt,
	)

	// Prompt: summarize_chapter
	s.AddPrompt(
		mcp.NewPrompt("summarize_chapter",
			mcp.WithPromptDescription("Get a summary of a specific chapter from the book"),
			mcp.WithArgument("chapter_id",
				mcp.ArgumentDescription("The chapter ID to summarize"),
			),
			mcp.WithArgument("locale",
				mcp.ArgumentDescription("Language: 'es' or 'en'"),
			),
		),
		handleSummarizeChapterPrompt,
	)

	// Iniciar servidor via stdio
	log.Println("Starting Gentleman Book MCP Server...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================
// TOOL HANDLERS - NIVEL 1
// ============================================

func handleListChapters(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	locale := req.GetString("locale", "es")

	chapters, err := parser.ListChapters(locale)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error listing chapters: %v", err)), nil
	}

	// Crear resumen de capítulos (sin contenido completo)
	type chapterSummary struct {
		ID       string         `json:"id"`
		Order    int            `json:"order"`
		Name     string         `json:"name"`
		Sections []book.Section `json:"sections"`
	}

	var summaries []chapterSummary
	for _, ch := range chapters {
		summaries = append(summaries, chapterSummary{
			ID:       ch.ID,
			Order:    ch.Order,
			Name:     ch.Name,
			Sections: ch.TitleList,
		})
	}

	result, _ := json.MarshalIndent(summaries, "", "  ")
	return mcp.NewToolResultText(string(result)), nil
}

func handleReadChapter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	chapterID := req.GetString("chapter_id", "")
	sectionID := req.GetString("section_id", "")
	locale := req.GetString("locale", "es")

	if chapterID == "" {
		return mcp.NewToolResultError("chapter_id is required"), nil
	}

	if sectionID != "" {
		// Leer solo la sección
		content, err := parser.GetSection(chapterID, sectionID, locale)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error reading section: %v", err)), nil
		}
		return mcp.NewToolResultText(content), nil
	}

	// Leer capítulo completo
	chapter, err := parser.GetChapter(chapterID, locale)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading chapter: %v", err)), nil
	}

	// Formatear respuesta
	response := fmt.Sprintf("# %s\n\n%s", chapter.Name, chapter.Content)
	return mcp.NewToolResultText(response), nil
}

func handleSearchBook(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.GetString("query", "")
	locale := req.GetString("locale", "es")

	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	results, err := parser.Search(query, locale)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No results found for: " + query), nil
	}

	resultJSON, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func handleGetBookIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	locale := req.GetString("locale", "es")

	index, err := parser.GetBookIndex(locale)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting index: %v", err)), nil
	}

	result, _ := json.MarshalIndent(index, "", "  ")
	return mcp.NewToolResultText(string(result)), nil
}

// ============================================
// RESOURCE HANDLERS - NIVEL 2
// ============================================

func handleBookIndexResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := req.Params.URI

	// Extraer locale del URI
	locale := "es"
	if strings.HasSuffix(uri, "/en") {
		locale = "en"
	}

	index, err := parser.GetBookIndex(locale)
	if err != nil {
		return nil, fmt.Errorf("error getting book index: %w", err)
	}

	indexJSON, _ := json.MarshalIndent(index, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(indexJSON),
		},
	}, nil
}

// ============================================
// PROMPT HANDLERS - NIVEL 2
// ============================================

func handleExplainConceptPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	concept := "architecture"
	locale := "es"

	if args := req.Params.Arguments; args != nil {
		if c := args["concept"]; c != "" {
			concept = c
		}
		if l := args["locale"]; l != "" {
			locale = l
		}
	}

	// Buscar contenido relevante en el libro
	results, _ := parser.Search(concept, locale)

	var contextSnippets string
	if len(results) > 0 {
		var snippets []string
		for i, r := range results {
			if i >= 5 { // Máximo 5 snippets
				break
			}
			snippets = append(snippets, fmt.Sprintf("From '%s' (%s):\n%s", r.ChapterName, r.Section, r.Snippet))
		}
		contextSnippets = strings.Join(snippets, "\n\n---\n\n")
	}

	promptText := fmt.Sprintf(`Based on the Gentleman Programming Book, explain the concept of "%s".

Here is relevant content from the book:

%s

Please provide a clear and comprehensive explanation based on this content.`, concept, contextSnippets)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Explain '%s' from the Gentleman Programming Book", concept),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(promptText),
			},
		},
	}, nil
}

func handleComparePatternsPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	patternA := "clean architecture"
	patternB := "hexagonal architecture"

	if args := req.Params.Arguments; args != nil {
		if a := args["pattern_a"]; a != "" {
			patternA = a
		}
		if b := args["pattern_b"]; b != "" {
			patternB = b
		}
	}

	// Buscar contenido de ambos patrones
	resultsA, _ := parser.Search(patternA, "es")
	resultsB, _ := parser.Search(patternB, "es")

	var contextA, contextB string
	if len(resultsA) > 0 {
		var snippets []string
		for i, r := range resultsA {
			if i >= 3 {
				break
			}
			snippets = append(snippets, r.Snippet)
		}
		contextA = strings.Join(snippets, "\n")
	}
	if len(resultsB) > 0 {
		var snippets []string
		for i, r := range resultsB {
			if i >= 3 {
				break
			}
			snippets = append(snippets, r.Snippet)
		}
		contextB = strings.Join(snippets, "\n")
	}

	promptText := fmt.Sprintf(`Compare and contrast "%s" and "%s" based on the Gentleman Programming Book.

Content about %s:
%s

Content about %s:
%s

Please provide a detailed comparison including:
1. Key differences
2. Similarities
3. When to use each one
4. Pros and cons`, patternA, patternB, patternA, contextA, patternB, contextB)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Compare '%s' vs '%s'", patternA, patternB),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(promptText),
			},
		},
	}, nil
}

func handleSummarizeChapterPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	chapterID := ""
	locale := "es"

	if args := req.Params.Arguments; args != nil {
		if id := args["chapter_id"]; id != "" {
			chapterID = id
		}
		if l := args["locale"]; l != "" {
			locale = l
		}
	}

	if chapterID == "" {
		return &mcp.GetPromptResult{
			Description: "Error: chapter_id is required",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.NewTextContent("Please provide a chapter_id to summarize."),
				},
			},
		}, nil
	}

	chapter, err := parser.GetChapter(chapterID, locale)
	if err != nil {
		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Error: %v", err),
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.NewTextContent(fmt.Sprintf("Could not find chapter: %s", chapterID)),
				},
			},
		}, nil
	}

	// Limitar contenido si es muy largo
	content := chapter.Content
	if len(content) > 10000 {
		content = content[:10000] + "\n\n... [content truncated]"
	}

	promptText := fmt.Sprintf(`Please provide a comprehensive summary of the following chapter from the Gentleman Programming Book:

# %s

%s

Include:
1. Main concepts covered
2. Key takeaways
3. Practical applications`, chapter.Name, content)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Summary of '%s'", chapter.Name),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(promptText),
			},
		},
	}, nil
}

// ============================================
// SEMANTIC SEARCH HANDLERS - NIVEL 3
// ============================================

func initSemanticEngine() {
	// Intentar OpenAI primero, luego Ollama
	var err error

	if os.Getenv("OPENAI_API_KEY") != "" {
		semanticEngine, err = embeddings.NewSemanticEngine(embeddings.ProviderOpenAI)
		if err == nil {
			log.Println("Semantic search enabled with OpenAI")
			return
		}
		log.Printf("OpenAI not available: %v", err)
	}

	// Intentar Ollama
	semanticEngine, err = embeddings.NewSemanticEngine(embeddings.ProviderOllama)
	if err == nil && semanticEngine.IsAvailable() {
		log.Println("Semantic search enabled with Ollama")
		return
	}

	log.Println("Semantic search not available (no OpenAI key or Ollama)")
	semanticEngine = nil
}

func handleSemanticSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if semanticEngine == nil {
		return mcp.NewToolResultError("Semantic search not available. Set OPENAI_API_KEY or ensure Ollama is running."), nil
	}

	if !semanticEngine.IsIndexed() {
		return mcp.NewToolResultError("Semantic index not built. Run 'build_semantic_index' first."), nil
	}

	query := req.GetString("query", "")
	locale := req.GetString("locale", "es")
	topK := req.GetInt("top_k", 5)

	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	results, err := semanticEngine.Search(ctx, query, locale, topK)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search error: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No semantic matches found for: " + query), nil
	}

	resultJSON, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func handleBuildSemanticIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if semanticEngine == nil {
		return mcp.NewToolResultError("Semantic search not available. Set OPENAI_API_KEY or ensure Ollama is running."), nil
	}

	localeParam := req.GetString("locale", "all")

	var locales []string
	if localeParam == "all" {
		locales = []string{"es", "en"}
	} else {
		locales = []string{localeParam}
	}

	var allChunks []embeddings.Chunk
	chunkID := 0

	for _, locale := range locales {
		chapters, err := parser.ListChapters(locale)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error reading chapters for %s: %v", locale, err)), nil
		}

		for _, chapter := range chapters {
			// Dividir el contenido en chunks (por secciones o párrafos)
			chunks := splitIntoChunks(chapter.Content, chapter.ID, chapter.Name, locale, &chunkID)
			allChunks = append(allChunks, chunks...)
		}
	}

	log.Printf("Indexing %d chunks...", len(allChunks))

	if err := semanticEngine.IndexChunks(ctx, allChunks); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error indexing: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully indexed %d chunks from %d locale(s)", len(allChunks), len(locales))), nil
}

func handleSemanticStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := map[string]interface{}{
		"available": semanticEngine != nil,
		"indexed":   false,
		"chunks":    0,
		"provider":  "none",
	}

	if semanticEngine != nil {
		status["indexed"] = semanticEngine.IsIndexed()
		status["chunks"] = semanticEngine.ChunkCount()

		if os.Getenv("OPENAI_API_KEY") != "" {
			status["provider"] = "openai"
		} else {
			status["provider"] = "ollama"
		}
	}

	result, _ := json.MarshalIndent(status, "", "  ")
	return mcp.NewToolResultText(string(result)), nil
}

// splitIntoChunks divide el contenido en chunks manejables
func splitIntoChunks(content string, chapterID, chapterName, locale string, idCounter *int) []embeddings.Chunk {
	var chunks []embeddings.Chunk

	// Dividir por secciones (headers ##)
	headerPattern := regexp.MustCompile(`(?m)^##\s+(.+)$`)
	sections := headerPattern.Split(content, -1)
	headers := headerPattern.FindAllStringSubmatch(content, -1)

	// Agregar contenido antes del primer header
	if len(sections) > 0 && strings.TrimSpace(sections[0]) != "" {
		*idCounter++
		chunks = append(chunks, embeddings.Chunk{
			ID:          fmt.Sprintf("chunk_%d", *idCounter),
			ChapterID:   chapterID,
			ChapterName: chapterName,
			Section:     "Introduction",
			Content:     truncateContent(strings.TrimSpace(sections[0]), 1000),
			Locale:      locale,
		})
	}

	// Procesar cada sección
	for i, header := range headers {
		sectionContent := ""
		if i+1 < len(sections) {
			sectionContent = strings.TrimSpace(sections[i+1])
		}

		if sectionContent == "" {
			continue
		}

		// Si el contenido es muy largo, dividirlo en chunks más pequeños
		sectionName := header[1]
		contentChunks := splitLongContent(sectionContent, 1000)

		for j, c := range contentChunks {
			*idCounter++
			suffix := ""
			if len(contentChunks) > 1 {
				suffix = fmt.Sprintf(" (part %d)", j+1)
			}
			chunks = append(chunks, embeddings.Chunk{
				ID:          fmt.Sprintf("chunk_%d", *idCounter),
				ChapterID:   chapterID,
				ChapterName: chapterName,
				Section:     sectionName + suffix,
				Content:     c,
				Locale:      locale,
			})
		}
	}

	return chunks
}

func splitLongContent(content string, maxLen int) []string {
	if len(content) <= maxLen {
		return []string{content}
	}

	var chunks []string
	paragraphs := strings.Split(content, "\n\n")
	current := ""

	for _, p := range paragraphs {
		if len(current)+len(p) > maxLen && current != "" {
			chunks = append(chunks, strings.TrimSpace(current))
			current = p
		} else {
			if current != "" {
				current += "\n\n"
			}
			current += p
		}
	}

	if current != "" {
		chunks = append(chunks, strings.TrimSpace(current))
	}

	return chunks
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
