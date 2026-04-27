import { test, expect } from '@grafana/plugin-e2e';

// This smoke test asserts that the plugin loads in all supported versions of Grafana via
// .github/workflows/ci.yml - resolve-versions.
test(
  'smoke: datasource config page loads',
  { tag: '@plugins' },
  async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'retrodaredevil-wildgraphql-datasource' });
    await expect(
      page.getByText('Wild GraphQL Data Source', { exact: true }),
    ).toBeVisible();
  },
);
