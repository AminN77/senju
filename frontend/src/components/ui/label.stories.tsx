import type { Meta, StoryObj } from "@storybook/react";

import { Input } from "./input";
import { Label } from "./label";

const meta = {
  title: "Primitives/Label",
  component: Label,
  args: {
    children: "Email",
  },
  tags: ["autodocs"],
} satisfies Meta<typeof Label>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-2">
      <Label htmlFor="label-default">Email</Label>
      <Input id="label-default" placeholder="name@senju.dev" />
    </div>
  ),
};

export const Disabled: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-2">
      <Label htmlFor="label-disabled">Email</Label>
      <Input id="label-disabled" placeholder="disabled@senju.dev" disabled />
    </div>
  ),
};

export const LongContent: Story = {
  args: {
    children:
      "Primary contact email for workflow notifications and pipeline failure alerts across all projects",
  },
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-2">
      <Label htmlFor="label-long">
        Primary contact email for workflow notifications and pipeline failure alerts across all projects
      </Label>
      <Input id="label-long" placeholder="name@senju.dev" />
    </div>
  ),
};
