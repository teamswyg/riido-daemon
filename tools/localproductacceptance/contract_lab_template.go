package main

import "html/template"

var contractLabTemplate = template.Must(template.New("contract-lab").Parse(`<!doctype html>
<html lang="ko">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Riido AI Agent Contract Lab</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f6f7f9;
      --panel: #ffffff;
      --ink: #17202f;
      --muted: #5c6778;
      --line: #d8dee8;
      --line-soft: #e8ecf2;
      --accent: #175e7a;
      --ok: #168a4a;
      --warn: #a16207;
      --bad: #ba3328;
      --info: #2457a7;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      background: var(--bg);
      color: var(--ink);
      font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 14px;
    }
    main { max-width: 1440px; margin: 0 auto; padding: 24px; }
    h1, h2, h3 { margin: 0; line-height: 1.2; }
    h1 { font-size: 24px; }
    h2 { font-size: 18px; }
    h3 { font-size: 15px; }
    p { margin: 0; line-height: 1.5; color: var(--muted); }
    button, input, select {
      font: inherit;
      border: 1px solid var(--line);
      border-radius: 6px;
      background: var(--panel);
      color: var(--ink);
    }
    button { cursor: pointer; padding: 6px 10px; }
    button[aria-pressed="true"] { border-color: var(--accent); color: var(--accent); font-weight: 700; }
    input, select { min-height: 34px; padding: 6px 8px; }
    code, pre { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 12px; }
    pre {
      margin: 0;
      padding: 12px;
      overflow: auto;
      white-space: pre-wrap;
      word-break: break-word;
      background: #f1f4f8;
      border: 1px solid var(--line-soft);
      border-radius: 6px;
      max-height: 420px;
    }
    table { width: 100%; border-collapse: collapse; background: var(--panel); border: 1px solid var(--line); }
    th, td { padding: 10px; border-bottom: 1px solid var(--line-soft); text-align: left; vertical-align: top; }
    th { font-size: 12px; color: var(--muted); background: #fbfcfe; }
    .shell { display: grid; gap: 16px; }
    .topbar, .panel, .toolbar {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 16px;
    }
    .topbar { display: grid; gap: 14px; }
    .title-row, .toolbar, .split, .tabs, .metric-grid, .area-grid, .scenario-row, .kv-grid, .manual-grid, .domain-grid, .field-row {
      display: grid;
      gap: 10px;
    }
    .title-row { grid-template-columns: 1fr auto; align-items: start; }
    .tabs { grid-template-columns: repeat(6, minmax(0, 1fr)); }
    .toolbar { grid-template-columns: 1fr 180px 220px; align-items: center; }
    .metric-grid { grid-template-columns: repeat(5, minmax(0, 1fr)); }
    .area-grid { grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); }
    .split { grid-template-columns: minmax(360px, 0.9fr) minmax(420px, 1.1fr); align-items: start; }
    .manual-grid { grid-template-columns: minmax(340px, 0.85fr) minmax(460px, 1.15fr); align-items: start; }
    .domain-grid { grid-template-columns: minmax(360px, 0.9fr) minmax(460px, 1.1fr); align-items: start; }
    .scenario-row { grid-template-columns: 1fr auto; align-items: start; }
    .kv-grid { grid-template-columns: 180px 1fr; }
    .field-row { grid-template-columns: 140px 1fr; align-items: center; }
    .metric, .area, .scenario, .intent-row {
      border: 1px solid var(--line-soft);
      border-radius: 8px;
      background: var(--panel);
      padding: 12px;
    }
    .metric strong { display: block; font-size: 20px; margin-top: 4px; }
    .area { display: grid; gap: 10px; align-content: start; }
    .area header, .scenario header { display: flex; gap: 8px; align-items: center; justify-content: space-between; }
    .scenario { display: grid; gap: 10px; margin-bottom: 8px; }
    .pill {
      display: inline-flex;
      align-items: center;
      min-height: 20px;
      padding: 2px 8px;
      border-radius: 999px;
      background: #eef2f7;
      color: var(--muted);
      font-size: 12px;
      font-weight: 700;
      text-transform: uppercase;
      white-space: nowrap;
    }
    .pill.passed { color: var(--ok); background: #e8f6ee; }
    .pill.failed { color: var(--bad); background: #fae9e7; }
    .pill.skipped, .pill.partial { color: var(--warn); background: #fbf1db; }
    .pill.expired { color: var(--bad); background: #fae9e7; }
    .pill.fresh { color: var(--ok); background: #e8f6ee; }
    .stack { display: grid; gap: 8px; }
    .muted { color: var(--muted); }
    .small { font-size: 12px; }
    .list { display: flex; flex-wrap: wrap; gap: 6px; }
    .manual-actions { display: flex; flex-wrap: wrap; gap: 8px; }
    .manual-actions button { min-width: 78px; }
    .manual-actions .danger { border-color: var(--bad); color: var(--bad); }
    textarea {
      width: 100%;
      min-height: 120px;
      resize: vertical;
      font: inherit;
      border: 1px solid var(--line);
      border-radius: 6px;
      padding: 8px;
      color: var(--ink);
      background: var(--panel);
    }
    .link-list { display: grid; gap: 6px; }
    .divider { height: 1px; background: var(--line-soft); margin: 4px 0; }
    .shot {
      display: block;
      width: 100%;
      max-height: 280px;
      object-fit: contain;
      background: #f1f4f8;
      border: 1px solid var(--line);
      border-radius: 6px;
      margin-top: 8px;
    }
    .sequence { display: grid; gap: 8px; }
    .sequence-step {
      display: grid;
      grid-template-columns: 48px 1fr auto;
      gap: 10px;
      align-items: center;
      border: 1px solid var(--line-soft);
      border-radius: 8px;
      background: var(--panel);
      padding: 10px;
    }
    .step-num { color: var(--muted); font-weight: 700; }
    .journey-steps { display: grid; gap: 8px; }
    .journey-step { display: grid; grid-template-columns: 28px 1fr auto; gap: 10px; align-items: center; text-align: left; min-height: 48px; }
    .journey-step[aria-pressed="true"] { background: #eef6f8; }
    .journey-page { border: 1px solid var(--line); border-radius: 8px; padding: 14px; background: #fbfcfe; }
    .journey-flow { display: grid; gap: 10px; grid-template-columns: repeat(3, minmax(0, 1fr)); }
    .runtime-detail { border-color: var(--accent); box-shadow: inset 4px 0 0 var(--accent); }
    .empty { padding: 18px; border: 1px dashed var(--line); border-radius: 8px; color: var(--muted); background: #fbfcfe; }
    @media (max-width: 900px) {
      main { padding: 16px; }
      .title-row, .toolbar, .split, .tabs, .metric-grid, .kv-grid, .manual-grid, .domain-grid, .journey-flow, .field-row { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body><main id="root"></main>
<script type="application/json" id="evidence">{{.Evidence}}</script>
<script type="application/json" id="qa-i18n">{{.I18N}}</script>
<script type="module">
	import React, { useMemo, useRef, useState } from "https://esm.sh/react@18.3.1";
import { createRoot } from "https://esm.sh/react-dom@18.3.1/client";

const h = React.createElement;
const evidence = JSON.parse(document.getElementById("evidence").textContent);
const i18nSpec = JSON.parse(document.getElementById("qa-i18n").textContent);
const runtimeDetailNodeID = "1179:27360";
const figmaURL = "https://www.figma.com/design/MUOd9lctoEHASUStN3vUuK/v.1.22-AI-Agent?node-id=1179-27360&m=dev";
const manualStorageKey = "riido-contract-lab-manual-qa-v1";
const domainStorageKey = "riido-contract-lab-domain-fixture-journey-v1";
const manualStatuses = ["untested", "running", "passed", "failed", "blocked"];
const locale = i18nSpec.default_locale || "ko";
const fallbackLocale = i18nSpec.fallback_locale || "en";
const messages = Object.fromEntries((i18nSpec.namespaces || []).flatMap((ns) =>
  (ns.messages || []).map(([key, ko, en]) => [ns.id + "." + key, { ko, en }])
));

function t(key, vars = {}) {
  const entry = messages[key] || {};
  const text = entry[locale] || entry[fallbackLocale] || key;
  return text.replace(/\{(\w+)\}/g, (_, name) => String(vars[name] ?? ""));
}

function statusText(status) {
  return t("status." + (status || "unknown"));
}

document.title = t("app.title");

const functionalAreas = [
  {
    id: "bootstrap-read-model",
    titleKey: "area.bootstrap.title",
    intentKey: "area.bootstrap.intent",
    match: ["contract.api.bootstrap", "contract.api.devices", "local.saas.runtime_snapshot.wait"]
  },
  {
    id: "domain-fixture",
    titleKey: "area.domain.title",
    intentKey: "area.domain.intent",
    match: ["domain.fixture"]
  },
  {
    id: "daemon-device-lifecycle",
    titleKey: "area.lifecycle.title",
    intentKey: "area.lifecycle.intent",
    match: ["local.saas.device_enroll.", "local.saas.daemon_binary.build", "local.saas.daemon_start.", "local.saas.daemon_status.", "local.saas.daemon_cleanup."]
  },
  {
    id: "agent-catalog-profile",
    titleKey: "area.catalog.title",
    intentKey: "area.catalog.intent",
    match: ["local.saas.agent_fixture.", "contract.task.frontend_identity_contract", "contract.api.profile_thumbnail.intent"]
  },
  {
    id: "task-assignment",
    titleKey: "area.assignment.title",
    intentKey: "area.assignment.intent",
    match: ["contract.task.discovery", "contract.task.fixture.", "contract.task.assignable_agents", "contract.task.assignment.create.", "contract.task.multi_assignment"]
  },
  {
    id: "thread-stream",
    titleKey: "area.thread.title",
    intentKey: "area.thread.intent",
    match: ["contract.task.thread_subscription", "contract.task.sse_replay", "contract.task.thread_message", "contract.task.assignment.cleanup."]
  },
  {
    id: "figma-intent",
    titleKey: "area.figma.title",
    intentKey: "area.figma.intent",
    match: ["figma.intent.catalog", "figma.onboarding", "figma.runtime.settings", "figma.runtime.detail"]
  },
  {
    id: "qa-loop",
    titleKey: "area.loop.title",
    intentKey: "area.loop.intent",
    match: ["contract.ui.", "local.qa.", "infra.local_qa_dashboard", "release."]
  }
];

function scenarios() {
  return Array.isArray(evidence.scenarios) ? evidence.scenarios : [];
}

function screenshotHref(path) {
  const prefix = ".riido-local/screenshots/";
  return path && path.startsWith(prefix) ? "../screenshots/" + path.slice(prefix.length) : path;
}

function manualEvidencePath() {
  const loopPath = scenarios().find((scenario) => scenario.id === "local.qa.loop.freshness")?.observed?.manual_evidence;
  const labPath = scenarios().find((scenario) => scenario.id === "contract.ui.lab")?.observed?.manual_evidence;
  return loopPath || labPath || ".riido-local/evidence/manual-qa-evidence.json";
}

function manualEvidenceFileName() {
  const parts = manualEvidencePath().split("/");
  return parts[parts.length - 1] || "manual-qa-evidence.json";
}

function domainJourneyScenario() {
  return scenarios().find((scenario) => scenario.id === "domain.fixture_journey") || { observed: {} };
}

function domainEntityScenarios() {
  return scenarios().filter((scenario) => scenario.id.startsWith("domain.fixture.") && scenario.id !== "domain.fixture_journey");
}

function domainCachePath() {
  return domainJourneyScenario().observed?.cache_path || ".riido-local/evidence/domain-fixture-journey-cache.json";
}

function domainEntityKey(scenario) {
  return scenario.id.replace("domain.fixture.", "");
}

function domainStepVerb(key) {
  return t("domain.entity." + key);
}

function domainFigmaEvidence(key) {
  const entries = figmaEntries();
  const direct = entries.find((entry) => JSON.stringify(entry).toLowerCase().includes(key));
  return direct || entries.find((entry) => entry.node_id === runtimeDetailNodeID) || entries[0];
}

function readDomainDraft() {
  try {
    const raw = window.localStorage.getItem(domainStorageKey);
    if (raw) {
      const parsed = JSON.parse(raw);
      return { environment: parsed.environment || "staging", entities: parsed.entities || {} };
    }
  } catch (err) {
    console.warn("domain journey draft reset", err);
  }
  return { environment: "staging", entities: {} };
}

function persistDomainDraft(draft) {
  window.localStorage.setItem(domainStorageKey, JSON.stringify(draft));
}

function buildDomainCache(draft) {
  return {
    schema_version: "riido-domain-fixture-cache.v1",
    id: "domain-fixture-journey",
    name: "Domain Fixture Journey",
    common_name: "도메인 픽스처 여정",
    observed_at: new Date().toISOString(),
    expires_at: evidence.expires_at || "",
    environment: "staging",
    verification_source: "local",
    source_evidence: evidence.id || "ai-agent-product-acceptance",
    entities: draft.entities || {}
  };
}

function downloadDomainCache(draft) {
  const body = JSON.stringify(buildDomainCache(draft), null, 2);
  const blob = new Blob([body], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "domain-fixture-journey-cache.json";
  link.click();
  URL.revokeObjectURL(url);
}

function s3DomainUploadCommand() {
  return "aws s3 cp " + domainCachePath() + " ${RIIDO_LOCAL_QA_S3_PREFIX:-s3://<local-qa-evidence-bucket>/daily}/latest/domain-fixture-journey-cache.json --cache-control no-store";
}

function readManualDraft() {
  try {
    const raw = window.localStorage.getItem(manualStorageKey);
    if (raw) {
      const parsed = JSON.parse(raw);
      return {
        tester: parsed.tester || "",
        environment: parsed.environment || "local",
        records: parsed.records || {}
      };
    }
  } catch (err) {
    console.warn("manual QA draft reset", err);
  }
  return { tester: "", environment: "local", records: {} };
}

function persistManualDraft(draft) {
  window.localStorage.setItem(manualStorageKey, JSON.stringify(draft));
}

function manualRecordFor(draft, scenario) {
  const existing = draft.records?.[scenario.id] || {};
  return {
    id: scenario.id,
    area: classifyScenario(scenario.id),
    source_status: scenario.status || "unknown",
    method: scenario.method || "",
    endpoint: scenario.endpoint || "",
    manual_status: existing.manual_status || "untested",
    note: existing.note || "",
    tested_at: existing.tested_at || "",
    tester: existing.tester || draft.tester || "",
    environment: existing.environment || draft.environment || "local"
  };
}

function manualStatusSummary(records) {
  const summary = manualStatuses.reduce((acc, status) => ({ ...acc, [status]: 0 }), { total: 0 });
  for (const record of Object.values(records || {})) {
    const status = record.manual_status || "untested";
    summary[status] = (summary[status] || 0) + 1;
    summary.total += 1;
  }
  return summary;
}

function manualEvidenceStatus(records) {
  const values = Object.values(records || {});
  if (values.length === 0) return "partial";
  if (values.some((record) => record.manual_status === "failed" || record.manual_status === "blocked")) return "partial";
  if (values.every((record) => record.manual_status === "passed")) return "passed";
  return "partial";
}

function buildManualEvidence(draft) {
  const records = Object.values(draft.records || {}).sort((a, b) => a.id.localeCompare(b.id));
  const observedAt = new Date().toISOString();
  return {
    schema_version: "riido-manual-qa-evidence.v1",
    id: "manual-qa-evidence",
    observed_at: observedAt,
    expires_at: evidence.expires_at || "",
    status: manualEvidenceStatus(draft.records),
    source_evidence: evidence.id || "ai-agent-product-acceptance",
    source_observed_at: evidence.observed_at || "",
    manual_evidence_path: manualEvidencePath(),
    tester: draft.tester || "",
    environment: draft.environment || "local",
    scenarios: records
  };
}

function downloadManualEvidence(draft) {
  const body = JSON.stringify(buildManualEvidence(draft), null, 2);
  const blob = new Blob([body], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = manualEvidenceFileName();
  link.click();
  URL.revokeObjectURL(url);
}

function s3ManualUploadCommand() {
  const path = manualEvidencePath();
  return "aws s3 cp " + path + " ${RIIDO_LOCAL_QA_S3_PREFIX:-s3://<local-qa-evidence-bucket>/daily}/latest/manual-qa-evidence.json --cache-control no-store";
}

function areaScenarios(area) {
  return scenarios().filter((scenario) => area.match.some((needle) => scenario.id === needle || scenario.id.startsWith(needle)));
}

function statusCounts(items) {
  return items.reduce((acc, item) => {
    const status = item.status || "unknown";
    acc[status] = (acc[status] || 0) + 1;
    acc.total += 1;
    return acc;
  }, { total: 0 });
}

function freshness(expiresAt) {
  const expires = Date.parse(expiresAt || "");
  if (Number.isNaN(expires)) return { state: "unknown", label: t("freshness.unknown"), remaining: "" };
  const now = Date.now();
  if (now >= expires) return { state: "expired", label: t("status.expired"), remaining: t("freshness.expiredHint") };
  const ms = expires - now;
  const hours = Math.floor(ms / 3600000);
  const minutes = Math.floor((ms % 3600000) / 60000);
  return { state: "fresh", label: t("status.fresh"), remaining: t("freshness.until", { hours, minutes }) };
}

function uniqueBy(items, keyFn) {
  const seen = new Set();
  const out = [];
  for (const item of items) {
    const key = keyFn(item);
    if (!key || seen.has(key)) continue;
    seen.add(key);
    out.push(item);
  }
  return out;
}

function figmaEntries() {
  const fromCatalog = scenarios().find((scenario) => scenario.id === "figma.intent.catalog")?.observed?.entries;
  if (Array.isArray(fromCatalog) && fromCatalog.length > 0) return fromCatalog;
  return uniqueBy(
    scenarios().flatMap((scenario) => Array.isArray(scenario.observed?.entries) ? scenario.observed.entries : []),
    (entry) => entry.node_id
  );
}

function classifyScenario(id) {
  const area = functionalAreas.find((candidate) => candidate.match.some((needle) => id === needle || id.startsWith(needle)));
  return area ? t(area.titleKey) : t("area.other");
}

function matchesQuery(scenario, query) {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  const blob = [
    scenario.id,
    scenario.status,
    scenario.method,
    scenario.endpoint,
    scenario.failure_summary,
    classifyScenario(scenario.id),
    JSON.stringify(scenario.observed || {})
  ].join(" ").toLowerCase();
  return blob.includes(q);
}

function StatusPill({ status }) {
  return h("span", { className: "pill " + (status || "unknown") }, statusText(status));
}

function KeyValue({ label, value }) {
  return h("div", { className: "kv-grid" }, [
    h("div", { key: "label", className: "muted" }, label),
    h("div", { key: "value" }, value || h("span", { className: "muted" }, t("state.notObserved")))
  ]);
}

function Summary() {
  const counts = statusCounts(scenarios());
  const fresh = freshness(evidence.expires_at);
  return h("section", { className: "topbar" }, [
    h("div", { key: "title", className: "title-row" }, [
      h("div", { key: "copy", className: "stack" }, [
        h("h1", { key: "h" }, t("app.title")),
        h("p", { key: "p" }, t("app.subtitle"))
      ]),
      h("div", { key: "status", className: "list" }, [
        h(StatusPill, { key: "run", status: evidence.status }),
        h("span", { key: "fresh", className: "pill " + fresh.state }, fresh.label)
      ])
    ]),
    h("div", { key: "metrics", className: "metric-grid" }, [
      h("div", { key: "total", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.scenarios")), h("strong", { key: "v" }, String(counts.total))]),
      h("div", { key: "passed", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.passed")), h("strong", { key: "v" }, String(counts.passed || 0))]),
      h("div", { key: "partial", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.partialSkipped")), h("strong", { key: "v" }, String((counts.partial || 0) + (counts.skipped || 0)))]),
      h("div", { key: "failed", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.failed")), h("strong", { key: "v" }, String(counts.failed || 0))]),
      h("div", { key: "expiry", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.freshness")), h("strong", { key: "v" }, fresh.remaining || fresh.label)])
    ]),
    h("div", { key: "kv", className: "kv-grid" }, [
      h("div", { key: "ol", className: "muted" }, t("label.observed")),
      h("code", { key: "ov" }, evidence.observed_at || t("status.unknown")),
      h("div", { key: "el", className: "muted" }, t("label.expires")),
      h("code", { key: "ev" }, evidence.expires_at || t("status.unknown"))
    ])
  ]);
}

function Tabs({ active, onChange }) {
  const tabs = [
    ["areas", "tab.areas"],
    ["sequence", "tab.sequence"],
    ["domain", "tab.domain"],
    ["manual", "tab.manual"],
    ["figma", "tab.figma"],
    ["loop", "tab.loop"]
  ];
  return h("nav", { className: "tabs" }, tabs.map(([id, label]) =>
    h("button", { key: id, type: "button", "aria-pressed": active === id, onClick: () => onChange(id) }, t(label))
  ));
}

function AreaCard({ area, active, onSelect }) {
  const items = areaScenarios(area);
  const counts = statusCounts(items);
  const endpoints = uniqueBy(items.filter((item) => item.endpoint).map((item) => item.endpoint), (endpoint) => endpoint).slice(0, 4);
  return h("section", { className: "area", "data-area": area.id }, [
    h("header", { key: "h" }, [
      h("h3", { key: "title" }, t(area.titleKey)),
      h(StatusPill, { key: "status", status: counts.failed ? "failed" : counts.total && counts.passed === counts.total ? "passed" : "partial" })
    ]),
    h("p", { key: "intent" }, t(area.intentKey)),
    h("div", { key: "counts", className: "list" }, [
      h("span", { key: "total", className: "pill" }, t("count.scenarios", { count: counts.total })),
      h("span", { key: "passed", className: "pill passed" }, t("count.passed", { count: counts.passed || 0 })),
      (counts.skipped || counts.partial) ? h("span", { key: "partial", className: "pill partial" }, t("count.partial", { count: (counts.skipped || 0) + (counts.partial || 0) })) : null,
      counts.failed ? h("span", { key: "failed", className: "pill failed" }, t("count.failed", { count: counts.failed })) : null
    ]),
    endpoints.length ? h("div", { key: "endpoints", className: "link-list small" }, endpoints.map((endpoint) => h("code", { key: endpoint }, endpoint))) : h("div", { key: "empty", className: "small muted" }, t("state.noEndpoint")),
    h("button", { key: "button", type: "button", "aria-pressed": active, onClick: onSelect }, active ? t("action.selected") : t("action.inspect"))
  ]);
}

function ScenarioList({ items, selectedID, onSelect }) {
  if (items.length === 0) return h("div", { className: "empty" }, t("state.noScenariosFilter"));
  return h("div", { className: "stack" }, items.map((scenario) =>
    h("section", { key: scenario.id + (scenario.endpoint || ""), className: "scenario" }, [
      h("header", { key: "head" }, [
        h("div", { key: "title", className: "stack" }, [
          h("strong", { key: "id" }, scenario.id),
          h("span", { key: "area", className: "small muted" }, classifyScenario(scenario.id))
        ]),
        h(StatusPill, { key: "status", status: scenario.status })
      ]),
      scenario.method || scenario.endpoint ? h("code", { key: "endpoint" }, [scenario.method, scenario.endpoint].filter(Boolean).join(" ")) : null,
      scenario.failure_summary ? h("p", { key: "failure" }, scenario.failure_summary) : null,
      scenario.repair ? h("p", { key: "repair" }, scenario.repair.summary) : null,
      h("button", { key: "button", type: "button", "aria-pressed": selectedID === scenario.id, onClick: () => onSelect(scenario.id) }, t("action.open"))
    ])
  ));
}

function ScenarioDetail({ scenario }) {
  if (!scenario) return h("div", { className: "empty" }, t("state.selectScenario"));
  const shot = screenshotHref(scenario.screenshot);
  return h("section", { className: "panel stack" }, [
    h("div", { key: "head", className: "title-row" }, [
      h("div", { key: "title", className: "stack" }, [
        h("h2", { key: "id" }, scenario.id),
        h("p", { key: "area" }, classifyScenario(scenario.id))
      ]),
      h(StatusPill, { key: "status", status: scenario.status })
    ]),
    h(KeyValue, { key: "method", label: t("label.method"), value: scenario.method }),
    h(KeyValue, { key: "endpoint", label: t("label.endpoint"), value: scenario.endpoint }),
    scenario.failure_summary ? h(KeyValue, { key: "failure", label: t("label.failure"), value: scenario.failure_summary }) : null,
    scenario.repair ? h(KeyValue, { key: "repair", label: t("label.repair"), value: scenario.repair.summary }) : null,
    shot ? h("a", { key: "shot", href: shot }, [t("scenario.visualEvidence"), h("img", { key: "img", className: "shot", src: shot, alt: scenario.id + " screenshot" })]) : null,
    h("div", { key: "json", className: "stack" }, [
      h("h3", { key: "label" }, t("scenario.observed")),
      h("pre", { key: "pre" }, JSON.stringify(scenario.observed || {}, null, 2))
    ])
  ]);
}

function AreasView({ activeArea, setActiveArea, query, status, selectedID, setSelectedID }) {
  const area = functionalAreas.find((item) => item.id === activeArea) || functionalAreas[0];
  const filtered = areaScenarios(area)
    .filter((scenario) => status === "all" || scenario.status === status)
    .filter((scenario) => matchesQuery(scenario, query));
  const selected = filtered.find((scenario) => scenario.id === selectedID) || filtered[0];
  return h("div", { className: "stack" }, [
    h("div", { key: "areas", className: "area-grid" }, functionalAreas.map((item) =>
      h(AreaCard, { key: item.id, area: item, active: activeArea === item.id, onSelect: () => setActiveArea(item.id) })
    )),
    h("div", { key: "split", className: "split" }, [
      h("section", { key: "left", className: "panel stack" }, [
        h("h2", { key: "h" }, t(area.titleKey)),
        h(ScenarioList, { key: "list", items: filtered, selectedID: selected?.id, onSelect: setSelectedID })
      ]),
      h(ScenarioDetail, { key: "detail", scenario: selected })
    ])
  ]);
}

function SequenceView({ query, status, selectedID, setSelectedID }) {
  const filtered = scenarios()
    .filter((scenario) => status === "all" || scenario.status === status)
    .filter((scenario) => matchesQuery(scenario, query));
  const selected = filtered.find((scenario) => scenario.id === selectedID) || filtered[0];
  return h("div", { className: "split" }, [
    h("section", { key: "sequence", className: "panel stack" }, [
      h("h2", { key: "h" }, t("scenario.replayTitle")),
      h("div", { key: "items", className: "sequence" }, filtered.map((scenario, index) =>
        h("div", { key: scenario.id + index, className: "sequence-step" }, [
          h("div", { key: "n", className: "step-num" }, String(index + 1).padStart(2, "0")),
          h("div", { key: "body", className: "stack" }, [
            h("strong", { key: "id" }, scenario.id),
            h("span", { key: "meta", className: "small muted" }, [scenario.method, scenario.endpoint].filter(Boolean).join(" ") || classifyScenario(scenario.id))
          ]),
          h("button", { key: "button", type: "button", "aria-pressed": selected?.id === scenario.id, onClick: () => setSelectedID(scenario.id) }, t("action.open"))
        ])
      ))
    ]),
    h(ScenarioDetail, { key: "detail", scenario: selected })
  ]);
}

function FigmaView() {
  const entries = figmaEntries();
  if (entries.length === 0) return h("section", { className: "panel" }, h("div", { className: "empty" }, t("state.noFigmaEntries")));
  return h("section", { className: "panel stack" }, [
    h("div", { key: "head", className: "title-row" }, [
      h("div", { key: "copy", className: "stack" }, [
        h("h2", { key: "h" }, t("figma.title")),
        h("p", { key: "p" }, t("figma.desc"))
      ]),
      h("a", { key: "link", href: figmaURL }, t("figma.sourceNode"))
    ]),
    h("table", { key: "table" }, [
      h("thead", { key: "thead" }, h("tr", null, [
        h("th", { key: "node" }, t("figma.node")),
        h("th", { key: "screen" }, t("figma.screen")),
        h("th", { key: "daemon" }, t("figma.daemonBoundary")),
        h("th", { key: "facts" }, t("figma.facts"))
      ])),
      h("tbody", { key: "tbody" }, entries.map((entry) =>
        h("tr", { key: entry.node_id, className: entry.node_id === runtimeDetailNodeID ? "runtime-detail" : "" }, [
          h("td", { key: "node" }, h("code", null, entry.node_id)),
          h("td", { key: "screen" }, [
            h("strong", { key: "name" }, entry.name),
            entry.node_id === runtimeDetailNodeID ? h("div", { key: "mark", className: "pill" }, t("figma.requestedNode")) : null,
            h("div", { key: "owners", className: "small muted" }, (entry.upstream_owner || []).join(", "))
          ]),
          h("td", { key: "daemon" }, entry.daemon_scope),
          h("td", { key: "facts", className: "stack" }, [
            h("div", { key: "daemonfacts" }, [
              h("strong", { key: "l" }, t("figma.daemonFacts") + " "),
              h("span", { key: "v" }, (entry.daemon_consumed_facts || []).join(", ") || t("figma.none"))
            ]),
            h("div", { key: "clientfacts" }, [
              h("strong", { key: "l" }, t("figma.clientFacts") + " "),
              h("span", { key: "v" }, (entry.client_owned_facts || []).join(", ") || t("figma.none"))
            ])
          ])
        ])
      ))
    ])
  ]);
}

function LoopView() {
  const fresh = freshness(evidence.expires_at);
  const catalog = scenarios().find((scenario) => scenario.id === "figma.intent.catalog");
  const lab = scenarios().find((scenario) => scenario.id === "contract.ui.lab");
  return h("section", { className: "panel stack" }, [
    h("h2", { key: "h" }, t("loop.title")),
    h("div", { key: "fresh", className: "metric-grid" }, [
      h("div", { key: "observed", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("label.observed")), h("strong", { key: "v" }, evidence.observed_at || t("status.unknown"))]),
      h("div", { key: "expires", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("label.expires")), h("strong", { key: "v" }, evidence.expires_at || t("status.unknown"))]),
      h("div", { key: "state", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("label.browserFreshness")), h("strong", { key: "v" }, fresh.label)]),
      h("div", { key: "figma", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.figmaEntries")), h("strong", { key: "v" }, String(catalog?.observed?.entries_count || figmaEntries().length || 0))]),
      h("div", { key: "lab", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.labScenario")), h("strong", { key: "v" }, statusText(lab?.status || "missing"))])
    ]),
    h("table", { key: "table" }, [
      h("thead", { key: "thead" }, h("tr", null, [
        h("th", { key: "phase" }, t("loop.phase")),
        h("th", { key: "evidence" }, t("loop.checked")),
        h("th", { key: "refresh" }, t("loop.refresh"))
      ])),
      h("tbody", { key: "tbody" }, [
        ["loop.p1", "loop.e1", "loop.r1"],
        ["loop.p2", "loop.e2", "loop.r2"],
        ["loop.p3", "loop.e3", "loop.r3"],
        ["loop.p4", "loop.e4", "loop.r4"]
      ].map((row) => h("tr", { key: row[0] }, row.map((cell, index) => h("td", { key: index }, t(cell))))))
    ]),
    h("pre", { key: "command" }, "go run ./tools/localqarunner -run-product -strict-coverage")
  ]);
}

function DomainJourneyView({ draft, updateDraft }) {
  const [selectedKey, setSelectedKey] = useState(domainEntityKey(domainEntityScenarios()[0] || { id: "domain.fixture.account" }));
  const [copyState, setCopyState] = useState("");
  const fileInputRef = useRef(null);
  const items = domainEntityScenarios();
  const selected = items.find((scenario) => domainEntityKey(scenario) === selectedKey) || items[0];
  const selectedEntity = selected ? domainEntityKey(selected) : "";
  const current = draft.entities?.[selectedEntity] || {};
  const summary = domainJourneyScenario();
  const observed = selected?.observed || {};
  const pageIndex = Math.max(0, items.findIndex((scenario) => domainEntityKey(scenario) === selectedEntity));
  const figma = domainFigmaEvidence(selectedEntity);
  const prevKey = items[pageIndex - 1] ? domainEntityKey(items[pageIndex - 1]) : "";
  const nextKey = items[pageIndex + 1] ? domainEntityKey(items[pageIndex + 1]) : "";

  function patchEntity(patch) {
    if (!selectedEntity) return;
    updateDraft((prev) => ({
      ...prev,
      environment: "staging",
      entities: {
        ...(prev.entities || {}),
        [selectedEntity]: {
          ...(prev.entities?.[selectedEntity] || {}),
          ...patch,
          source: patch.source || prev.entities?.[selectedEntity]?.source || "human-verified",
          updated_at: new Date().toISOString()
        }
      }
    }));
  }

  function useObservedID() {
    patchEntity({
      id: observed.cached_id || observed.configured_id || "",
      name: observed.title || selectedEntity,
      source: observed.source || "contract-lab"
    });
  }

  async function copyDomainJSON() {
    await navigator.clipboard.writeText(JSON.stringify(buildDomainCache(draft), null, 2));
    setCopyState(t("domain.copied"));
  }

  async function copyDomainS3() {
    await navigator.clipboard.writeText(s3DomainUploadCommand());
    setCopyState(t("domain.s3Copied"));
  }

  function importDomainCache(file) {
    if (!file) return;
    file.text().then((text) => {
      const imported = JSON.parse(text);
      updateDraft(() => ({ environment: "staging", entities: imported.entities || {} }));
      setCopyState(t("domain.imported"));
    }).catch((err) => setCopyState(t("domain.importFailed", { message: err.message })));
  }

  function moveStep(key) {
    if (key) setSelectedKey(key);
  }

  if (items.length === 0) {
    return h("section", { className: "panel" }, h("div", { className: "empty" }, t("state.noDomainEvidence")));
  }

  return h("div", { className: "domain-grid" }, [
    h("section", { key: "left", className: "panel stack" }, [
      h("div", { key: "head", className: "title-row" }, [
        h("div", { key: "copy", className: "stack" }, [
          h("h2", { key: "h" }, t("domain.title")),
          h("p", { key: "p" }, t("domain.desc"))
        ]),
        h(StatusPill, { key: "status", status: summary.status || "partial" })
      ]),
      h("div", { key: "page", className: "journey-page stack" }, [
        h("div", { key: "count", className: "small muted" }, t("domain.page", { page: pageIndex + 1, total: items.length })),
        h("h3", { key: "verb" }, domainStepVerb(selectedEntity)),
        h("div", { key: "flow", className: "journey-flow" }, [
          h("div", { key: "remote", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("label.remote")), h("strong", { key: "v" }, "staging")]),
          h("div", { key: "local", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("label.verification")), h("strong", { key: "v" }, "local")]),
          h("div", { key: "cache", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("label.cache")), h("strong", { key: "v" }, current.id ? t("domain.ready") : t("domain.needed"))])
        ]),
        h("div", { key: "pager", className: "manual-actions" }, [
          h("button", { key: "prev", type: "button", disabled: !prevKey, onClick: () => moveStep(prevKey) }, t("action.previous")),
          h("button", { key: "next", type: "button", disabled: !nextKey, onClick: () => moveStep(nextKey) }, t("action.next"))
        ])
      ]),
      h("div", { key: "meta", className: "kv-grid" }, [
        h("div", { key: "envl", className: "muted" }, t("label.remote")),
        h("code", { key: "envv" }, summary.observed?.remote_environment || "staging"),
        h("div", { key: "locl", className: "muted" }, t("label.verification")),
        h("code", { key: "locv" }, summary.observed?.verification_source || "local"),
        h("div", { key: "cachel", className: "muted" }, t("label.cachePath")),
        h("code", { key: "cachev" }, domainCachePath())
      ]),
      h("div", { key: "actions", className: "manual-actions" }, [
        h("button", { key: "download", type: "button", onClick: () => downloadDomainCache(draft) }, t("action.downloadCache")),
        h("button", { key: "copy", type: "button", onClick: copyDomainJSON }, t("action.copyJSON")),
        h("button", { key: "s3", type: "button", onClick: copyDomainS3 }, t("action.copyS3")),
        h("button", { key: "import", type: "button", onClick: () => fileInputRef.current?.click() }, t("action.importCache")),
        h("input", { key: "file", ref: fileInputRef, type: "file", accept: "application/json", style: { display: "none" }, onChange: (event) => importDomainCache(event.target.files?.[0]) })
      ]),
      copyState ? h("p", { key: "copy-state", className: "small" }, copyState) : null,
      h("div", { key: "entities", className: "journey-steps" }, items.map((scenario, index) => {
        const key = domainEntityKey(scenario);
        const cached = draft.entities?.[key]?.id || scenario.observed?.cached_id || "";
        return h("button", {
          key,
          type: "button",
          className: "journey-step",
          "aria-pressed": selectedEntity === key,
          onClick: () => setSelectedKey(key)
        }, [
          h("span", { key: "n", className: "step-num" }, String(index + 1).padStart(2, "0")),
          h("span", { key: "label" }, domainStepVerb(key)),
          h("span", { key: "state", className: "pill " + (cached ? "passed" : scenario.status || "partial") }, cached ? t("status.cached") : statusText(scenario.status || "missing"))
        ]);
      }))
    ]),
    h("section", { key: "right", className: "panel stack" }, [
      h("div", { key: "head", className: "title-row" }, [
        h("div", { key: "title", className: "stack" }, [
          h("h2", { key: "h" }, domainStepVerb(selectedEntity)),
          h("p", { key: "p" }, t("domain.detailDesc"))
        ]),
        h(StatusPill, { key: "status", status: selected?.status })
      ]),
      h(KeyValue, { key: "create", label: t("label.create"), value: observed.create_endpoint }),
      h(KeyValue, { key: "verify", label: t("label.verify"), value: observed.verify_endpoint }),
      h(KeyValue, { key: "source", label: t("label.source"), value: observed.source }),
      h(KeyValue, { key: "figma", label: t("label.figmaEvidence"), value: figma ? h("a", { href: figmaURL }, figma.name + " / " + figma.node_id) : t("state.notLinked") }),
      h(KeyValue, { key: "evidence", label: t("label.evidenceRow"), value: selected?.id }),
      h("label", { key: "id", className: "field-row" }, [
        h("span", { key: "l", className: "muted" }, t("label.cachedID")),
        h("input", { key: "v", value: current.id || "", onChange: (event) => patchEntity({ id: event.target.value }) })
      ]),
      h("label", { key: "name", className: "field-row" }, [
        h("span", { key: "l", className: "muted" }, t("label.name")),
        h("input", { key: "v", value: current.name || "", onChange: (event) => patchEntity({ name: event.target.value }) })
      ]),
      h("div", { key: "buttons", className: "manual-actions" }, [
        h("button", { key: "observed", type: "button", onClick: useObservedID }, t("action.useObserved")),
        h("button", { key: "verified", type: "button", onClick: () => patchEntity({ source: "human-verified" }) }, t("action.markVerified"))
      ]),
      selected?.failure_summary ? h("p", { key: "failure" }, selected.failure_summary) : null,
      h("pre", { key: "json" }, JSON.stringify({ observed, cache_record: current }, null, 2))
    ])
  ]);
}

function ManualQAView({ draft, updateDraft, selectedID, setSelectedID }) {
  const [copyState, setCopyState] = useState("");
  const fileInputRef = useRef(null);
  const allScenarios = scenarios();
  const selected = allScenarios.find((scenario) => scenario.id === selectedID) || allScenarios[0];
  const selectedRecord = selected ? manualRecordFor(draft, selected) : null;
  const summary = manualStatusSummary(draft.records);
  const manualPayload = buildManualEvidence(draft);

  function patchDraft(patch) {
    updateDraft((current) => ({ ...current, ...patch }));
  }

  function setManualStatus(status) {
    if (!selected) return;
    updateDraft((current) => {
      const record = manualRecordFor(current, selected);
      return {
        ...current,
        records: {
          ...(current.records || {}),
          [selected.id]: {
            ...record,
            manual_status: status,
            tested_at: new Date().toISOString(),
            tester: current.tester || record.tester,
            environment: current.environment || record.environment
          }
        }
      };
    });
  }

  function setManualNote(note) {
    if (!selected) return;
    updateDraft((current) => {
      const record = manualRecordFor(current, selected);
      return {
        ...current,
        records: {
          ...(current.records || {}),
          [selected.id]: {
            ...record,
            note,
            tester: current.tester || record.tester,
            environment: current.environment || record.environment
          }
        }
      };
    });
  }

  async function copyManualJSON() {
    await navigator.clipboard.writeText(JSON.stringify(manualPayload, null, 2));
    setCopyState(t("manual.jsonCopied"));
  }

  async function copyUploadCommand() {
    await navigator.clipboard.writeText(s3ManualUploadCommand());
    setCopyState(t("manual.s3Copied"));
  }

  function openImportPicker() {
    if (fileInputRef.current) fileInputRef.current.click();
  }

  function importManualEvidence(file) {
    if (!file) return;
    file.text().then((text) => {
      const imported = JSON.parse(text);
      const records = {};
      for (const record of imported.scenarios || []) {
        if (record.id) records[record.id] = record;
      }
      updateDraft((current) => ({
        ...current,
        tester: imported.tester || current.tester,
        environment: imported.environment || current.environment,
        records
      }));
      setCopyState(t("manual.imported"));
    }).catch((err) => setCopyState(t("manual.importFailed", { message: err.message })));
  }

  function resetDraft() {
    const next = { tester: draft.tester || "", environment: draft.environment || "local", records: {} };
    updateDraft(() => next);
    setCopyState(t("manual.reset"));
  }

  if (!selected) return h("section", { className: "panel" }, h("div", { className: "empty" }, t("state.noManualScenarios")));

  return h("div", { className: "manual-grid" }, [
    h("section", { key: "left", className: "panel stack" }, [
      h("div", { key: "head", className: "title-row" }, [
        h("div", { key: "copy", className: "stack" }, [
          h("h2", { key: "h" }, t("manual.title")),
          h("p", { key: "p" }, t("manual.desc"))
        ]),
        h(StatusPill, { key: "status", status: manualPayload.status })
      ]),
      h("div", { key: "fields", className: "stack" }, [
        h("label", { key: "tester", className: "field-row" }, [
          h("span", { key: "l", className: "muted" }, t("label.tester")),
          h("input", { key: "v", value: draft.tester, onChange: (event) => patchDraft({ tester: event.target.value }), placeholder: t("manual.testerPlaceholder") })
        ]),
        h("label", { key: "env", className: "field-row" }, [
          h("span", { key: "l", className: "muted" }, t("label.environment")),
          h("select", { key: "v", value: draft.environment, onChange: (event) => patchDraft({ environment: event.target.value }) }, ["local", "staging"].map((value) =>
            h("option", { key: value, value }, value)
          ))
        ])
      ]),
      h("div", { key: "metrics", className: "metric-grid" }, [
        h("div", { key: "total", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.touched")), h("strong", { key: "v" }, String(summary.total))]),
        h("div", { key: "running", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.running")), h("strong", { key: "v" }, String(summary.running || 0))]),
        h("div", { key: "passed", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.passed")), h("strong", { key: "v" }, String(summary.passed || 0))]),
        h("div", { key: "failed", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.failed")), h("strong", { key: "v" }, String(summary.failed || 0))]),
        h("div", { key: "blocked", className: "metric" }, [h("span", { key: "l", className: "muted" }, t("metric.blocked")), h("strong", { key: "v" }, String(summary.blocked || 0))])
      ]),
      h("div", { key: "exports", className: "manual-actions" }, [
        h("button", { key: "download", type: "button", onClick: () => downloadManualEvidence(draft) }, t("action.downloadJSON")),
        h("button", { key: "copy", type: "button", onClick: copyManualJSON }, t("action.copyJSON")),
        h("button", { key: "s3", type: "button", onClick: copyUploadCommand }, t("action.copyS3")),
        h("button", { key: "import", type: "button", onClick: openImportPicker }, t("action.importJSON")),
        h("input", { key: "file", ref: fileInputRef, type: "file", accept: "application/json", style: { display: "none" }, onChange: (event) => importManualEvidence(event.target.files?.[0]) }),
        h("button", { key: "reset", type: "button", className: "danger", onClick: resetDraft }, t("action.reset"))
      ]),
      copyState ? h("p", { key: "copy-state", className: "small" }, copyState) : null,
      h("div", { key: "path", className: "stack" }, [
        h(KeyValue, { key: "save", label: t("label.savePath"), value: h("code", null, manualEvidencePath()) }),
        h(KeyValue, { key: "s3cmd", label: t("label.s3Handoff"), value: h("code", null, s3ManualUploadCommand()) })
      ]),
      h("div", { key: "scenario-list", className: "stack" }, allScenarios.map((scenario) => {
        const record = manualRecordFor(draft, scenario);
        return h("button", {
          key: scenario.id,
          type: "button",
          "aria-pressed": selected.id === scenario.id,
          onClick: () => setSelectedID(scenario.id)
        }, scenario.id + " - " + statusText(record.manual_status));
      }))
    ]),
    h("section", { key: "right", className: "panel stack" }, [
      h("div", { key: "head", className: "title-row" }, [
        h("div", { key: "title", className: "stack" }, [
          h("h2", { key: "id" }, selected.id),
          h("p", { key: "area" }, classifyScenario(selected.id))
        ]),
        h(StatusPill, { key: "status", status: selectedRecord.manual_status })
      ]),
      h("div", { key: "buttons", className: "manual-actions" }, [
        h("button", { key: "running", type: "button", "aria-pressed": selectedRecord.manual_status === "running", onClick: () => setManualStatus("running") }, t("action.start")),
        h("button", { key: "passed", type: "button", "aria-pressed": selectedRecord.manual_status === "passed", onClick: () => setManualStatus("passed") }, t("action.pass")),
        h("button", { key: "failed", type: "button", "aria-pressed": selectedRecord.manual_status === "failed", onClick: () => setManualStatus("failed") }, t("action.fail")),
        h("button", { key: "blocked", type: "button", "aria-pressed": selectedRecord.manual_status === "blocked", onClick: () => setManualStatus("blocked") }, t("action.block"))
      ]),
      h(KeyValue, { key: "source", label: t("label.sourceStatus"), value: statusText(selected.status) }),
      h(KeyValue, { key: "endpoint", label: t("label.endpoint"), value: [selected.method, selected.endpoint].filter(Boolean).join(" ") }),
      h(KeyValue, { key: "tested", label: t("label.testedAt"), value: selectedRecord.tested_at }),
      h("label", { key: "note", className: "stack" }, [
        h("span", { key: "label", className: "muted" }, t("label.manualNote")),
        h("textarea", { key: "textarea", value: selectedRecord.note, onChange: (event) => setManualNote(event.target.value), placeholder: t("manual.notePlaceholder") })
      ]),
      h("div", { key: "json", className: "stack" }, [
        h("h3", { key: "h" }, t("label.currentManualRecord")),
        h("pre", { key: "pre" }, JSON.stringify(selectedRecord, null, 2))
      ])
    ])
  ]);
}

function Toolbar({ query, setQuery, status, setStatus, area, setArea }) {
  return h("section", { className: "toolbar" }, [
    h("input", { key: "q", value: query, onChange: (event) => setQuery(event.target.value), placeholder: t("toolbar.search") }),
    h("select", { key: "status", value: status, onChange: (event) => setStatus(event.target.value) }, ["all", "passed", "partial", "skipped", "failed"].map((value) =>
      h("option", { key: value, value }, statusText(value))
    )),
    h("select", { key: "area", value: area, onChange: (event) => setArea(event.target.value) }, functionalAreas.map((item) =>
      h("option", { key: item.id, value: item.id }, t(item.titleKey))
    ))
  ]);
}

function App() {
  const [tab, setTab] = useState("areas");
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("all");
  const [area, setArea] = useState(functionalAreas[0].id);
  const [selectedID, setSelectedID] = useState(scenarios()[0]?.id || "");
  const [manualDraft, setManualDraft] = useState(readManualDraft);
  const [domainDraft, setDomainDraft] = useState(readDomainDraft);
  const selectedArea = useMemo(() => area, [area]);
  function updateManualDraft(nextOrUpdater) {
    setManualDraft((current) => {
      const next = typeof nextOrUpdater === "function" ? nextOrUpdater(current) : nextOrUpdater;
      persistManualDraft(next);
      return next;
    });
  }
  function updateDomainDraft(nextOrUpdater) {
    setDomainDraft((current) => {
      const next = typeof nextOrUpdater === "function" ? nextOrUpdater(current) : nextOrUpdater;
      persistDomainDraft(next);
      return next;
    });
  }
  return h("div", { className: "shell" }, [
    h(Summary, { key: "summary" }),
    h(Tabs, { key: "tabs", active: tab, onChange: setTab }),
    tab === "areas" || tab === "sequence"
      ? h(Toolbar, { key: "toolbar", query, setQuery, status, setStatus, area: selectedArea, setArea })
      : null,
    tab === "areas" ? h(AreasView, { key: "areas", activeArea: selectedArea, setActiveArea: setArea, query, status, selectedID, setSelectedID }) : null,
    tab === "sequence" ? h(SequenceView, { key: "sequence", query, status, selectedID, setSelectedID }) : null,
    tab === "domain" ? h(DomainJourneyView, { key: "domain", draft: domainDraft, updateDraft: updateDomainDraft }) : null,
    tab === "manual" ? h(ManualQAView, { key: "manual", draft: manualDraft, updateDraft: updateManualDraft, selectedID, setSelectedID }) : null,
    tab === "figma" ? h(FigmaView, { key: "figma" }) : null,
    tab === "loop" ? h(LoopView, { key: "loop" }) : null
  ]);
}

createRoot(document.getElementById("root")).render(h(App));
</script></body></html>`))
