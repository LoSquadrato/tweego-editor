# Tweego Editor

A modern and powerful editor for Twine interactive stories, based on [Tweego](https://www.motoslave.net/tweego/).


## 🎯 Project Goals

Tweego Editor was created to simplify the lives of interactive fiction authors by offering:

- 🔍 **On-the-fly preview**: Hover over passages to see a content preview without opening them
- 🎨 **Visual coding**: Automatically color passages based on tags to better organize your story
- 🔬 **Variable debugging**: Track how variables change along a specific path
- 📊 **Graph view**: Visualize your story structure as an interactive graph
- ⚡ **Performance**: Lightweight backend written in Go for speed and reliability
- 🌐 **REST API**: Complete API with WebSocket support for real-time updates

## 🏗️ Architecture

The project is divided into two main components:

### Backend (Go) - ✅ Complete
- Parser for `.twee` files (Tweego format)
- Multi-format support (Harlowe, SugarCube, Chapbook)
- Tweego compiler wrapper
- File watcher with auto-recompilation
- REST API + WebSocket for real-time events
- Path simulator for variable debugging (coming soon)

### Frontend (In development)
- Visual editor with node-graph
- Hover preview of passages
- Path simulator for variable tracking
- Export/import compatible with Twine 2

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** ([Install Go](https://go.dev/doc/install))
- **Tweego** ([Install Tweego](https://www.motoslave.net/tweego/))

### Installation

```bash
# Clone the repository
git clone https://github.com/LoSquadrato/tweego-editor.git
cd tweego-editor

# Initialize Go module
go mod init tweego-editor

# Install dependencies
go get github.com/fsnotify/fsnotify
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
go get github.com/gorilla/websocket

# Download dependencies
go mod tidy

# Run the application
go run main.go
```

### Installing Tweego (Linux/WSL)

```bash
# Download Tweego
cd ~
wget https://github.com/tmedwards/tweego/releases/download/v2.1.1/tweego-2.1.1-linux-x64.zip

# Install unzip if needed
sudo apt install unzip

# Extract
unzip tweego-2.1.1-linux-x64.zip

# Move to PATH
sudo mv tweego /usr/local/bin/
sudo chmod +x /usr/local/bin/tweego

# Verify installation
tweego --version
```

## 📖 Usage

The application offers 4 modes:

### 1. Parser Test
Parse a `.twee` file and extract all information:
```bash
go run main.go
# Select option: 1
```

### 2. Compiler Test
Compile a `.twee` file to HTML:
```bash
go run main.go
# Select option: 2
```

### 3. Watch Mode
Auto-recompile on file changes:
```bash
go run main.go
# Select option: 3
```

### 4. API Server
Start the REST API server:
```bash
go run main.go
# Select option: 4
```

Server will be available at `http://localhost:8080`

## 🌐 API Documentation

### Available Endpoints

#### Health Check
```bash
GET /api/health
```
Response:
```json
{
  "status": "ok",
  "version": "0.1.0"
}
```

#### Parse Story
```bash
POST /api/story/parse
Content-Type: application/json

{
  "file_path": "test_story.twee"
}
```

Response includes passages with links, variables, and clean previews.

#### Compile Story
```bash
POST /api/story/compile
Content-Type: application/json

{
  "file_path": "test_story.twee",
  "format": "harlowe-3",
  "output": "output.html"
}
```

#### Get All Passages
```bash
GET /api/story/:file/passages
```

#### Get Single Passage
```bash
GET /api/story/:file/passage/:title
```

#### Start File Watcher
```bash
POST /api/watch/start
Content-Type: application/json

{
  "paths": ["."],
  "format": "harlowe-3",
  "output": "output.html",
  "auto_compile": true
}
```

#### Stop File Watcher
```bash
POST /api/watch/stop
```

#### Get Watcher Status
```bash
GET /api/watch/status
```

#### List Available Formats
```bash
GET /api/formats
```

Response:
```json
{
  "success": true,
  "formats": [
    "chapbook-1",
    "harlowe-1",
    "harlowe-2",
    "harlowe-3",
    "paperthin-1",
    "snowman-1",
    "snowman-2",
    "sugarcube-1",
    "sugarcube-2"
  ]
}
```

#### Get Tweego Version
```bash
GET /api/version
```

#### WebSocket Connection
```
WS /ws
```
Connect to receive real-time file watcher events.

### API Testing Examples

```bash
# Health check
curl http://localhost:8080/api/health

# Parse a story
curl -X POST http://localhost:8080/api/story/parse \
  -H "Content-Type: application/json" \
  -d '{"file_path": "test_story.twee"}'

# Compile a story
curl -X POST http://localhost:8080/api/story/compile \
  -H "Content-Type: application/json" \
  -d '{"file_path": "test_story.twee", "format": "harlowe-3", "output": "output.html"}'

# Get available formats
curl http://localhost:8080/api/formats

# Start watcher
curl -X POST http://localhost:8080/api/watch/start \
  -H "Content-Type: application/json" \
  -d '{"paths": ["."], "format": "harlowe-3", "auto_compile": true}'

# Check watcher status
curl http://localhost:8080/api/watch/status

# Stop watcher
curl -X POST http://localhost:8080/api/watch/stop
```

## 📁 Project Structure

```
tweego-editor/
├── main.go                    # Entry point with CLI menu
├── test_story.twee           # Example story for testing
├── parser/
│   ├── passage.go            # Data structures (Story, Passage)
│   └── twee_parser.go        # .twee file parser
├── formats/
│   ├── interface.go          # StoryFormat interface
│   └── harlowe/
│       └── parser.go         # Harlowe implementation
├── compiler/
│   └── tweego.go             # Tweego wrapper
├── watcher/
│   └── file_watcher.go       # File monitoring with auto-compile
├── api/
│   └── server.go             # REST API + WebSocket server
└── output/                   # Compiled HTML files (gitignored)
```

## 🎨 Supported Story Formats

- ✅ **Harlowe** (1.2.4, 2.1.0, 3.1.0)
- ✅ **SugarCube** (1.0.35, 2.30.0)
- ✅ **Chapbook** (1.0.0)
- ✅ **Snowman** (1.4.0, 2.0.2)
- ✅ **Paperthin** (1.0.0) [proofing]

Each format implements the `StoryFormat` interface:
- `ParseVariables()` - Extract variables from content
- `ParseLinks()` - Extract links between passages
- `StripCode()` - Remove code for clean preview
- `GetFormatName()` - Return format name

## 🤝 How to Contribute

Contributions, issues, and feature requests are welcome!

### Areas Where Help is Needed

- 🎨 **Frontend Developer**: React/Electron for the UI
- 🔧 **Parser Development**: Improve support for SugarCube and Chapbook
- 📚 **Documentation**: Examples, tutorials, guides
- 🐛 **Testing**: Unit tests, integration tests
- 🌍 **Translation**: Localization support

### Workflow

1. Fork the project
2. Create a branch for your feature (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 🐛 Troubleshooting

### Tweego not found
```bash
# Verify Tweego is in PATH
which tweego
tweego --version

# If not found, install following the instructions above
```

### Port 8080 already in use
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change port in api/server.go
```

### API returns "exit status 1" errors
This is often normal for Tweego commands that write to stderr. The latest version handles this correctly.

## 📝 TODO & Roadmap

### v0.1.0 ✅ (Current)
- [x] Basic parser for `.twee` files
- [x] Data structures for Story and Passage
- [x] Basic support for Harlowe
- [x] Tweego compiler wrapper
- [x] File watcher with hot-reload
- [x] REST API with 10+ endpoints
- [x] WebSocket for real-time events

### v0.2.0 (In Progress)
- [ ] Unit tests for parser and compiler
- [ ] Enhanced SugarCube parser
- [ ] Enhanced Chapbook parser
- [ ] Path simulator for variable debugging
- [ ] API authentication
- [ ] Rate limiting

### v0.3.0
- [ ] React frontend with node-graph
- [ ] Hover preview of passages
- [ ] Integrated editor
- [ ] Export to Twine 2
- [ ] Dark mode

### v1.0.0
- [ ] Standalone Electron application
- [ ] Complete multi-format support
- [ ] Plugin system
- [ ] Complete documentation
- [ ] Internationalization (i18n)

## 📖 Documentation

- [.twee Format Specification](https://github.com/iftechfoundation/twine-specs/blob/master/twee-3-specification.md)
- [Tweego Documentation](https://www.motoslave.net/tweego/docs/)
- [Harlowe Manual](https://twine2.neocities.org/)
- [SugarCube Documentation](http://www.motoslave.net/sugarcube/2/docs/)

## 📜 License

This project is distributed under the MIT License. See the `LICENSE` file for details.

**Note**: This project uses [Tweego](https://www.motoslave.net/tweego/) which is licensed under the BSD 2-Clause License.

## 👥 Authors

- **LoSquadrato** - *Initial work*

See also the list of [contributors](https://github.com/LoSquadrato/tweego-editor/contributors) who participated in this project.

## 🙏 Acknowledgments

- Chris Klimas for [Twine](https://twinery.org/)
- Thomas Michael Edwards for [Tweego](https://www.motoslave.net/tweego/) (BSD 2-Clause License)
- The Interactive Fiction community
- All contributors and supporters

---

**Note**: This project is in active development. Expect breaking changes until v1.0.0!

## 💬 Community & Support

- 🐛 Found a bug? [Open an issue](https://github.com/LoSquadrato/tweego-editor/issues)
- 💡 Have an idea? [Start a discussion](https://github.com/LoSquadrato/tweego-editor/discussions)
- ⭐ Like the project? Give it a star!
- 🔄 Want updates? Watch the repository

---

Made with ❤️ for the Interactive Fiction community

## 📊 Project Stats

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-alpha-orange)
![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen)