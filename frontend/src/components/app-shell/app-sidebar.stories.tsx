import type { Meta, StoryObj } from "@storybook/react";

import { AppSidebar } from "./app-sidebar";

const meta = {
  title: "App Shell/AppSidebar",
  component: AppSidebar,
  args: {
    collapsed: false,
    mobileOpen: true,
    onToggleCollapsed: () => {},
    onCloseMobile: () => {},
  },
  tags: ["autodocs"],
} satisfies Meta<typeof AppSidebar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Expanded: Story = {};

export const Collapsed: Story = {
  args: {
    collapsed: true,
  },
};

export const LoadingNotApplicable: Story = {
  render: () => (
    <p className="text-body-md text-text-secondary">
      Sidebar state is structural and has no loading variant at the primitive shell level.
    </p>
  ),
};

export const ErrorNotApplicable: Story = {
  render: () => (
    <p className="text-body-md text-text-secondary">
      Sidebar error handling is delegated to route content and data-fetching layers.
    </p>
  ),
};
