import { gzipSync } from "node:zlib";
import { readFileSync } from "node:fs";
import { join } from "node:path";

const CLIENT_REFERENCE_MANIFEST_PATH = join(process.cwd(), ".next", "server", "app", "page_client-reference-manifest.js");
const BUDGETS = [
  {
    name: "marketing",
    maxBytes: 150 * 1024,
    appEntries: ["[project]/src/app/page"],
    required: true,
  },
  {
    name: "console",
    maxBytes: 250 * 1024,
    appEntries: ["[project]/src/app/dashboard/page", "[project]/src/app/jobs/page"],
    required: false,
  },
];

function readClientReferenceManifest() {
  const raw = readFileSync(CLIENT_REFERENCE_MANIFEST_PATH, "utf8");
  const marker = 'globalThis.__RSC_MANIFEST["/page"] = ';
  const start = raw.indexOf(marker);
  if (start === -1) {
    throw new Error("Unable to locate __RSC_MANIFEST payload in page_client-reference-manifest.js");
  }
  const jsonPayload = raw.slice(start + marker.length, raw.lastIndexOf(";")).trim();
  return JSON.parse(jsonPayload);
}

function entryFiles(manifest, appEntry) {
  const entryJSFiles = manifest.entryJSFiles ?? {};
  const files = entryJSFiles[appEntry] ?? [];
  return files.filter((file) => file.endsWith(".js"));
}

function gzipBytesForFiles(files) {
  const seen = new Set();
  let total = 0;

  for (const relFile of files) {
    if (seen.has(relFile)) {
      continue;
    }
    seen.add(relFile);
    const absFile = join(process.cwd(), ".next", relFile.startsWith("/") ? relFile.slice(1) : relFile);
    const source = readFileSync(absFile);
    total += gzipSync(source).byteLength;
  }

  return total;
}

const manifest = readClientReferenceManifest();
const failures = [];

for (const budget of BUDGETS) {
  const matchingEntries = budget.appEntries
    .map((entry) => ({ entry, files: entryFiles(manifest, entry) }))
    .filter(({ files }) => files.length > 0);

  if (matchingEntries.length === 0) {
    if (budget.required) {
      failures.push(`${budget.name} (no matching app entry found)`);
      continue;
    }

    console.log(`[bundle-budget] skipped '${budget.name}' (no matching app entry found yet)`);
    continue;
  }

  for (const { entry, files } of matchingEntries) {
    const gzippedBytes = gzipBytesForFiles(files);
    const withinBudget = gzippedBytes <= budget.maxBytes;
    const status = withinBudget ? "PASS" : "FAIL";

    console.log(
      `[bundle-budget] ${status} ${budget.name} entry '${entry}': ${gzippedBytes} bytes gzip (limit ${budget.maxBytes})`
    );

    if (!withinBudget) {
      failures.push(`${budget.name}:${entry} (${gzippedBytes} > ${budget.maxBytes})`);
    }
  }
}

if (failures.length > 0) {
  console.error(`[bundle-budget] budget exceeded: ${failures.join(", ")}`);
  process.exit(1);
}
