import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules")) {
            if (id.includes("react-dom") || id.includes("react-router") || id.includes("@emotion") || id.match(/\/react\//)) {
              return "vendor-react";
            }
            if (id.includes("@mui/x-charts") || id.includes("@mui/x-charts-vendor")) {
              return "vendor-mui-charts";
            }
            if (id.includes("@mui/x-date-pickers")) {
              return "vendor-mui-pickers";
            }
            if (id.includes("@mui")) {
              return "vendor-mui";
            }
            if (id.includes("@xyflow") || id.includes("d3-hierarchy")) {
              return "vendor-visualizer";
            }
            if (id.includes("material-react-table")) {
              return "vendor-table";
            }
          }
        },
      },
    },
  },
  server: {
    proxy: {
      '/api/v2': {
        target: 'https://api-v2.zrok.io',
        changeOrigin: true,
      }
    },
    allowedHosts: [
      ".share.zrok.io"
    ]
  }
})
