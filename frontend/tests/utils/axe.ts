import { configureAxe } from "jest-axe";

export const axe = configureAxe({
  rules: {
    // JSDOM does not model landmark regions like a full browser.
    region: { enabled: false },
  },
});
