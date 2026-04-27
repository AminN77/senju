import type { Meta, StoryObj } from "@storybook/react";

import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "./radio-group";

const meta = {
  title: "Primitives/RadioGroup",
  component: RadioGroup,
  tags: ["autodocs"],
} satisfies Meta<typeof RadioGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <RadioGroup defaultValue="fast">
      <div className="flex items-center gap-2">
        <RadioGroupItem id="speed-fast" value="fast" />
        <Label htmlFor="speed-fast">Fast mode</Label>
      </div>
      <div className="flex items-center gap-2">
        <RadioGroupItem id="speed-safe" value="safe" />
        <Label htmlFor="speed-safe">Safe mode</Label>
      </div>
    </RadioGroup>
  ),
};
