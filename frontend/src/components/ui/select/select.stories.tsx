import type { Meta, StoryObj } from "@storybook/react";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";

const meta = {
  title: "Primitives/Select",
  component: Select,
  tags: ["autodocs"],
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Select defaultValue="wgs">
      <SelectTrigger aria-label="Pipeline type">
        <SelectValue placeholder="Select pipeline" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="wgs">Whole Genome</SelectItem>
        <SelectItem value="wes">Whole Exome</SelectItem>
      </SelectContent>
    </Select>
  ),
};
