import type { Meta, StoryObj } from "@storybook/react";

import { Button } from "./button";

const meta = {
  title: "Primitives/Button",
  component: Button,
  args: {
    children: "Run smoke action",
    variant: "default",
    size: "default",
    disabled: false,
  },
  argTypes: {
    variant: {
      control: "select",
      options: ["default", "secondary", "outline", "ghost"],
    },
    size: {
      control: "select",
      options: ["default", "sm", "lg", "icon"],
    },
  },
  tags: ["autodocs"],
} satisfies Meta<typeof Button>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const VariantsAndSizes: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      {(["default", "secondary", "outline", "ghost"] as const).map((variant) => (
        <div className="flex flex-wrap items-center gap-2" key={variant}>
          <span className="w-24 text-body-md text-text-secondary">{variant}</span>
          {(["default", "sm", "lg"] as const).map((size) => (
            <Button key={`${variant}-${size}`} variant={variant} size={size}>
              {variant} / {size}
            </Button>
          ))}
        </div>
      ))}
    </div>
  ),
};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};

export const LongContent: Story = {
  args: {
    children: "Run a long-running genomics workflow with validation and downstream QC checks",
  },
};

export const LoadingNotApplicable: Story = {
  render: () => (
    <div className="space-y-2">
      <Button>Run smoke action</Button>
      <p className="text-body-md text-text-secondary">
        Button has no built-in loading prop; loading visuals are composed by higher-level domain components.
      </p>
    </div>
  ),
};

export const ErrorNotApplicable: Story = {
  render: () => (
    <div className="space-y-2">
      <Button variant="outline">Run smoke action</Button>
      <p className="text-body-md text-text-secondary">
        Button has no intrinsic error state; error semantics belong to surrounding form or status UI.
      </p>
    </div>
  ),
};
