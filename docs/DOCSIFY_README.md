# Docsify Documentation Setup

This documentation is built with [Docsify](https://docsify.js.org/), a magical documentation site generator that creates beautiful documentation websites from Markdown files without the need for a build process.

## 🚀 Quick Start

### Option 1: Using the provided scripts

**Linux/macOS:**
```bash
cd docs
./serve.sh
```

**Windows:**
```cmd
cd docs
serve.bat
```

### Option 2: Manual setup

**With docsify-cli (recommended):**
```bash
# Install docsify-cli globally
npm install -g docsify-cli

# Serve the documentation
cd docs
docsify serve . --port 3000
```

**With Python:**
```bash
cd docs
python -m http.server 3000
```

**With Node.js (no installation required):**
```bash
cd docs
npx docsify-cli serve . --port 3000
```

Then open [http://localhost:3000](http://localhost:3000) in your browser.

## 📁 File Structure

```
docs/
├── index.html          # Docsify configuration and entry point
├── README.md           # Homepage content
├── _sidebar.md         # Sidebar navigation
├── _navbar.md          # Top navigation bar
├── _coverpage.md       # Cover page content
├── .nojekyll          # Tells GitHub Pages to skip Jekyll processing
├── serve.sh           # Linux/macOS serve script
├── serve.bat          # Windows serve script
└── [content folders]  # Documentation content organized by topic
```

## 🎨 Customization

### Themes
The documentation uses the Vue theme. You can change it by modifying the CSS link in `index.html`:

```html
<link rel="stylesheet" href="//cdn.jsdelivr.net/npm/docsify@4/lib/themes/vue.css">
```

Available themes:
- `vue.css` (default)
- `buble.css`
- `dark.css`
- `pure.css`

### Plugins
Current plugins enabled:
- **Search** - Full-text search functionality
- **Copy Code** - Copy code blocks with one click
- **Pagination** - Previous/Next navigation
- **Tabs** - Tabbed content support
- **Zoom Image** - Click to zoom images
- **Syntax Highlighting** - Code syntax highlighting for multiple languages

### Configuration
Main configuration is in `docs/index.html` within the `window.$docsify` object. Key settings:

- `name` - Site name
- `repo` - GitHub repository link
- `loadSidebar` - Enable sidebar navigation
- `loadNavbar` - Enable top navigation
- `coverpage` - Enable cover page
- `search` - Search configuration
- `auto2top` - Auto scroll to top when route changes

## 🔧 Adding Content

### New Pages
1. Create a new `.md` file in the appropriate directory
2. Add the page to `_sidebar.md` for navigation
3. Use standard Markdown syntax

### Navigation
- Edit `_sidebar.md` to modify the sidebar navigation
- Edit `_navbar.md` to modify the top navigation bar
- Edit `_coverpage.md` to modify the cover page

### Images and Assets
Place images in the same directory as your markdown files or create an `assets` folder. Reference them using relative paths:

```markdown
![Description](./images/screenshot.png)
```

## 🚀 Deployment

### GitHub Pages
The documentation is automatically deployed to GitHub Pages via GitHub Actions when changes are pushed to the main branch. The workflow:

1. Validates the Docsify setup
2. Verifies documentation structure
3. Deploys the `docs` directory as static files

### Manual Deployment
You can also deploy manually to any static hosting service by uploading the entire `docs` directory.

## 📚 Docsify Features

### Markdown Extensions
Docsify supports enhanced Markdown features:

**Alerts:**
```markdown
> [!NOTE]
> This is a note

> [!WARNING]
> This is a warning

> [!TIP]
> This is a tip
```

**Tabs:**
```markdown
<!-- tabs:start -->

#### ** English **
Hello!

#### ** French **
Bonjour!

<!-- tabs:end -->
```

**Code Highlighting:**
```go
func main() {
    fmt.Println("Hello, World!")
}
```

### Search
The search functionality indexes all content automatically. Users can search using the search box in the sidebar.

### Responsive Design
The documentation is fully responsive and works well on desktop, tablet, and mobile devices.

## 🆘 Troubleshooting

### Common Issues

**Port already in use:**
```bash
# Use a different port
docsify serve . --port 3001
```

**CORS issues with file:// protocol:**
Always use a local server (not opening index.html directly in browser).

**Images not loading:**
Check that image paths are relative and correct. Docsify serves from the docs directory root.

**Sidebar not showing:**
Ensure `_sidebar.md` exists and `loadSidebar: true` is set in the configuration.

### Getting Help
- [Docsify Documentation](https://docsify.js.org/)
- [Docsify GitHub Repository](https://github.com/docsifyjs/docsify)
- [Project Issues](https://github.com/gowright/framework/issues)

---

Happy documenting! 📖✨