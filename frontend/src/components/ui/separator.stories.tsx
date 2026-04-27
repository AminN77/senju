import type { Meta, StoryObj } from "@storybook/react";

import { Separator } from "./separator";

const meta = {
  title: "Primitives/Separator",
  component: Separator,
  tags: ["autodocs"],
} satisfies Meta<typeof Separator>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="w-full max-w-md space-y-4">
      <p className="text-body-md text-text-primary">Upload</p>
      <Separator />
      <p className="text-body-md text-text-primary">QC</p>
    </div>
  ),
};

export const Vertical: Story = {
  render: () => (
    <div className="flex h-16 items-center gap-3">
      <span className="text-body-md text-text-primary">Jobs</span>
      <Separator orientation="vertical" />
      <span className="text-body-md text-text-primary">Variants</span>
    </div>
  ),
};

export const LongContent: Story = {
  render: () => (
    <div className="w-full max-w-md space-y-4">
      <p className="text-body-md text-text-primary">
        FASTQ upload and validation state with extensive explanatory helper text for baseline
        overflow checks
      </p>
      <Separator />
      <p className="text-body-md text-text-primary">
        Downstream execution stage details and post-processing artifact pointers
      </p>
    </div>
  ),
};

export const DisabledNotApplicable: Story = {
  render: () => (
    <div className="w-full max-w-md space-y-4">
      <p className="text-body-md text-text-primary">Section A</p>
      <Separator />
      <p className="text-body-md text-text-primary">Section B</p>
      <p className="text-body-md text-text-secondary">
        Separator is non-interactive and does not implement a disabled state.
      </p>
    </div>
  ),
};

export const LoadingNotApplicable: Story = {
  render: () => (
    <div className="w-full max-w-md space-y-4">
      <p className="text-body-md text-text-primary">Section A</p>
      <Separator />
      <p className="text-body-md text-text-primary">Section B</p>
      <p className="text-body-md text-text-secondary">
        Separator does not own loading behavior; loading indicators belong to nearby content blocks.
      </p>
    </div>
  ),
};

export const EmptyNotApplicable: Story = {
  render: () => (
    <div className="w-full max-w-md space-y-4">
      <p className="text-body-md text-text-primary">Section A</p>
      <Separator />
      <p className="text-body-md text-text-primary">Section B</p>
      <p className="text-body-md text-text-secondary">
        Separator is purely structural, so an empty-content state is not applicable.
      </p>
    </div>
  ),
};

export const ErrorNotApplicable: Story = {
  render: () => (
    <div className="w-full max-w-md space-y-4">
      <p className="text-body-md text-text-primary">Section A</p>
      <Separator />
      <p className="text-body-md text-text-primary">Section B</p>
      <p className="text-body-md text-text-secondary">
        Separator has no intrinsic error state; error signaling belongs to related form or status
        components.
      </p>
    </div>
  ),
};
