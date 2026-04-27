import type { Meta, StoryObj } from "@storybook/react";

import { LinkButton } from "./link-button";

const meta = {
  title: "Primitives/LinkButton",
  component: LinkButton,
  args: {
    href: "/",
    children: "Go home",
  },
  tags: ["autodocs"],
} satisfies Meta<typeof LinkButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
