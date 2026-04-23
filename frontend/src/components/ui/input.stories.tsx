import type { Meta, StoryObj } from "@storybook/react";

import { Input } from "./input";

const meta = {
  title: "Primitives/Input",
  component: Input,
  args: {
    placeholder: "name@senju.dev",
    disabled: false,
  },
  tags: ["autodocs"],
} satisfies Meta<typeof Input>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Disabled: Story = {
  args: {
    disabled: true,
    placeholder: "Disabled input",
  },
};

export const LongContent: Story = {
  args: {
    defaultValue:
      "This is a long value to validate overflow and typography behavior in the input primitive under baseline Storybook states.",
  },
};
