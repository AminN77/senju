"use client";

import * as React from "react";

import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

export interface MultiSelectOption {
  label: string;
  value: string;
}

/**
 * Grouped checkbox multi-select primitive for selecting multiple discrete values.
 * Use `Combobox` for searchable single selection.
 * A11y: renders semantic checkbox controls; consumers provide field labeling context.
 */
export function MultiSelect({
  options,
  value,
  onValueChange,
  className,
}: {
  options: MultiSelectOption[];
  value?: string[];
  onValueChange?: (value: string[]) => void;
  className?: string;
}) {
  const selected = new Set(value ?? []);
  return (
    <div className={cn("space-y-2", className)}>
      {options.map((option) => {
        const inputId = `multi-select-${option.value}`;
        return (
          <div key={option.value} className="flex items-center gap-2">
            <Checkbox
              id={inputId}
              checked={selected.has(option.value)}
              onCheckedChange={(checked) => {
                const next = new Set(selected);
                if (checked) {
                  next.add(option.value);
                } else {
                  next.delete(option.value);
                }
                onValueChange?.([...next]);
              }}
            />
            <Label htmlFor={inputId}>{option.label}</Label>
          </div>
        );
      })}
    </div>
  );
}
