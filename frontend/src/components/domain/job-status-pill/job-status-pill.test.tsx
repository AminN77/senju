import { describe, expect, it } from "vitest";

import { axe } from "../../../../tests/utils/axe";
import { render } from "../../../../tests/utils/render";
import { JobStatusPill, type JobStatus } from "./job-status-pill";

describe("JobStatusPill", () => {
  it("renders each supported status label", () => {
    const statuses: JobStatus[] = [
      "queued",
      "running",
      "succeeded",
      "failed",
      "canceled",
      "paused",
      "checkpointed",
    ];

    const { container } = render(
      <div className="flex flex-wrap gap-2">
        {statuses.map((status) => (
          <JobStatusPill key={status} status={status} />
        ))}
      </div>
    );

    for (const status of statuses) {
      expect(container.textContent).toContain(status.charAt(0).toUpperCase() + status.slice(1));
    }
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<JobStatusPill status="running" />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
