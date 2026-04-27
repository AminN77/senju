import { Settings } from "lucide-react";
import type { Meta, StoryObj } from "@storybook/react";

import { IconButton } from "./icon-button";

const meta = {
  title: "Primitives/IconButton",
  component: IconButton,
  args: {
    "aria-label": "Open settings",
    children: <Settings className="h-4 w-4" />,
  },
  tags: ["autodocs"],
} satisfies Meta<typeof IconButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Disabled: Story = { args: { disabled: true } };
