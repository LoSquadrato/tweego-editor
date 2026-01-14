# Tweego Editor

## ⚠️ PROJECT STATUS: SUSPENDED FOR REVISION

**Last Updated**: January 2026  
**Status**: On Hold  
**Reason**: Strategic Pivot  

---

## 📋 Executive Summary

This project was originally conceived as a **debugging-first backend tool** for Twine interactive fiction authors, focusing on:
- PathSimulator for variable tracking across story paths
- Conditional validation
- Pre-execution debugging

### We are pivoting:

to **[FormatBridge for Twine](https://github.com/LoSquadrato/formatbridge)** - an automated story format converter.

---

## 🎯 Original Vision

### Core Features (Planned)
1. **PathSimulator**: Debug variables through specific story paths
2. **Conditional Validation**: Verify path accessibility given conditions
3. **Variable Tracking**: Type-specific change tracking
4. **Multi-format Support**: Harlowe, SugarCube, Chapbook
5. **REST API**: Integration with any editor

### What Was Built

#### ✅ Completed Components
- Unified Twee parser (format-agnostic)
- Tweego compiler wrapper
- File watcher with auto-recompilation
- REST API with WebSocket support
- Basic Harlowe format support: (Partially developed)
  - Link parsing
  - Variable extraction
  - Literal parsing (arrays, datamaps, datasets)
  - Property access evaluation
  - Basic conditionals

## 📚 Useful Resources from This Project

### Reusable Components
The following work can be adapted for FormatBridge:

1. **Twee Parser** (`parser/twee_parser.go`)
   - Format-agnostic structure parsing
   - StoryData extraction
   - Can be ported to TypeScript

2. **Harlowe Parser** (`formats/harlowe/`)
   - Macro parsing logic
   - Literal parsing (arrays, datamaps)
   - Useful as reference for TypeScript implementation

3. **Operations Validation** (concept)
   - Type-safe operation checking
   - Can be implemented in TypeScript with better type system

---

## 🤝 Contributing

While this specific project is suspended, you're welcome to:

1. **Feedback**: Did you need this tool?
2. **Ideas**: What would make PathSimulator valuable for you?
3. **Contributions to FormatBridge**: Join the new project!

---

## 📜 License

MIT License - See [LICENSE](./LICENSE) for details

This project uses [Tweego](https://www.motoslave.net/tweego/) (BSD 2-Clause License) as an external tool.

---

## 🙏 Acknowledgments

- Chris Klimas for [Twine](https://twinery.org/)
- Thomas Michael Edwards for [Tweego](https://www.motoslave.net/tweego/)
- Everyone who contributed ideas and suggestions

---

## 🔮 Future Possibilities

This project isn't dead—it's **on hold**. 

---

**Made with ❤️ for the Interactive Fiction community**

_Project suspended: January 2026_  