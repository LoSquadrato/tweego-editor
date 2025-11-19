# Tweego Editor

A modern and powerful editor for Twine interactive stories, based on [Tweego](https://www.motoslave.net/tweego/).

## 🎯 Project Goals

Tweego Editor was created to simplify the lives of interactive fiction authors by offering:

- 🔍 **On-the-fly preview**: Hover over passages to see a content preview without opening them
- 🎨 **Visual coding**: Automatically color passages based on tags to better organize your story
- 🔬 **Variable debugging**: Track how variables change along a specific path
- 📊 **Graph view**: Visualize your story structure as an interactive graph
- ⚡ **Performance**: Lightweight backend written in Go for speed and reliability

## 🏗️ Architecture

The project is divided into two main components:

### Backend (Go)
- Parser for `.twee` files (Tweego format)
- Multi-format support (Harlowe, SugarCube, Chapbook)
- API for the frontend editor
- File watcher for automatic recompilation
- Path simulator for variable debugging

### Frontend (In development)
- Visual editor with node-graph
- Hover preview of passages
- Path simulator for variable tracking
- Export/import compatible with Twine 2

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ ([Install Go](https://go.dev/doc/install))
- Tweego ([Install Tweego](https://www.motoslave.net/tweego/))

### Installation

```bash
# Clone the repository
git clone https://github.com/LoSquadrato/tweego-editor.git
cd tweego-editor

# Initialize Go module
go mod init tweego-editor
go mod tidy

# Test the parser
go run main.go
```

### Usage Example

```bash
# Parse a .twee file
go run main.go test_story.twee
```

## 📁 Project Structure

```
tweego-editor/
├── main.go                    # Entry point
├── parser/
│   ├── passage.go            # Data structures
│   └── twee_parser.go        # .twee parser
├── formats/
│   ├── interface.go          # Interface for story formats
│   └── harlowe/
│       └── parser.go         # Harlowe implementation
├── compiler/
│   └── tweego.go             # Tweego wrapper (TODO)
├── watcher/
│   └── file_watcher.go       # File monitoring (TODO)
└── api/
    └── server.go             # REST/WebSocket API (TODO)
```

## 🎨 Supported Story Formats

- ✅ **Harlowe** (in development)
- 🔜 **SugarCube** (planned)
- 🔜 **Chapbook** (planned)

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

### Workflow

1. Fork the project
2. Create a branch for your feature (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 TODO & Roadmap

### v0.1.0 (Current)
- [x] Basic parser for `.twee` files
- [x] Data structures for Story and Passage
- [x] Basic support for Harlowe
- [ ] Unit tests for parser
- [ ] Tweego compiler wrapper

### v0.2.0
- [ ] File watcher with hot-reload
- [ ] REST API for frontend
- [ ] SugarCube parser implementation
- [ ] Path simulator for variable debugging

### v0.3.0
- [ ] React frontend with node-graph
- [ ] Hover preview of passages
- [ ] Integrated editor
- [ ] Export to Twine 2

### v1.0.0
- [ ] Standalone Electron application
- [ ] Complete multi-format support
- [ ] Plugin system
- [ ] Complete documentation

## 📖 Documentation

- [.twee Format](https://github.com/iftechfoundation/twine-specs/blob/master/twee-3-specification.md)
- [Tweego Documentation](https://www.motoslave.net/tweego/docs/)
- [Harlowe Manual](https://twine2.neocities.org/)

## 📜 License

This project is distributed under the MIT License. See the `LICENSE` file for details.

## 👥 Authors

- **Nicola Zaramella (LoSquadrato)** - *Initial work*

See also the list of [contributors](https://github.com/LoSquadrato/tweego-editor/graphs/contributors) who participated in this project.

## 🙏 Acknowledgments

- Chris Klimas for [Twine](https://twinery.org/)
- Thomas Michael Edwards for [Tweego](https://www.motoslave.net/tweego/)
- The Interactive Fiction community

---

**Note**: This project is in active development. Expect breaking changes until v1.0.0!

## 💬 Community & Support

- 🐛 Found a bug? [Open an issue](https://github.com/LoSquadrato/tweego-editor/issues)
- 💡 Have an idea? [Start a discussion](https://github.com/LoSquadrato/tweego-editor/discussions)
- 💬 Want to chat? [Join our Discord](#) (TODO)

---

Made with ❤️ for the Interactive Fiction community