"use client";

import { Button } from "@/components/ui/button";

const shortcuts = [
  { keys: "/", action: "Focus global search" },
  { keys: "g d", action: "Go to Dashboard" },
  { keys: "g p", action: "Go to Projects" },
  { keys: "g j", action: "Go to Jobs" },
  { keys: "g v", action: "Go to Variants" },
  { keys: "?", action: "Open shortcut cheatsheet" },
  { keys: "Esc", action: "Close dialogs and menus" },
] as const;

interface ShortcutDialogProps {
  open: boolean;
  onClose: () => void;
}

export function ShortcutDialog({ open, onClose }: ShortcutDialogProps) {
  if (!open) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-surface-sunken/70 p-4"
      role="presentation"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="shortcut-dialog-title"
        className="w-full max-w-xl rounded-lg border border-border-default bg-surface-raised p-6 shadow-2"
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 id="shortcut-dialog-title" className="text-body-md font-semibold text-text-primary">
            Keyboard shortcuts
          </h2>
          <Button type="button" size="sm" variant="outline" onClick={onClose}>
            Close
          </Button>
        </div>

        <ul className="space-y-3">
          {shortcuts.map((shortcut) => (
            <li key={shortcut.keys} className="flex items-center justify-between gap-3">
              <kbd className="rounded border border-border-default bg-surface-sunken px-2 py-1 font-mono text-body-sm text-text-primary">
                {shortcut.keys}
              </kbd>
              <span className="text-body-md text-text-secondary">{shortcut.action}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
