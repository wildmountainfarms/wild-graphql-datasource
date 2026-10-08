import { monacoLanguageRegistry } from '@grafana/data';

// Grafana owns the shared MonacoEnvironment worker factory. GraphiQL's standalone
// setup replaces it, and Grafana can replace it again during initialization.
// Register GraphQL with Grafana instead: otherwise its factory falls back to a
// generic editor worker, whose loadForeignModule rejects with "Unexpected usage".
// Keep Grafana's existing JSON and other language workers intact.
if (!monacoLanguageRegistry.getIfExists('graphql')) {
  monacoLanguageRegistry.register({
    id: 'graphql',
    name: 'GraphQL',
    init: () => {
      const worker = new Worker(new URL('monaco-graphql/esm/graphql.worker.js', import.meta.url));
      // Monaco's fallback hides the original startup error behind "Unexpected
      // usage". Keep the worker URL and original error available for diagnosis.
      worker.addEventListener('error', (event) => {
        console.error('GraphQL worker failed to start', event.message, event.filename, event.lineno);
      });
      return worker;
    },
  });
}
