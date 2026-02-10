import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'protoc-gen-nats-micro',
  description: 'Generate type-safe NATS microservices from Protocol Buffers',
  base: '/protoc-gen-nats-micro/',

  head: [
    ['meta', { name: 'theme-color', content: '#27ae60' }],
    ['meta', { name: 'og:type', content: 'website' }],
    ['meta', { name: 'og:title', content: 'protoc-gen-nats-micro' }],
    ['meta', { name: 'og:description', content: 'Generate type-safe NATS microservices from Protocol Buffers' }],
  ],

  themeConfig: {
    logo: undefined,

    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API Reference', link: '/api/reference' },
      { text: 'Examples', link: '/examples/go' },
      {
        text: 'GitHub',
        link: 'https://github.com/Toyz/protoc-gen-nats-micro'
      }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Quick Start', link: '/guide/quick-start' },
          ]
        },
        {
          text: 'Features',
          items: [
            { text: 'Streaming RPC', link: '/guide/streaming' },
            { text: 'KV & Object Store', link: '/guide/kv-object-store' },
            { text: 'Interceptors & Headers', link: '/guide/interceptors' },
            { text: 'Error Handling', link: '/guide/error-handling' },
          ]
        }
      ],
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Proto Options', link: '/api/reference' },
            { text: 'Service Options', link: '/api/service-options' },
            { text: 'Endpoint Options', link: '/api/endpoint-options' },
          ]
        }
      ],
      '/examples/': [
        {
          text: 'Examples',
          items: [
            { text: 'Go', link: '/examples/go' },
            { text: 'TypeScript', link: '/examples/typescript' },
            { text: 'Python', link: '/examples/python' },
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/Toyz/protoc-gen-nats-micro' }
    ],

    search: {
      provider: 'local'
    },

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present'
    },

    editLink: {
      pattern: 'https://github.com/Toyz/protoc-gen-nats-micro/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
    }
  }
})
