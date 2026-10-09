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

  test('Header names query uses explodeArrayPaths to produce one row per header', async ({ gotoPanelEditPage }) => {
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

  test('expanded query editor preserves undo history and inline size', async ({ gotoPanelEditPage, page }) => {
    await gotoPanelEditPage({ dashboard: { uid: DASHBOARD_UID }, id: '17' });

    const editor = page.locator('.graphiql-query-editor');
    const input = editor.locator('textarea');
    const container = page.locator('.wild-graphql-editor-container');
    const expand = page.getByRole('button', { name: 'Expand editor', exact: true });
    const close = page.getByRole('button', { name: 'Close expanded editor', exact: true });
    const dialog = page.getByRole('dialog', { name: 'Query editor', exact: true });

    await input.focus();
    await page.keyboard.press('ControlOrMeta+End');
    await page.keyboard.insertText('\n# inlineUndoMarker');
    await expect(editor).toContainText('inlineUndoMarker');
    await expand.click();
    await expect(dialog).toBeVisible();
    // GraphiQL's own dialogs render through portals outside the expanded editor.
    await page.getByRole('button', { name: 'Open settings dialog', exact: true }).click();
    const settings = page.getByRole('dialog', { name: 'Settings', exact: true });
    await expect(settings).toBeVisible();
    await settings.getByRole('button', { name: 'Close dialog' }).click();
    await expect(settings).not.toBeVisible();
    await expect(dialog).toBeVisible();
    await input.focus();
    await page.keyboard.press('ControlOrMeta+z');
    await expect(editor).not.toContainText('inlineUndoMarker');

    await page.keyboard.press('ControlOrMeta+End');
    await page.keyboard.insertText('\n# expandedUndoMarker');
    await expect(editor).toContainText('expandedUndoMarker');
    await close.click();
    await expect(expand).toBeFocused();
    await expect(editor).toContainText('expandedUndoMarker');
    await input.focus();
    await page.keyboard.press('ControlOrMeta+z');
    await expect(editor).not.toContainText('expandedUndoMarker');

    // Native resizing writes an inline height. It should survive a modal round trip.
    await container.evaluate((element) => { element.style.height = '600px'; });
    await expect(container).toHaveCSS('height', '600px');
    await expand.click();
    // Wrap focus backwards into the editor before dismissing it with the keyboard.
    await page.keyboard.press('Shift+Tab');
    await page.keyboard.press('Escape');
    await expect(expand).toBeFocused();
    await expect(container).toHaveCSS('height', '600px');
    await container.evaluate((element) => { element.style.height = '200px'; });
    await expect(container).toHaveCSS('height', '450px');
  });

  test('Escape closes GraphiQL popups without leaving the panel editor', async ({ gotoPanelEditPage, page }) => {
    await gotoPanelEditPage({ dashboard: { uid: DASHBOARD_UID }, id: '17' });
    const editor = page.locator('.graphiql-query-editor');
    const expand = page.getByRole('button', { name: 'Expand editor', exact: true });
    const expandedEditor = page.getByRole('dialog', { name: 'Query editor', exact: true });

    for (const expanded of [false, true]) {
      for (const name of ['Open settings dialog', 'Open short keys dialog']) {
        if (expanded) {
          await expand.click();
        }
        await page.getByRole('button', { name, exact: true }).click();
        const popup = page.locator('.graphiql-dialog');
        await expect(popup).toBeVisible();
        await page.keyboard.press('Escape');
        await expect(popup).not.toBeVisible();
        await expect(editor).toBeVisible();
        if (expanded) {
          await expect(expandedEditor).toBeVisible();
          // Do not manually restore focus: GraphiQL may leave it on the body.
          await page.keyboard.press('Escape');
          await expect(expandedEditor).not.toBeVisible();
          await expect(expand).toBeFocused();
          await expect(editor).toBeVisible();
          await expect(page).toHaveURL(/editPanel=17/);
        }
      }
    }

    // Grafana's normal Escape shortcut still works outside the popups.
    await page.keyboard.press('Escape');
    await expect(editor).not.toBeVisible();
  });

  // --- Timeseries: Generated Processor Temperatures (panel 3) ---

  test('Generated Processor Temperatures timeseries renders data without errors', async ({ gotoPanelEditPage }) => {
    test.slow(process.env.GRAFANA_VERSION === 'dev-preview-react19', 'testing if dev image is slower than stable releases.');
    const panelEditPage = await gotoPanelEditPage({
      dashboard: { uid: DASHBOARD_UID },
      id: '3',
    });

    await expect(panelEditPage.refreshPanel()).toBeOK();
    await expect(panelEditPage.panel.locator).toBeVisible();
    await expect(panelEditPage.panel.locator.getByText('No data')).not.toBeVisible();
  });
});
