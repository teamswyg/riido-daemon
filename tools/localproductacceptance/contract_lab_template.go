package main

import "html/template"

var contractLabTemplate = template.Must(template.New("contract-lab").Parse(`<!doctype html>
<html lang="ko">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Riido AI Agent Contract Lab</title>
  <style>
    body { margin: 0; font-family: ui-sans-serif, system-ui, sans-serif; background: #f7f8fb; color: #18202f; }
    main { max-width: 1120px; margin: 0 auto; padding: 32px 20px 48px; }
    h1 { font-size: 28px; margin: 0 0 8px; }
    p { line-height: 1.55; color: #4a5568; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(290px, 1fr)); gap: 12px; }
    .card { background: white; border: 1px solid #dce2ea; border-radius: 8px; padding: 16px; }
    .passed { border-left: 4px solid #168a4a; }
    .failed { border-left: 4px solid #c2412d; }
    .skipped, .partial { border-left: 4px solid #b7791f; }
    code, pre { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
    pre { white-space: pre-wrap; background: #f1f4f8; padding: 10px; border-radius: 6px; }
    .pill { display: inline-block; padding: 2px 8px; border-radius: 999px; background: #e7edf5; font-size: 12px; }
  </style>
</head>
<body><main id="root"></main>
<script type="application/json" id="evidence">{{.Evidence}}</script>
<script type="module">
import React from "https://esm.sh/react@18.3.1";
import { createRoot } from "https://esm.sh/react-dom@18.3.1/client";
const evidence = JSON.parse(document.getElementById("evidence").textContent);
function Card({s}) {
  return React.createElement("section", {className: "card " + s.status}, [
    React.createElement("h2", {key: "h"}, s.id),
    React.createElement("span", {key: "p", className: "pill"}, s.status),
    s.method && React.createElement("p", {key: "m"}, s.method + " " + s.endpoint),
    s.failure_summary && React.createElement("p", {key: "f"}, s.failure_summary),
    s.repair && React.createElement("p", {key: "r"}, s.repair.summary),
    s.observed && React.createElement("pre", {key: "o"}, JSON.stringify(s.observed, null, 2)),
  ]);
}
function App() {
  return React.createElement(React.Fragment, null, [
    React.createElement("h1", {key: "h"}, "Riido AI Agent Contract Lab"),
    React.createElement("p", {key: "p"}, "Generated from real local acceptance evidence. Use these cards as the frontend handoff for API order, identifiers, and stream behavior."),
    React.createElement("div", {key: "g", className: "grid"}, evidence.scenarios.map(s => React.createElement(Card, {key: s.id + (s.endpoint || ""), s}))),
  ]);
}
createRoot(document.getElementById("root")).render(React.createElement(App));
</script></body></html>`))
