import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";

import { axe } from "../../../tests/utils/axe";
import { render } from "../../../tests/utils/render";
import { Checkbox } from "./checkbox";
import { FormError, FormField, FormHint } from "./form-field";
import { IconButton } from "./icon-button";
import { Input } from "./input";
import { Label } from "./label";
import { MultiSelect } from "./multi-select";
import { NumberInput } from "./number-input";
import { RadioGroup, RadioGroupItem } from "./radio-group";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";
import { Switch } from "./switch";
import { Textarea } from "./textarea";

describe("core primitives batch 1", () => {
  it("passes a11y checks for representative controls", async () => {
    const { container } = render(
      <div className="space-y-3">
        <Label htmlFor="notes">Notes</Label>
        <Textarea id="notes" />
        <NumberInput aria-label="Read depth" />
        <Checkbox aria-label="Enable processing" />
        <Switch aria-label="Enable notifications" />
        <IconButton aria-label="Open settings">S</IconButton>
        <FormField>
          <FormHint>Hint</FormHint>
          <FormError>Error</FormError>
        </FormField>
      </div>
    );
    expect(await axe(container)).toHaveNoViolations();
  });

  it("renders radio group options with accessible roles", () => {
    render(
      <RadioGroup defaultValue="fast">
        <RadioGroupItem id="fast" value="fast" />
        <Label htmlFor="fast">Fast</Label>
        <RadioGroupItem id="safe" value="safe" />
        <Label htmlFor="safe">Safe</Label>
      </RadioGroup>
    );

    expect(screen.getByRole("radio", { name: "Fast" })).toBeChecked();
    expect(screen.getByRole("radio", { name: "Safe" })).not.toBeChecked();
  });

  it("renders select trigger with current value", () => {
    render(
      <Select defaultValue="wgs">
        <SelectTrigger aria-label="Pipeline type">
          <SelectValue placeholder="Select pipeline" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="wgs">Whole Genome</SelectItem>
          <SelectItem value="wes">Whole Exome</SelectItem>
        </SelectContent>
      </Select>
    );

    expect(screen.getByRole("combobox", { name: "Pipeline type" })).toHaveTextContent("Whole Genome");
  });

  it("supports multi-select value changes", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();
    render(
      <MultiSelect
        options={[
          { label: "Germline", value: "germline" },
          { label: "Somatic", value: "somatic" },
        ]}
        value={["germline"]}
        onValueChange={onValueChange}
      />
    );

    await user.click(screen.getByRole("checkbox", { name: "Somatic" }));
    expect(onValueChange).toHaveBeenCalled();
  });

  it("works with react-hook-form registration flow", async () => {
    type FormValues = {
      email: string;
      depth: number;
      notes: string;
    };

    function FormHarness() {
      const form = useForm<FormValues>({
        defaultValues: { email: "", depth: 0, notes: "" },
      });
      return (
        <form onSubmit={form.handleSubmit(() => {})}>
          <Input aria-label="Email" {...form.register("email")} />
          <NumberInput aria-label="Depth" {...form.register("depth", { valueAsNumber: true })} />
          <Textarea aria-label="Notes" {...form.register("notes")} />
          <button type="submit">Submit</button>
        </form>
      );
    }

    const user = userEvent.setup();
    render(<FormHarness />);
    await user.type(screen.getByLabelText("Email"), "test@senju.dev");
    await user.clear(screen.getByLabelText("Depth"));
    await user.type(screen.getByLabelText("Depth"), "10");
    await user.type(screen.getByLabelText("Notes"), "valid notes");
    await user.click(screen.getByRole("button", { name: "Submit" }));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
