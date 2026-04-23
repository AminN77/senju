import type { Meta, StoryObj } from "@storybook/react";

import { Input } from "./input";

const meta = {
  title: "Primitives/Input",
  component: Input,
  args: {
    "aria-label": "Email address",
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
    "aria-label": "Long content input",
    defaultValue:
      "This is a long value to validate overflow and typography behavior in the input primitive under baseline Storybook states.",
  },
};

export const Loading: Story = {
  args: {
    "aria-label": "Loading input",
    disabled: true,
    defaultValue: "Loading profile data...",
  },
};

export const Error: Story = {
  args: {
    "aria-label": "Invalid email input",
    "aria-invalid": true,
    defaultValue: "invalid-email",
  },
};

export const Empty: Story = {
  args: {
    "aria-label": "Empty input",
    placeholder: "Enter a value",
    defaultValue: "",
  },
};
