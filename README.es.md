# 📚 Gentleman Book MCP Server

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/MCP-1.0-purple?style=for-the-badge" alt="MCP Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<p align="center">
  <b>Dale a los asistentes de IA acceso directo al Gentleman Programming Book</b>
</p>

<p align="center">
  <a href="#características">Características</a> •
  <a href="#instalación">Instalación</a> •
  <a href="#configuración">Configuración</a> •
  <a href="#uso">Uso</a> •
  <a href="./README.md">English</a>
</p>

---

## ¿Qué es esto?

Este es un **servidor MCP (Model Context Protocol)** que permite a asistentes de IA como Claude leer, buscar y entender el contenido del [Gentleman Programming Book](https://github.com/Alan-TheGentleman/gentleman-programming-book).

Pensalo como darle a tu asistente de IA una línea directa a 18 capítulos de conocimiento sobre arquitectura de software, buenas prácticas y sabiduría de desarrollo.

## Características

### 🔧 Nivel 1: Tools Básicos

| Tool             | Descripción                                 |
| ---------------- | ------------------------------------------- |
| `list_chapters`  | Lista los 18 capítulos con metadata         |
| `read_chapter`   | Lee cualquier capítulo o sección específica |
| `search_book`    | Búsqueda por keywords en todo el contenido  |
| `get_book_index` | Tabla de contenidos completa                |

### 📦 Nivel 2: Resources y Prompts

| Tipo     | Nombre              | Descripción                          |
| -------- | ------------------- | ------------------------------------ |
| Resource | `book://index/es`   | Índice en español                    |
| Resource | `book://index/en`   | Índice en inglés                     |
| Prompt   | `explain_concept`   | Explica cualquier concepto del libro |
| Prompt   | `compare_patterns`  | Compara patrones de arquitectura     |
| Prompt   | `summarize_chapter` | Obtiene resúmenes de capítulos       |

### 🧠 Nivel 3: Búsqueda Semántica (IA)

| Tool                   | Descripción                                    |
| ---------------------- | ---------------------------------------------- |
| `semantic_search`      | Búsqueda en lenguaje natural usando embeddings |
| `build_semantic_index` | Construye el índice vectorial                  |
| `semantic_status`      | Verifica el estado del motor semántico         |

**Soporta tanto OpenAI como Ollama** para generación de embeddings.

## Instalación

### Prerequisitos

- Go 1.21 o superior
- El [Gentleman Programming Book](https://github.com/Alan-TheGentleman/gentleman-programming-book) clonado localmente

### Compilar desde source

```bash
# Clonar este repositorio
git clone https://github.com/Alan-TheGentleman/gentleman-book-mcp.git
cd gentleman-book-mcp

# Compilar el binario
go build -o bin/gentleman-book-mcp ./cmd/server

# El binario está en ./bin/gentleman-book-mcp
```

### Verificar instalación

```bash
./bin/gentleman-book-mcp --help
```

## Configuración

### Variables de Entorno

| Variable                 | Descripción                                 | Default                                           |
| ------------------------ | ------------------------------------------- | ------------------------------------------------- |
| `BOOK_PATH`              | Ruta a los archivos MDX del libro           | `~/work/gentleman-programming-book/src/data/book` |
| `OPENAI_API_KEY`         | API key de OpenAI (para búsqueda semántica) | -                                                 |
| `OLLAMA_BASE_URL`        | URL del servidor Ollama                     | `http://localhost:11434`                          |
| `OLLAMA_EMBEDDING_MODEL` | Modelo de Ollama para embeddings            | `nomic-embed-text`                                |

### Configuración en Claude Desktop

Agregar a `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) o `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "gentleman-book": {
      "command": "/ruta/absoluta/a/gentleman-book-mcp",
      "env": {
        "BOOK_PATH": "/ruta/a/gentleman-programming-book/src/data/book"
      }
    }
  }
}
```

### Con OpenAI (para búsqueda semántica)

```json
{
  "mcpServers": {
    "gentleman-book": {
      "command": "/ruta/absoluta/a/gentleman-book-mcp",
      "env": {
        "BOOK_PATH": "/ruta/a/gentleman-programming-book/src/data/book",
        "OPENAI_API_KEY": "sk-..."
      }
    }
  }
}
```

### Con Ollama (gratis, local)

1. Instalar [Ollama](https://ollama.ai)
2. Descargar un modelo de embeddings: `ollama pull nomic-embed-text`
3. Iniciar Ollama: `ollama serve`
4. Usar la configuración estándar (Ollama se auto-detecta)

## Uso

Una vez configurado, reiniciá Claude Desktop y empezá a chatear!

### Ejemplos de conversación

**Listar capítulos:**

```
Vos: ¿Qué capítulos tiene el Gentleman Programming Book?
Claude: [Usa list_chapters] El libro tiene 18 capítulos cubriendo...
```

**Leer contenido específico:**

```
Vos: Leeme el capítulo sobre arquitectura hexagonal
Claude: [Usa read_chapter] Acá está el capítulo de arquitectura hexagonal...
```

**Buscar temas:**

```
Vos: Buscá información sobre TDD en el libro
Claude: [Usa search_book] Encontré varias menciones de TDD...
```

**Búsqueda semántica (si está configurada):**

```
Vos: ¿Cómo debería estructurar una aplicación React para que sea mantenible?
Claude: [Usa semantic_search] Basándome en las recomendaciones del libro...
```

**Usar prompts:**

```
Vos: Explicame clean architecture según el libro
Claude: [Usa explain_concept prompt] Según el Gentleman Programming Book...
```

## Contenido del Libro

El servidor provee acceso a **18 capítulos** en inglés y español:

| #   | Capítulo                    | Temas                               |
| --- | --------------------------- | ----------------------------------- |
| 1   | Clean Agile                 | Agile, Waterfall, XP, TDD           |
| 2   | Comunicación                | Trabajo remoto, Dinámica de equipos |
| 3   | Arquitectura Hexagonal      | Puertos, Adaptadores, Dominio       |
| 4   | GoLang                      | Fundamentos de Go                   |
| 5   | Guía de NVIM                | Setup y uso de Neovim               |
| 6   | Algoritmos                  | Big O, Búsqueda, Ordenamiento       |
| 7   | Clean Architecture          | Capas, Casos de Uso, Dominio        |
| 8   | Clean Architecture Frontend | Scope Rule, Patrones frontend       |
| 9   | React                       | Hooks, Estado, Composición          |
| 10  | TypeScript                  | Tipos, Interfaces, Patrones         |
| 11  | Frontend Radar              | Comparación de frameworks           |
| 12  | Angular                     | Componentes, Servicios, Testing     |
| 13  | Barrels                     | Organización de módulos             |
| 14  | Historia del Frontend       | Evolución de la web                 |
| 15  | Desarrollo con IA           | Claude Code, Workflows con IA       |
| 16  | Manual Frontend             | Testing, Seguridad, Performance     |
| 17  | Soft Skills                 | Liderazgo, Comunicación             |
| 18  | Arquitectura de Software    | Microservicios, Patrones            |

## Arquitectura

```
gentleman-book-mcp/
├── cmd/
│   └── server/
│       └── main.go              # Entry point del servidor MCP
├── internal/
│   ├── book/
│   │   ├── models.go            # Estructuras de datos
│   │   └── parser.go            # Parser de archivos MDX
│   └── embeddings/
│       └── embeddings.go        # Motor de búsqueda semántica
├── go.mod
├── go.sum
├── README.md                    # Documentación en inglés
└── README.es.md                 # Documentación en español
```

## Desarrollo

```bash
# Correr en modo desarrollo
go run ./cmd/server

