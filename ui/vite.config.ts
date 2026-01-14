import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { extensionScriptsPlugin } from './vite-plugin-extension-scripts'

// ==============================================================
// Extension Script Injection Configuration
// ==============================================================
//
// To inject scripts from extensions at build time, import your
// extension manifests and pass them to extensionScriptsPlugin:
//
// import billingExtension from '@acme/zrok-billing-extension';
// import demoExtension from './examples/demo-extension/src';
//
// Then add to plugins array:
//   extensionScriptsPlugin({
//     extensions: [billingExtension, demoExtension],
//     verbose: true, // Enable to see injection logs during build
//   }),
//
// ==============================================================

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    // Uncomment and configure to enable build-time script injection:
    // extensionScriptsPlugin({
    //   extensions: [],
    //   verbose: true,
    // }),
  ],
  server: {
    proxy: {
      '/api/v2': {
        target: 'http://localhost:18080',
        changeOrigin: true,
      }
    },
    allowedHosts: [
      ".share.zrok.io"
    ]
  }
})
