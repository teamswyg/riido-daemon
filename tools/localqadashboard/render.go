package main

import (
	"bytes"
	"html/template"
)

func renderDashboard(view dashboardView) (string, error) {
	var b bytes.Buffer
	if err := dashboardTemplate.Execute(&b, view); err != nil {
		return "", err
	}
	return b.String(), nil
}

var dashboardTemplate = template.Must(template.New("dashboard").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Riido Local QA Evidence</title>
<style>
body{font-family:Arial,Helvetica,sans-serif;margin:32px;background:#f7f7f4;color:#181816}
main{max-width:1120px;margin:0 auto}
.summary{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px;margin:20px 0}
.card{background:white;border:1px solid #deded8;border-radius:8px;padding:14px}
.status{font-weight:700;text-transform:uppercase}
table{width:100%;border-collapse:collapse;background:white;border:1px solid #deded8}
th,td{padding:10px;border-bottom:1px solid #ededeb;text-align:left;vertical-align:top}
code{font-size:12px;word-break:break-word}
.shot{display:block;max-width:220px;max-height:160px;margin-top:8px;border:1px solid #deded8;border-radius:6px}
details{max-width:420px}
h2{margin-top:28px}
</style>
</head>
<body>
<main>
<h1>Riido Local QA Evidence</h1>
<section class="summary">
<div class="card"><div>Run</div><div class="status">{{.Run.Status}}</div></div>
<div class="card"><div>Coverage Status</div><div class="status">{{.Run.CoverageStatus}}</div></div>
<div class="card"><div>Provider Status</div><div class="status">{{.Evidence.Status}}</div></div>
<div class="card"><div>Observed</div><div>{{.ObservedAt}}</div></div>
<div class="card"><div>Expires</div><div>{{.ExpiresAt}}</div></div>
<div class="card"><div>Freshness</div><div id="freshness-status" class="status" data-expires="{{.ExpiresAt}}">unknown</div></div>
<div class="card"><div>Coverage</div><div>{{.CoverageSummary.Passed}}/{{.CoverageSummary.Total}} passed</div></div>
</section>
{{if .Run.OpenRepairs}}<h2>Open Repairs</h2><table><thead><tr><th>Provider</th><th>Class</th><th>Owner</th><th>Mode</th><th>Summary</th></tr></thead><tbody>{{range .Run.OpenRepairs}}<tr><td>{{.ProviderID}}</td><td>{{.Class}}</td><td>{{.Owner}}</td><td>{{.Mode}}</td><td>{{.Summary}}{{if .SuggestedCommand}}<br><code>{{.SuggestedCommand}}</code>{{end}}</td></tr>{{end}}</tbody></table>{{end}}
<h2>Coverage</h2>
<table><thead><tr><th>Scenario</th><th>Tier</th><th>Surface</th><th>Status</th><th>Repair</th></tr></thead><tbody>
{{range .CoverageRows}}<tr><td>{{.Title}}<br><code>{{.ID}}</code></td><td>{{.Tier}}</td><td>{{.Surface}}</td><td class="status">{{.Status}}</td><td>{{if .Repair.Class}}{{.Repair.Class}}<br>{{.Repair.Summary}}{{if .Repair.SuggestedCommand}}<br><code>{{.Repair.SuggestedCommand}}</code>{{end}}{{else}}{{.Detail}}{{end}}{{if .Screenshot}}<br><a href="{{.Screenshot}}">screenshot<img class="shot" src="{{.Screenshot}}" alt="{{.ID}} screenshot"></a>{{end}}</td></tr>{{end}}
</tbody></table>
<h2>Provider Evidence</h2>
<table>
<thead><tr><th>Provider</th><th>Available</th><th>Version</th><th>Integration</th><th>Evidence</th></tr></thead>
<tbody>
{{range .Evidence.Providers}}
<tr>
<td>{{.ID}}</td>
<td>{{.Available}}</td>
<td>{{.Version}}</td>
<td class="status">{{.IntegrationStatus}}</td>
<td><code>{{.IntegrationCommand}}</code>{{if .FailureSummary}}<details><summary>details</summary><pre>{{.FailureSummary}}</pre></details>{{end}}</td>
</tr>
{{end}}
</tbody>
</table>
</main>
<script>
(function(){var el=document.getElementById("freshness-status");if(!el){return;}var expires=Date.parse(el.dataset.expires);if(Number.isNaN(expires)){el.textContent="unknown";return;}var now=Date.now();el.textContent=now<expires?"fresh":"expired";el.title="Evaluated in browser at "+new Date(now).toISOString();})();
</script>
</body>
</html>`))
