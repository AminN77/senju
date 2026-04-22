import { FlaskConical } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

export default function Home() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-col gap-6 px-6 py-20">
      <h1 className="text-body-md font-semibold text-text-primary">shadcn/ui primitive smoke route</h1>
      <p className="text-body-md text-text-secondary">
        Button, Input, Label, and Separator render from source in `src/components/ui/`.
      </p>

      <section className="grid gap-6 md:grid-cols-2">
        <article className="rounded-md border border-border-default bg-surface-raised p-6 shadow-1">
          <h2 className="mb-4 text-body-md font-medium text-text-primary">Dark theme</h2>
          <div className="space-y-4">
            <Label htmlFor="dark-email">Email</Label>
            <Input id="dark-email" placeholder="name@senju.dev" type="email" />
            <Separator />
            <Button>
              <FlaskConical className="h-4 w-4" />
              Run smoke action
            </Button>
          </div>
        </article>

        <article data-theme="light" className="rounded-md border border-border-default bg-surface-base p-6 shadow-1">
          <h2 className="mb-4 text-body-md font-medium text-text-primary">Light theme</h2>
          <div className="space-y-4">
            <Label htmlFor="light-email">Email</Label>
            <Input id="light-email" placeholder="name@senju.dev" type="email" />
            <Separator />
            <Button variant="secondary">
              <FlaskConical className="h-4 w-4" />
              Run smoke action
            </Button>
          </div>
        </article>
      </section>
    </main>
  );
}
