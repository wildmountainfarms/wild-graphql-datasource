import { test, expect } from '@grafana/plugin-e2e';

const DASHBOARD_UID = 'a3c397f1-e970-4242-b470-9fda26e91b5f';

// The SWAPI endpoint is a public external API. Tests that hit it may be slow
// or flaky if the service is temporarily unavailable.
test.describe('Star Wars Film Info dashboard', () => {
  test('dashboard loads and first panels are visible', async ({ gotoDashboardPage, page }) => {
    const dashboardPage = await gotoDashboardPage({ uid: DASHBOARD_UID });

    await expect(dashboardPage.getPanelByTitle('Film Information').locator).toBeVisible();
    await expect(dashboardPage.getPanelByTitle('A New Hope Opening Crawl').locator).toBeVisible();
    // Scroll to the bottom to trigger rendering of panels below the fold
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await expect(dashboardPage.getPanelByTitle('Starships').locator).toBeVisible({ timeout: 15000 });
  });

  test('Film Information table has correct columns and contains known films', async ({ gotoDashboardPage }) => {
    const dashboardPage = await gotoDashboardPage({ uid: DASHBOARD_UID });
    const panel = dashboardPage.getPanelByTitle('Film Information');

    await expect(panel.fieldNames).toContainText(['Title', 'Directory', 'Release']);
    // A New Hope and The Empire Strikes Back are in the SWAPI dataset
    await expect(panel.data).toContainText(['A New Hope', 'The Empire Strikes Back']);
  });

  test('A New Hope Opening Crawl displays the opening text', async ({ gotoDashboardPage }) => {
    const dashboardPage = await gotoDashboardPage({ uid: DASHBOARD_UID });
    const panel = dashboardPage.getPanelByTitle('A New Hope Opening Crawl');

    await expect(panel.data).toContainText(['It is a period of civil war']);
  });

  test('Database Activity panel renders as timeseries', async ({ gotoDashboardPage }) => {
    const dashboardPage = await gotoDashboardPage({ uid: DASHBOARD_UID });
    const panel = dashboardPage.getPanelByTitle('Database Activity December 2014');

    // Timeseries panels don't expose table data, so we assert the panel rendered without error
    await expect(panel.locator).toBeVisible();
    await expect(panel.locator.getByText('No data')).not.toBeVisible();
  });

  test('Film Information query executes and returns data in panel editor', async ({ gotoPanelEditPage }) => {
    const panelEditPage = await gotoPanelEditPage({
      dashboard: { uid: DASHBOARD_UID },
      id: '1',
    });

    await expect(panelEditPage.refreshPanel()).toBeOK();
    await expect(panelEditPage.panel.fieldNames).toContainText(['Title', 'Directory', 'Release']);
    await expect(panelEditPage.panel.data).toContainText(['A New Hope']);
  });

});
