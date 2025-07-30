import react from '@vitejs/plugin-react';
import { defineConfig, transformWithEsbuild } from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    {
      name: 'treat-js-files-as-jsx',
      async transform(code, id) {
        if (!/src\/.*\.js$/.test(id)) {
          return null;
        }
        return transformWithEsbuild(code, id, {
          loader: 'jsx',
          jsx: 'automatic',
        });
      },
    },
    react()
  ],
  optimizeDeps: {
    esbuildOptions: {
      loader: {
        '.js': 'jsx',
      },
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3002',
        changeOrigin: true,
      },
      '/mj': {
        target: 'http://localhost:3002',
        changeOrigin: true,
      },
      '/pg': {
        target: 'http://localhost:3002',
        changeOrigin: true,
      },
    },
  },
});
