import type { Meta, StoryObj } from "@storybook/react";

import { Textarea } from "./textarea";

const meta = {
  title: "Primitives/Textarea",
  component: Textarea,
  args: {
    "aria-label": "Notes",
    placeholder: "Add notes",
  },
  tags: ["autodocs"],
} satisfies Meta<typeof Textarea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Disabled: Story = { args: { disabled: true, defaultValue: "Read-only value" } };
