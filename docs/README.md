# Gowright Documentation

This directory contains the complete documentation for the Gowright testing framework, built with **Astro Starlight** and hosted on GitHub Pages.

## Local Development

To run the documentation site locally:

1. **Install Node.js** (version 18 or later):
   ```bash
   # Download from https://nodejs.org/
   # Or using package managers:
   
   # On macOS with Homebrew
   brew install node
   
   # On Ubuntu/Debian
   curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
   sudo apt-get install -y nodejs
   ```

2. **Install dependencies**:
   ```bash
   cd docs
   npm install
   ```

3. **Run the development server**:
   ```bash
   npm run dev
   ```

4. **View the site** at `http://localhost:4321`

## Documentation Structure

- `src/content/docs/` - Documentation content in Markdown
  - `index.mdx` - Homepage with hero section
  - `getting-started/` - Installation and quick start guides
  - `configuration/` - Configuration documentation
  - `testing/` - Testing module guides
  - `examples/` - Practical examples
  - `api/` - API reference documentation
  - `guides/` - Migration and best practices
  - `community/` - Contributing and changelog
- `src/components/` - Custom Astro components
- `src/styles/` - Custom CSS styling
- `src/assets/` - Images, logos, and other assets
- `astro.config.mjs` - Astro and Starlight configuration
- `tsconfig.json` - TypeScript configuration

## GitHub Pages Deployment

The documentation is automatically deployed to GitHub Pages when changes are pushed to the main branch. The deployment is handled by the GitHub Actions workflow in `.github/workflows/pages.yml`.

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run astro` - Run Astro CLI commands

## Contributing to Documentation

1. Make changes to the markdown files in `src/content/docs/`
2. Test locally using `npm run dev`
3. Submit a pull request with your changes
4. Once merged, changes will be automatically deployed

## Technology Stack

- **Framework**: Astro 4 with Starlight theme
- **Content**: Markdown and MDX with frontmatter
- **Styling**: Built-in Starlight theme with custom CSS
- **Search**: Built-in search functionality
- **Deployment**: GitHub Pages with static site generation
- **Build Tool**: Astro's optimized build system

## Starlight Features

- **🔍 Full-text search** - Built-in search with no configuration
- **🌐 Internationalization** - Ready for multiple languages
- **📱 Mobile-friendly** - Responsive design out of the box
- **♿ Accessible** - WCAG 2.1 AA compliant
- **🎨 Customizable** - Easy theming and component overrides
- **⚡ Fast** - Optimized for performance with minimal JavaScript

## Customization

- **Theme**: Custom CSS in `src/styles/custom.css`
- **Navigation**: Configured in `astro.config.mjs` sidebar
- **Components**: Custom components in `src/components/`
- **Assets**: Logos and images in `src/assets/`
- **Colors**: CSS custom properties for brand colors

## Content Organization

The documentation follows Starlight's content structure:

```
src/content/docs/
├── index.mdx                 # Homepage
├── getting-started/
│   ├── quick-start.md
│   └── installation.md
├── configuration/
│   ├── index.md
│   ├── browser.md
│   ├── api.md
│   └── database.md
├── testing/
│   ├── api.md
│   ├── ui.md
│   ├── database.md
│   └── integration.md
├── examples/
│   ├── index.md
│   ├── api.md
│   ├── ui.md
│   └── database.md
├── api/
│   ├── index.md
│   ├── configuration.md
│   └── testing.md
├── guides/
│   ├── migration.md
│   ├── best-practices.md
│   └── troubleshooting.md
└── community/
    ├── contributing.md
    ├── changelog.md
    └── support.md
```

## Next Steps

After setting up the documentation:

1. **Customize the theme** by editing `src/styles/custom.css`
2. **Add your content** to the appropriate sections
3. **Update the navigation** in `astro.config.mjs`
4. **Add images and assets** to `src/assets/`
5. **Test the build** with `npm run build`