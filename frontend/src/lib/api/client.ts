import createClient from "openapi-fetch";

import type { components, paths } from "./generated/schema";

const DEFAULT_API_BASE_URL = "http://localhost:8080";
const REQUEST_ID_HEADER = "x-request-id";

export type ApiProblemDetails = components["schemas"]["ProblemDetails"];

export class ApiClientError extends Error {
  public readonly status: number;
  public readonly requestId: string | null;
  public readonly problem?: ApiProblemDetails;

  public constructor({
    status,
    message,
    requestId,
    problem,
  }: {
    status: number;
    message: string;
    requestId: string | null;
    problem?: ApiProblemDetails;
  }) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
    this.requestId = requestId;
    this.problem = problem;
  }
}

type GetAuthToken = () => string | undefined | Promise<string | undefined>;

export function createApiClient({
  baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL,
  getAuthToken,
  fetchImpl = fetch,
}: {
  baseUrl?: string;
  getAuthToken?: GetAuthToken;
  fetchImpl?: typeof fetch;
} = {}) {
  return createClient<paths>({
    baseUrl,
    fetch: async (request: Request) => {
      const headers = new Headers(request.headers);
      headers.set(REQUEST_ID_HEADER, crypto.randomUUID());

      const token = await getAuthToken?.();
      if (token) {
        headers.set("authorization", `Bearer ${token}`);
      }

      return fetchImpl(new Request(request, { headers }));
    },
  });
}

export const apiClient = createApiClient();

export async function unwrapApiResult<TData, TError>(
  result: { data?: TData; error?: TError; response: Response },
  fallbackMessage: string
): Promise<TData> {
  if (result.error !== undefined || !result.response.ok || result.data === undefined) {
    throw await normalizeApiError(result.response, fallbackMessage, result.error);
  }

  return result.data;
}

export async function getVersion(
  client = apiClient
): Promise<components["schemas"]["VersionInfo"]> {
  const result = await client.GET("/version");
  return unwrapApiResult(result, "Failed to fetch API version");
}

async function normalizeApiError(
  response: Response,
  fallbackMessage: string,
  payloadError?: unknown
): Promise<ApiClientError> {
  const requestId = response.headers.get(REQUEST_ID_HEADER);
  const messageParts = [fallbackMessage];

  const problem = isProblemDetails(payloadError)
    ? payloadError
    : await readProblemDetails(response);
  if (problem?.title) {
    messageParts.push(problem.title);
  } else if (response.statusText) {
    messageParts.push(response.statusText);
  }

  return new ApiClientError({
    status: response.status,
    requestId,
    problem,
    message: messageParts.join(": "),
  });
}

async function readProblemDetails(response: Response): Promise<ApiProblemDetails | undefined> {
  const contentType = response.headers.get("content-type") ?? "";
  if (
    !contentType.includes("application/problem+json") &&
    !contentType.includes("application/json")
  ) {
    return undefined;
  }

  try {
    const payload: unknown = await response.clone().json();
    if (isProblemDetails(payload)) {
      return payload;
    }
  } catch {
    return undefined;
  }

  return undefined;
}

function isProblemDetails(value: unknown): value is ApiProblemDetails {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<ApiProblemDetails>;
  return (
    typeof candidate.type === "string" &&
    typeof candidate.title === "string" &&
    typeof candidate.status === "number"
  );
}
