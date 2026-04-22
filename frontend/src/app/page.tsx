import { FlaskConical } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ThemeSmokeToggle } from "@/components/theme-smoke-toggle";

export default function Home() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-col gap-6 px-6 py-20">
      <h1 className="text-body-md font-semibold text-text-primary">shadcn/ui primitive smoke route</h1>
      <p className="text-body-md text-text-secondary">
        Button, Input, Label, and Separator render from source in `src/components/ui/`. Use the theme buttons to
        preview light and dark on the page — `data-theme` is only set on <code className="font-mono">&lt;html&gt;</code>.
      </p>

      <div className="space-y-4">
        <ThemeSmokeToggle />
        <article className="rounded-md border border-border-default bg-surface-raised p-6 shadow-1">
          <h2 className="mb-4 text-body-md font-medium text-text-primary">Theme preview</h2>
          <div className="space-y-4">
            <Label htmlFor="smoke-email">Email</Label>
            <Input id="smoke-email" placeholder="name@senju.dev" type="email" />
            <Separator />
            <div className="flex flex-wrap gap-2">
              <Button>
                <FlaskConical className="h-4 w-4" />
                Run smoke action
              </Button>
              <Button variant="secondary">
                <FlaskConical className="h-4 w-4" />
                Secondary
              </Button>
            </div>
          </div>
        </article>
      </div>
    </main>
  );
}
