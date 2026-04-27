import type { Meta, StoryObj } from "@storybook/react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FormError, FormField, FormHint } from "./form-field";

const meta = {
  title: "Primitives/FormField",
  component: FormField,
  tags: ["autodocs"],
} satisfies Meta<typeof FormField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <FormField>
      <Label htmlFor="email">Email</Label>
      <Input id="email" placeholder="name@senju.dev" />
      <FormHint>We use this for run notifications.</FormHint>
    </FormField>
  ),
};

export const ErrorState: Story = {
  render: () => (
    <FormField>
      <Label htmlFor="invalid-email">Email</Label>
      <Input id="invalid-email" aria-invalid defaultValue="invalid" />
      <FormError>Enter a valid email address.</FormError>
    </FormField>
  ),
};
