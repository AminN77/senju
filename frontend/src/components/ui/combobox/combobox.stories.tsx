import type { Meta, StoryObj } from "@storybook/react";

import { Combobox } from "./combobox";

const meta = {
  title: "Primitives/Combobox",
  component: Combobox,
  args: {
    id: "sample-combobox",
    options: [
      { label: "Sample A", value: "sample-a" },
      { label: "Sample B", value: "sample-b" },
    ],
    placeholder: "Search samples",
  },
  tags: ["autodocs"],
} satisfies Meta<typeof Combobox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
