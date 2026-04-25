import { describe, expect, it } from "vitest";

import { axe } from "../../../tests/utils/axe";
import { render } from "../../../tests/utils/render";
import TokensDemoPage from "./page";

describe("TokensDemoPage", () => {
  it("has no wcag violations", async () => {
    const { container } = render(<TokensDemoPage />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
