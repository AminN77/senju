import type { Meta, StoryObj } from "@storybook/react";

import { Switch } from "./switch";

const meta = {
  title: "Primitives/Switch",
  component: Switch,
  args: {
    "aria-label": "Enable notifications",
  },
  tags: ["autodocs"],
} satisfies Meta<typeof Switch>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Checked: Story = { args: { checked: true } };
