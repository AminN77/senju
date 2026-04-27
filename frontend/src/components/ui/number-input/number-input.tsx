import * as React from "react";

import { Input } from "@/components/ui/input";

/**
 * Numeric input primitive with tokenized styles and browser-native semantics.
 * Use `Input` for non-numeric content.
 * A11y: consumer must provide a visible label or an `aria-label`.
 */
export const NumberInput = React.forwardRef<HTMLInputElement, React.ComponentProps<typeof Input>>(
  ({ inputMode = "decimal", ...props }, ref) => {
    return <Input ref={ref} type="number" inputMode={inputMode} {...props} />;
  }
);

NumberInput.displayName = "NumberInput";
