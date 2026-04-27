import type { Meta, StoryObj } from "@storybook/react";

import { NumberInput } from "./number-input";

const meta = {
  title: "Primitives/NumberInput",
  component: NumberInput,
  args: {
    "aria-label": "Read depth",
    min: 0,
    placeholder: "0",
  },
  tags: ["autodocs"],
} satisfies Meta<typeof NumberInput>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Disabled: Story = { args: { disabled: true, defaultValue: 42 } };
