import { test, expect } from '@grafana/plugin-e2e';

test(
  'smoke: datasource config page loads',
  { tag: '@plugins' },
  async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'retrodaredevil-wildgraphql-datasource' });
    await expect(
      page.getByText('Type: Wild GraphQL Data Source', { exact: true }),
    ).toBeVisible();
  },
);
