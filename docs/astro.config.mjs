import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://your-org.github.io',
  base: '/gowright',
  integrations: [
    starlight({
      title: 'Gowright Testing Framework',
      social: {
        github: 'https://github.com/your-org/gowright',
      },
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            { label: 'Introduction', link: '/' },
            { label: 'Quick Start', link: '/getting-started/quick-start/' },
            { label: 'Installation', link: '/getting-started/installation/' },
          ],
        },
        {
          label: 'Configuration',
          items: [
            { label: 'Overview', link: '/configuration/' },
          ],
        },
        {
          label: 'Testing Modules',
          items: [
            { label: 'API Testing', link: '/testing/api/' },
            { label: 'UI Testing', link: '/testing/ui/' },
            { label: 'Database Testing', link: '/testing/database/' },
            { label: 'Integration Testing', link: '/testing/integration/' },
          ],
        },
        {
          label: 'Examples',
          items: [
            { label: 'Basic Examples', link: '/examples/' },
          ],
        },
        {
          label: 'API Reference',
          items: [
            { label: 'Core Framework', link: '/api/' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { label: 'Migration Guide', link: '/guides/migration/' },
            { label: 'Best Practices', link: '/guides/best-practices/' },
            { label: 'Troubleshooting', link: '/guides/troubleshooting/' },
          ],
        },
        {
          label: 'Community',
          items: [
            { label: 'Contributing', link: '/community/contributing/' },
            { label: 'Changelog', link: '/community/changelog/' },
            { label: 'Support', link: '/community/support/' },
          ],
        },
      ],
    }),
  ],
});