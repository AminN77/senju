import { AxeBuilder } from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

const themes = ["light", "dark"] as const;

test("hello-world route has no wcag21aa violations in both themes", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1, name: /smoke route/i })).toBeVisible();

  for (const theme of themes) {
    await page.evaluate((nextTheme) => {
      document.documentElement.setAttribute("data-theme", nextTheme);
    }, theme);

    const accessibilityScanResults = await new AxeBuilder({ page }).withTags(["wcag21aa"]).analyze();
    expect(accessibilityScanResults.violations).toEqual([]);
  }
});