# Compilar
go build -o bin/gentleman-book-mcp ./cmd/server

# Testear con MCP Inspector
npx @anthropic-ai/mcp-inspector ./bin/gentleman-book-mcp
```

## Troubleshooting

### "Book path does not exist"

Asegurate que la variable `BOOK_PATH` apunte a la ubicación correcta de los archivos MDX del libro.

### "Semantic search not available"

O seteá `OPENAI_API_KEY` o asegurate que Ollama esté corriendo con un modelo de embeddings instalado.

### El servidor no responde

Verificá que el binario tenga permisos de ejecución: `chmod +x ./bin/gentleman-book-mcp`

## Contribuir

¡Las contribuciones son bienvenidas! Podés:

- Reportar bugs
- Sugerir nuevas features
- Enviar pull requests

## Licencia

MIT License - ver [LICENSE](LICENSE) para detalles.

## Proyectos Relacionados

- [Gentleman Programming Book](https://github.com/Alan-TheGentleman/gentleman-programming-book) - El libro
- [Model Context Protocol](https://modelcontextprotocol.io) - Especificación MCP
- [mcp-go](https://github.com/mark3labs/mcp-go) - SDK de Go para MCP

---

<p align="center">
  Hecho con ❤️ por <a href="https://github.com/Alan-TheGentleman">Gentleman Programming</a>
</p>
