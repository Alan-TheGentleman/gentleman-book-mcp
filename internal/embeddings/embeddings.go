package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
)

// Provider define el tipo de proveedor de embeddings
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderOllama Provider = "ollama"
)

// EmbeddingClient es la interfaz para generar embeddings
type EmbeddingClient interface {
	Embed(ctx context.Context, text string) ([]float64, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}

// Chunk representa un fragmento de texto con su embedding
type Chunk struct {
	ID          string    `json:"id"`
	ChapterID   string    `json:"chapterId"`
	ChapterName string    `json:"chapterName"`
	Section     string    `json:"section"`
	Content     string    `json:"content"`
	Embedding   []float64 `json:"embedding"`
	Locale      string    `json:"locale"`
}

// SemanticResult representa un resultado de búsqueda semántica
type SemanticResult struct {
	ChapterID   string  `json:"chapterId"`
	ChapterName string  `json:"chapterName"`
	Section     string  `json:"section"`
	Content     string  `json:"content"`
	Score       float64 `json:"score"`
	Locale      string  `json:"locale"`
}

// VectorStore almacena y busca chunks por similitud
type VectorStore struct {
	chunks []Chunk
	mu     sync.RWMutex
}

// NewVectorStore crea un nuevo vector store
func NewVectorStore() *VectorStore {
	return &VectorStore{
		chunks: make([]Chunk, 0),
	}
}

// Add agrega un chunk al store
func (v *VectorStore) Add(chunk Chunk) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.chunks = append(v.chunks, chunk)
}

// AddBatch agrega múltiples chunks
func (v *VectorStore) AddBatch(chunks []Chunk) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.chunks = append(v.chunks, chunks...)
}

// Search busca los chunks más similares a un embedding
func (v *VectorStore) Search(queryEmbedding []float64, locale string, topK int) []SemanticResult {
	v.mu.RLock()
	defer v.mu.RUnlock()

	type scored struct {
		chunk Chunk
		score float64
	}

	var results []scored
	for _, chunk := range v.chunks {
		if locale != "" && chunk.Locale != locale {
			continue
		}
		score := cosineSimilarity(queryEmbedding, chunk.Embedding)
		results = append(results, scored{chunk: chunk, score: score})
	}

	// Ordenar por score descendente
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Tomar top K
	if len(results) > topK {
		results = results[:topK]
	}

	var semanticResults []SemanticResult
	for _, r := range results {
		semanticResults = append(semanticResults, SemanticResult{
			ChapterID:   r.chunk.ChapterID,
			ChapterName: r.chunk.ChapterName,
			Section:     r.chunk.Section,
			Content:     r.chunk.Content,
			Score:       r.score,
			Locale:      r.chunk.Locale,
		})
	}

	return semanticResults
}

// Count retorna el número de chunks
func (v *VectorStore) Count() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.chunks)
}

// Clear limpia el store
func (v *VectorStore) Clear() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.chunks = make([]Chunk, 0)
}

// cosineSimilarity calcula la similitud de coseno entre dos vectores
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ============================================
// OPENAI CLIENT
// ============================================

type OpenAIClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

type openAIRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type openAIResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewOpenAIClient crea un cliente de OpenAI
func NewOpenAIClient(apiKey string) *OpenAIClient {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &OpenAIClient{
		apiKey: apiKey,
		model:  "text-embedding-3-small",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OpenAIClient) Embed(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

func (c *OpenAIClient) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not set")
	}

	reqBody := openAIRequest{
		Input: texts,
		Model: c.model,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, err
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI error: %s", openAIResp.Error.Message)
	}

	// Ordenar por index
	embeddings := make([][]float64, len(texts))
	for _, d := range openAIResp.Data {
		embeddings[d.Index] = d.Embedding
	}

	return embeddings, nil
}

// ============================================
// OLLAMA CLIENT
// ============================================

type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaResponse struct {
	Embedding []float64 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

// NewOllamaClient crea un cliente de Ollama
func NewOllamaClient(baseURL string, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
	}
	if model == "" {
		model = os.Getenv("OLLAMA_EMBEDDING_MODEL")
		if model == "" {
			model = "nomic-embed-text"
		}
	}
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *OllamaClient) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama connection error: %w (is Ollama running?)", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, err
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("Ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Embedding, nil
}

func (c *OllamaClient) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	// Ollama no soporta batch nativo, hacemos secuencial
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		emb, err := c.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("error embedding text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

// ============================================
// SEMANTIC ENGINE
// ============================================

// SemanticEngine combina el cliente de embeddings con el vector store
type SemanticEngine struct {
	client     EmbeddingClient
	store      *VectorStore
	isIndexed  bool
	indexMutex sync.Mutex
}

// NewSemanticEngine crea un nuevo motor semántico
func NewSemanticEngine(provider Provider) (*SemanticEngine, error) {
	var client EmbeddingClient

	switch provider {
	case ProviderOpenAI:
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY not set")
		}
		client = NewOpenAIClient(apiKey)
	case ProviderOllama:
		client = NewOllamaClient("", "")
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	return &SemanticEngine{
		client:    client,
		store:     NewVectorStore(),
		isIndexed: false,
	}, nil
}

// IsAvailable verifica si el motor está disponible
func (e *SemanticEngine) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.client.Embed(ctx, "test")
	return err == nil
}

// IndexChunks indexa una lista de chunks
func (e *SemanticEngine) IndexChunks(ctx context.Context, chunks []Chunk) error {
	e.indexMutex.Lock()
	defer e.indexMutex.Unlock()

	// Extraer textos
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	// Generar embeddings en batches de 100
	batchSize := 100
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		embeddings, err := e.client.EmbedBatch(ctx, texts[i:end])
		if err != nil {
			return fmt.Errorf("error generating embeddings: %w", err)
		}

		for j, emb := range embeddings {
			chunks[i+j].Embedding = emb
		}
	}

	e.store.AddBatch(chunks)
	e.isIndexed = true

	return nil
}

// Search realiza una búsqueda semántica
func (e *SemanticEngine) Search(ctx context.Context, query string, locale string, topK int) ([]SemanticResult, error) {
	if !e.isIndexed {
		return nil, fmt.Errorf("index not built, call IndexChunks first")
	}

	queryEmbedding, err := e.client.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	return e.store.Search(queryEmbedding, locale, topK), nil
}

// IsIndexed retorna si el índice está construido
func (e *SemanticEngine) IsIndexed() bool {
	return e.isIndexed
}

// ChunkCount retorna el número de chunks indexados
func (e *SemanticEngine) ChunkCount() int {
	return e.store.Count()
}
