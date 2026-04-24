import type { Meta, StoryObj } from "@storybook/react";

import { ShortcutDialog } from "./shortcut-dialog";

const meta = {
  title: "App Shell/ShortcutDialog",
  component: ShortcutDialog,
  args: {
    open: true,
    onClose: () => {},
  },
  tags: ["autodocs"],
} satisfies Meta<typeof ShortcutDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Closed: Story = {
  args: {
    open: false,
  },
};

export const DisabledNotApplicable: Story = {
  render: () => (
    <p className="text-body-md text-text-secondary">
      Shortcut dialog does not expose disabled state; it is either open or closed.
    </p>
  ),
};

export const LongContent: Story = {
  render: () => <ShortcutDialog open onClose={() => {}} />,
};
