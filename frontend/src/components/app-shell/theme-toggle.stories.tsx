import type { Meta, StoryObj } from "@storybook/react";

import { ThemeToggle } from "./theme-toggle";

const meta = {
  title: "App Shell/ThemeToggle",
  component: ThemeToggle,
  args: {
    compact: false,
  },
  tags: ["autodocs"],
} satisfies Meta<typeof ThemeToggle>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Compact: Story = {
  args: {
    compact: true,
  },
};

export const EmptyNotApplicable: Story = {
  render: () => (
    <p className="text-body-md text-text-secondary">
      Theme toggle is a fixed-action control and does not have an empty state.
    </p>
  ),
};

export const ErrorNotApplicable: Story = {
  render: () => (
    <p className="text-body-md text-text-secondary">
      Persistence failures fall back to localStorage transparently; no dedicated error state is
      exposed.
    </p>
  ),
};
