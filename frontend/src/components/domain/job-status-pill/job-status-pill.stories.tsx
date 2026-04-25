import type { Meta, StoryObj } from "@storybook/nextjs";

import { JobStatusPill } from "./job-status-pill";

const meta = {
  title: "Domain/JobStatusPill",
  component: JobStatusPill,
  tags: ["autodocs"],
} satisfies Meta<typeof JobStatusPill>;

export default meta;
type Story = StoryObj<typeof meta>;

export const AllStates: Story = {
  args: {
    status: "queued",
  },
  render: () => (
    <div className="flex flex-wrap gap-2 bg-surface-raised p-4">
      <JobStatusPill status="queued" />
      <JobStatusPill status="running" />
      <JobStatusPill status="succeeded" />
      <JobStatusPill status="failed" />
      <JobStatusPill status="canceled" />
      <JobStatusPill status="paused" />
      <JobStatusPill status="checkpointed" />
    </div>
  ),
};
