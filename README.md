# 📚 Gentleman Book MCP Server

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/MCP-1.0-purple?style=for-the-badge" alt="MCP Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<p align="center">
  <b>Give AI assistants direct access to the Gentleman Programming Book</b>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#usage">Usage</a> •
  <a href="./README.es.md">Español</a>
</p>

---

## What is this?

This is an **MCP (Model Context Protocol) server** that allows AI assistants like Claude to read, search, and understand content from the [Gentleman Programming Book](https://github.com/Alan-TheGentleman/gentleman-programming-book).

Think of it as giving your AI assistant a direct line to 18 chapters of software architecture knowledge, best practices, and development wisdom.

## Features

### 🔧 Level 1: Basic Tools

| Tool             | Description                             |
| ---------------- | --------------------------------------- |
| `list_chapters`  | List all 18 chapters with metadata      |
| `read_chapter`   | Read any chapter or specific section    |
| `search_book`    | Keyword-based search across all content |
| `get_book_index` | Complete table of contents              |

### 📦 Level 2: Resources & Prompts

| Type     | Name                | Description                       |
| -------- | ------------------- | --------------------------------- |
| Resource | `book://index/es`   | Spanish table of contents         |
| Resource | `book://index/en`   | English table of contents         |
| Prompt   | `explain_concept`   | Explain any concept from the book |
| Prompt   | `compare_patterns`  | Compare architectural patterns    |
| Prompt   | `summarize_chapter` | Get chapter summaries             |

### 🧠 Level 3: Semantic Search (AI-Powered)

| Tool                   | Description                              |
| ---------------------- | ---------------------------------------- |
| `semantic_search`      | Natural language search using embeddings |
| `build_semantic_index` | Build the vector index                   |
| `semantic_status`      | Check semantic engine status             |

**Supports both OpenAI and Ollama** for embeddings generation.

## Installation

### Prerequisites

- Go 1.21 or higher
- The [Gentleman Programming Book](https://github.com/Alan-TheGentleman/gentleman-programming-book) cloned locally

### Build from source

```bash
# Clone this repository
git clone https://github.com/Alan-TheGentleman/gentleman-book-mcp.git
cd gentleman-book-mcp

# Build the binary
go build -o bin/gentleman-book-mcp ./cmd/server

# The binary is now at ./bin/gentleman-book-mcp
```

### Verify installation

```bash
./bin/gentleman-book-mcp --help
```

## Configuration

### Environment Variables

| Variable                 | Description                          | Default                                           |
| ------------------------ | ------------------------------------ | ------------------------------------------------- |
| `BOOK_PATH`              | Path to book MDX files               | `~/work/gentleman-programming-book/src/data/book` |
| `OPENAI_API_KEY`         | OpenAI API key (for semantic search) | -                                                 |
| `OLLAMA_BASE_URL`        | Ollama server URL                    | `http://localhost:11434`                          |
| `OLLAMA_EMBEDDING_MODEL` | Ollama model for embeddings          | `nomic-embed-text`                                |

### Claude Desktop Setup

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "gentleman-book": {
      "command": "/absolute/path/to/gentleman-book-mcp",
      "env": {
        "BOOK_PATH": "/path/to/gentleman-programming-book/src/data/book"
      }
    }
  }
}
```

### With OpenAI (for semantic search)

```json
{
  "mcpServers": {
    "gentleman-book": {
      "command": "/absolute/path/to/gentleman-book-mcp",
      "env": {
        "BOOK_PATH": "/path/to/gentleman-programming-book/src/data/book",
        "OPENAI_API_KEY": "sk-..."
      }
    }
  }
}
```

### With Ollama (free, local)

1. Install [Ollama](https://ollama.ai)
2. Pull an embedding model: `ollama pull nomic-embed-text`
3. Start Ollama: `ollama serve`
4. Use the standard configuration (Ollama is auto-detected)

## Usage

Once configured, restart Claude Desktop and start chatting!

### Example conversations

**List chapters:**

```
You: What chapters are in the Gentleman Programming Book?
Claude: [Uses list_chapters] The book has 18 chapters covering...
```

**Read specific content:**

```
You: Read me the chapter about hexagonal architecture
Claude: [Uses read_chapter] Here's the hexagonal architecture chapter...
```

**Search for topics:**

```
You: Find information about TDD in the book
Claude: [Uses search_book] I found several mentions of TDD...
```

**Semantic search (if configured):**

```
You: How should I structure a React application for maintainability?
Claude: [Uses semantic_search] Based on the book's recommendations...
```

**Use prompts:**

```
You: Explain clean architecture according to the book
Claude: [Uses explain_concept prompt] According to the Gentleman Programming Book...
```

## Book Content

The server provides access to **18 chapters** in both English and Spanish:

| #   | Chapter                     | Topics                         |
| --- | --------------------------- | ------------------------------ |
| 1   | Clean Agile                 | Agile, Waterfall, XP, TDD      |
| 2   | Communication               | Remote work, Team dynamics     |
| 3   | Hexagonal Architecture      | Ports, Adapters, Domain        |
| 4   | GoLang                      | Go fundamentals                |
| 5   | NVIM Guide                  | Neovim setup and usage         |
| 6   | Algorithms                  | Big O, Search, Sort            |
| 7   | Clean Architecture          | Layers, Use Cases, Domain      |
| 8   | Clean Architecture Frontend | Scope Rule, Frontend patterns  |
| 9   | React                       | Hooks, State, Composition      |
| 10  | TypeScript                  | Types, Interfaces, Patterns    |
| 11  | Frontend Radar              | Framework comparison           |
| 12  | Angular                     | Components, Services, Testing  |
| 13  | Barrels                     | Module organization            |
| 14  | Frontend History            | Web evolution                  |
| 15  | AI-Driven Development       | Claude Code, AI workflows      |
| 16  | Frontend Manual             | Testing, Security, Performance |
| 17  | Soft Skills                 | Leadership, Communication      |
| 18  | Software Architecture       | Microservices, Patterns        |

## Architecture

```
gentleman-book-mcp/
├── cmd/
│   └── server/
│       └── main.go              # MCP server entry point
├── internal/
│   ├── book/
│   │   ├── models.go            # Data structures
│   │   └── parser.go            # MDX file parser
│   └── embeddings/
│       └── embeddings.go        # Semantic search engine
├── go.mod
├── go.sum
├── README.md                    # English documentation
└── README.es.md                 # Spanish documentation
```

## Development

```bash
# Run in development mode
go run ./cmd/server

# Build
go build -o bin/gentleman-book-mcp ./cmd/server

# Test with MCP Inspector
npx @anthropic-ai/mcp-inspector ./bin/gentleman-book-mcp
```

## Troubleshooting

### "Book path does not exist"

Make sure the `BOOK_PATH` environment variable points to the correct location of the book's MDX files.

### "Semantic search not available"

Either set `OPENAI_API_KEY` or ensure Ollama is running with an embedding model installed.

### Server not responding

Check that the binary has execute permissions: `chmod +x ./bin/gentleman-book-mcp`

## Contributing

Contributions are welcome! Feel free to:

- Report bugs
- Suggest new features
- Submit pull requests

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [Gentleman Programming Book](https://github.com/Alan-TheGentleman/gentleman-programming-book) - The book itself
- [Model Context Protocol](https://modelcontextprotocol.io) - MCP specification
- [mcp-go](https://github.com/mark3labs/mcp-go) - Go SDK for MCP

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/Alan-TheGentleman">Gentleman Programming</a>
</p>
