import { defineConfig } from 'orval';

export default defineConfig({
  microchat: {
    input: {
      target: '../doc/openapi.json',
    },
    output: {
      mode: 'tags-split',
      target: './src/lib/api/generated',
      client: 'react-query',
      baseUrl: '',
      httpClient: 'fetch',
    },
  },
});
