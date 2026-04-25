import { cn } from "@/lib/utils";

export type JobStatus =
  | "queued"
  | "running"
  | "succeeded"
  | "failed"
  | "canceled"
  | "paused"
  | "checkpointed";

const statusLabel: Record<JobStatus, string> = {
  queued: "Queued",
  running: "Running",
  succeeded: "Succeeded",
  failed: "Failed",
  canceled: "Canceled",
  paused: "Paused",
  checkpointed: "Checkpointed",
};

const statusClassName: Record<JobStatus, string> = {
  queued: "bg-neutral-700 text-neutral-300 ring-neutral-500/60",
  running: "bg-info-solid/20 text-info-solid ring-info-solid/40",
  succeeded: "bg-success-solid/20 text-success-solid ring-success-solid/40",
  failed: "bg-danger-solid/20 text-danger-solid ring-danger-solid/40",
  canceled: "bg-neutral-700 text-neutral-300 ring-neutral-500/60",
  paused: "bg-warning-solid/20 text-warning-solid ring-warning-solid/40",
  checkpointed: "bg-warning-solid/20 text-warning-solid ring-warning-solid/40",
};

export function JobStatusPill({ status, className }: { status: JobStatus; className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-1 text-caption font-medium ring-1 ring-inset",
        statusClassName[status],
        className
      )}
      aria-label={`Job status: ${statusLabel[status]}`}
    >
      <span
        className={cn(
          "mr-1.5 inline-block h-1.5 w-1.5 rounded-full",
          status === "running" && "animate-pulse"
        )}
      />
      {statusLabel[status]}
    </span>
  );
}
