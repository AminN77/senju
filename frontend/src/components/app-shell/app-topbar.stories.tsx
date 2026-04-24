import type { Meta, StoryObj } from "@storybook/react";

import { AppTopBar } from "./app-topbar";

const meta = {
  title: "App Shell/AppTopBar",
  component: AppTopBar,
  args: {
    breadcrumb: ["Dashboard"],
    onOpenMobileSidebar: () => {},
    onOpenShortcuts: () => {},
    onSignOut: async () => {},
  },
  tags: ["autodocs"],
} satisfies Meta<typeof AppTopBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SingleBreadcrumb: Story = {};

export const NestedBreadcrumb: Story = {
  args: {
    breadcrumb: ["Projects", "Project Alpha", "Files"],
  },
};

export const DisabledNotApplicable: Story = {
  render: () => (
    <p className="text-body-md text-text-secondary">
      Topbar actions remain interactive by design; disabled states are context-specific to
      downstream actions.
    </p>
  ),
};

export const LongContent: Story = {
  args: {
    breadcrumb: [
      "Machine Learning",
      "Impact Models",
      "Genome-Wide Association Cohort",
      "Training Run 2026-04-24",
    ],
  },
};
