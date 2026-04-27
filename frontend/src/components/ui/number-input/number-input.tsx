import * as React from "react";

import { Input } from "@/components/ui/input";

/**
 * Numeric input primitive with tokenized styles and browser-native semantics.
 * Use `Input` for non-numeric content.
 * A11y: consumer must provide a visible label or an `aria-label`.
 */
export const NumberInput = React.forwardRef<
  HTMLInputElement,
  Omit<React.ComponentProps<typeof Input>, "type">
>(({ inputMode = "decimal", ...props }, ref) => {
  return <Input ref={ref} {...props} type="number" inputMode={inputMode} />;
});

NumberInput.displayName = "NumberInput";
