import { expect, test } from "@playwright/test";

const themes = ["light", "dark"] as const;

test("hello-world route smoke test renders in both themes", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1, name: /smoke route/i })).toBeVisible();

  for (const theme of themes) {
    await page.evaluate((nextTheme) => {
      document.documentElement.setAttribute("data-theme", nextTheme);
    }, theme);
    await expect(
      page.getByRole("button", { name: new RegExp(`Preview ${theme}`, "i") })
    ).toBeVisible();
  }
});
