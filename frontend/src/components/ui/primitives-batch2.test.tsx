import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { axe } from "../../../tests/utils/axe";
import { render } from "../../../tests/utils/render";
import { Button } from "./button";
import { Dialog, DialogContent, DialogDescription, DialogTitle, DialogTrigger } from "./dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./tabs";

describe("core primitives batch 2", () => {
  it("dialog opens and closes with escape key", async () => {
    const user = userEvent.setup();
    render(
      <Dialog>
        <DialogTrigger asChild>
          <Button>Open dialog</Button>
        </DialogTrigger>
        <DialogContent>
          <DialogTitle>Title</DialogTitle>
          <DialogDescription>Description</DialogDescription>
        </DialogContent>
      </Dialog>
    );

    await user.click(screen.getByRole("button", { name: "Open dialog" }));
    expect(screen.getByText("Description")).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expect(screen.queryByText("Description")).not.toBeInTheDocument();
  });

  it("tabs switch content and remain accessible", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <Tabs defaultValue="a">
        <TabsList>
          <TabsTrigger value="a">Overview</TabsTrigger>
          <TabsTrigger value="b">Logs</TabsTrigger>
        </TabsList>
        <TabsContent value="a">Overview body</TabsContent>
        <TabsContent value="b">Logs body</TabsContent>
      </Tabs>
    );

    await user.click(screen.getByRole("tab", { name: "Logs" }));
    expect(screen.getByText("Logs body")).toBeInTheDocument();
    expect(await axe(container)).toHaveNoViolations();
  });
});
