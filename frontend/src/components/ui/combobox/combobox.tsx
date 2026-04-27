"use client";

import * as React from "react";

import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export interface ComboboxOption {
  label: string;
  value: string;
}

/**
 * Lightweight token-driven combobox for searchable single selection.
 * Use `Select` for non-filterable option lists.
 * A11y: uses native datalist semantics and requires an external label.
 */
export function Combobox({
  id,
  options,
  value,
  onValueChange,
  placeholder,
  "aria-label": ariaLabel,
  disabled,
  className,
}: {
  id: string;
  options: ComboboxOption[];
  value?: string;
  onValueChange?: (value: string) => void;
  placeholder?: string;
  "aria-label"?: string;
  disabled?: boolean;
  className?: string;
}) {
  const listId = `${id}-list`;
  return (
    <div className={cn("w-full", className)}>
      <Input
        id={id}
        aria-label={ariaLabel ?? "Combobox"}
        value={value}
        onChange={(event) => onValueChange?.(event.target.value)}
        list={listId}
        placeholder={placeholder}
        disabled={disabled}
      />
      <datalist id={listId}>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </datalist>
    </div>
  );
}
