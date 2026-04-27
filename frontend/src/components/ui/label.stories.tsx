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
        Primary contact email for workflow notifications and pipeline failure alerts across all
        projects
      </Label>
      <Input id="label-long" placeholder="name@senju.dev" />
    </div>
  ),
};

export const VariantAndSizeNotApplicable: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-2">
      <Label htmlFor="label-na">Email</Label>
      <Input id="label-na" placeholder="name@senju.dev" />
      <p className="text-body-md text-text-secondary">
        Label does not expose variant or size props; visual variants are owned by composed/domain
        form components.
      </p>
    </div>
  ),
};

export const LoadingNotApplicable: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-2">
      <Label htmlFor="label-loading">Email</Label>
      <Input id="label-loading" placeholder="Loading state lives in parent form control" disabled />
      <p className="text-body-md text-text-secondary">
        Label itself is static text and does not implement a loading state.
      </p>
    </div>
  ),
};

export const EmptyNotApplicable: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-2">
      <Label htmlFor="label-empty">Email</Label>
      <Input id="label-empty" placeholder="name@senju.dev" />
      <p className="text-body-md text-text-secondary">
        Empty label text is intentionally not a supported state for accessibility reasons.
      </p>
    </div>
  ),
};

export const ErrorNotApplicable: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-2">
      <Label htmlFor="label-error">Email</Label>
      <Input id="label-error" aria-invalid defaultValue="invalid-email" />
      <p className="text-body-md text-text-secondary">
        Error states are applied to the associated input or helper text, not to Label itself.
      </p>
    </div>
  ),
};
