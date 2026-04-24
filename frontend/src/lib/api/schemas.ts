import { z } from "zod";

import type { components } from "./generated/schema";

export type FastqUploadMetadataRequest = components["schemas"]["FastqUploadMetadataRequest"];

export const fastqUploadMetadataRequestSchema: z.ZodType<FastqUploadMetadataRequest> = z.object({
  sample_id: z.string().trim().min(1),
  r1_uri: z.string().trim().min(1),
  r2_uri: z.string().trim().min(1),
  library_id: z.string().optional(),
  platform: z.string().optional(),
  notes: z.string().optional(),
});

export type VariantQueryFilters = components["schemas"]["VariantQueryFilters"];

export const variantQueryFiltersSchema: z.ZodType<VariantQueryFilters> = z.object({
  chromosome: z.string().optional(),
  position_min: z.number().int().nonnegative().optional(),
  position_max: z.number().int().nonnegative().optional(),
  gene: z.string().optional(),
});
