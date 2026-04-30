import { test, expect } from '@grafana/plugin-e2e';

// The graphql-echo service is a local Docker container provisioned by docker-compose.yaml.
// It provides deterministic, predictable GraphQL responses — no external network dependency.
const DASHBOARD_UID = 'df8c5904-af34-4555-96ea-d31359396b10';

// Panel editor tests conflict when multiple Grafana editor sessions run in parallel
test.describe.configure({ mode: 'serial' });

test.describe('GraphQL Echo dashboard', () => {
  // --- Table: Expected Headers (panels 1) ---
  // The panel has a Grafana "Reduce" + "Organize" transform that pivots the aliased
  // fields into rows, renaming "Field" -> "Header" and "First *" -> "Value".

  test('Expected Headers table shows Header/Value columns with the proxied host', async ({ gotoDashboardPage }) => {
    const dashboardPage = await gotoDashboardPage({ uid: DASHBOARD_UID });
    const panel = dashboardPage.getPanelByTitle('Expected Headers');

    await expect(panel.fieldNames).toContainText(['Header', 'Value']);
    // Grafana proxies backend requests to graphql-echo:8080, so the host field must appear
    await expect(panel.data).toContainText(['host', 'graphql-echo:8080']);
  });

  // --- Panel editor round-trips ---

  test('Expected Headers query executes in panel editor with correct transformed output', async ({
    gotoPanelEditPage,
  }) => {
    const panelEditPage = await gotoPanelEditPage({
      dashboard: { uid: DASHBOARD_UID },
      id: '1',
    });

    await expect(panelEditPage.refreshPanel()).toBeOK();
    await expect(panelEditPage.panel.fieldNames).toContainText(['Header', 'Value']);
    await expect(panelEditPage.panel.data).toContainText(['host', 'graphql-echo:8080']);
  });

  // --- Table: Header names (panel 17) — tests explodeArrayPaths ---
  // Uses explodeArrayPaths: ["headerNames"] so each HTTP header name becomes its own row.

  test('Header names query uses explodeArrayPaths to produce one row per header', async ({
    gotoPanelEditPage,
  }) => {
    const panelEditPage = await gotoPanelEditPage({
      dashboard: { uid: DASHBOARD_UID },
      id: '17',
    });

    await expect(panelEditPage.refreshPanel()).toBeOK();
    await expect(panelEditPage.panel.fieldNames).toContainText(['headerNames']);
    await expect(panelEditPage.panel.data).toContainText(['host', 'user-agent']);
  });

  // --- Table: All headers (panel 2) — tests dataPath with nested object ---

  test('All headers query uses dataPath to return name and values columns per header', async ({
    gotoPanelEditPage,
  }) => {
    const panelEditPage = await gotoPanelEditPage({
      dashboard: { uid: DASHBOARD_UID },
      id: '2',
    });

    await expect(panelEditPage.refreshPanel()).toBeOK();
    await expect(panelEditPage.panel.fieldNames).toContainText(['name', 'values']);
    await expect(panelEditPage.panel.data).toContainText(['host', 'user-agent']);
  });

  // --- Query editor regression: crash fix + state isolation ---
  // Before the fix, EditorContextProvider called useStorage() which returned null
  // when StorageContextProvider was absent, causing a crash on any panel editor open.
  // Also verifies state isolation: without no-op storage, navigating from panel 1 to
  // panel 17 would restore panel 1's query from localStorage into panel 17's editor.

  test('query editor renders and shows its own query when navigating between panels', async ({
    gotoPanelEditPage,
    page,
  }) => {
    // Open panel 1 first — seeds any localStorage state from the expected-headers query
    await gotoPanelEditPage({ dashboard: { uid: DASHBOARD_UID }, id: '1' });

    // Navigate to panel 17 (Header names)
    await gotoPanelEditPage({ dashboard: { uid: DASHBOARD_UID }, id: '17' });

    const queryEditor = page.locator('.graphiql-query-editor');
    await expect(queryEditor).toBeVisible();
    // Panel 17 uses "headerNames"; panel 1 uses "expectHeader" — must not bleed across
    await expect(queryEditor).toContainText('headerNames');
    await expect(queryEditor).not.toContainText('expectHeader');
  });

  // --- Timeseries: Generated Processor Temperatures (panel 3) ---

  test('Generated Processor Temperatures timeseries renders data without errors', async ({
    gotoPanelEditPage,
  }) => {
    const panelEditPage = await gotoPanelEditPage({
      dashboard: { uid: DASHBOARD_UID },
      id: '3',
    });

    await expect(panelEditPage.refreshPanel()).toBeOK();
    await expect(panelEditPage.panel.locator).toBeVisible();
    await expect(panelEditPage.panel.locator.getByText('No data')).not.toBeVisible();
  });
});
