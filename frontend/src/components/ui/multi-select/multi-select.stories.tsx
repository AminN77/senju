import type { Meta, StoryObj } from "@storybook/react";

import { MultiSelect } from "./multi-select";

const meta = {
  title: "Primitives/MultiSelect",
  component: MultiSelect,
  args: {
    options: [
      { label: "Germline", value: "germline" },
      { label: "Somatic", value: "somatic" },
    ],
    value: ["germline"],
  },
  tags: ["autodocs"],
} satisfies Meta<typeof MultiSelect>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
