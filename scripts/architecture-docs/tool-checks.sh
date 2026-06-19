#!/usr/bin/env bash
set -euo pipefail

tool_tests=(
  ./tools/figmaboundary
  ./tools/contextmapdocs
  ./tools/figmaboundarydocs
  ./tools/providervalidation
  ./tools/agentexecutionevidence
  ./tools/loopevidence
  ./tools/lockingdocs
  ./tools/securityredactiondocs
  ./tools/securityinvariantsdocs
  ./tools/providerruntimeresponsibilitydocs
  ./tools/providerintegrationgatedocs
  ./tools/providerruntimeboundarydocs
  ./tools/redactiondrift
  ./tools/providerintegrationevidence
  ./tools/providermigrationdocs
  ./tools/runtimesecretevidence
  ./tools/docmap
  ./tools/repoverification
  ./tools/semanticeventactivity
  ./tools/eventauthority
  ./tools/providerdraftmapping
  ./tools/terminalresultmapping
  ./tools/shutdownauthority
  ./tools/approvaltimeout
  ./tools/processlifecycle
  ./tools/draftfields
  ./tools/sessionlifecycle
  ./tools/branchgate
  ./tools/compatibilitygate
  ./tools/clisurface
  ./tools/configreference
  ./tools/executablesearchpath
  ./tools/knowledgecoverage
  ./tools/releaseartifacts
  ./tools/integrationmatrix
  ./tools/moduledecomposition
  ./tools/storedistributiondocs
  ./tools/agentexecutiondesign
  ./tools/runtimeupgrade
  ./tools/localdaemoncontract
  ./tools/runtimeeligibility
  ./tools/taskrequirements
  ./tools/saasassignment
  ./tools/validationevidence
  ./tools/policybundleevidence
  ./tools/toolusegateevidence
  ./tools/nativeconfigmcp
  ./tools/fullaccessharness
  ./tools/assignmentfsm
  ./tools/privacymetadata
  ./tools/unsafebypassevidence
)

doc_tools=(
  repoverification
  contextmapdocs
  lockingdocs
  securityredactiondocs
  securityinvariantsdocs
  providerruntimeresponsibilitydocs
  providerintegrationgatedocs
  providerruntimeboundarydocs
  semanticeventactivity
  eventauthority
  providerdraftmapping
  terminalresultmapping
  shutdownauthority
  approvaltimeout
  processlifecycle
  draftfields
  sessionlifecycle
  runtimeupgrade
  clisurface
  configreference
  executablesearchpath
  knowledgecoverage
  releaseartifacts
  integrationmatrix
  moduledecomposition
  storedistributiondocs
  agentexecutiondesign
  figmaboundarydocs
  providermigrationdocs
  localdaemoncontract
  runtimeeligibility
  taskrequirements
  saasassignment
  validationevidence
  policybundleevidence
  toolusegateevidence
  nativeconfigmcp
  fullaccessharness
  assignmentfsm
  privacymetadata
  unsafebypassevidence
)

go test "${tool_tests[@]}" -count=1
go run ./tools/loopevidence -check -evidence-out /tmp/loop-engineering-evidence.json
go run ./tools/docmap -check
go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo . -check-policy-table

for tool in "${doc_tools[@]}"; do
  go run "./tools/$tool" -check-doc
done

go run ./tools/knowledgecoverage -manifest docs/executable-knowledge.riido.json -check-doc
go run ./tools/branchgate -check-doc -check-script
go run ./tools/compatibilitygate -check-doc
go run ./tools/redactiondrift
go run ./tools/providerintegrationevidence -check-doc
go run ./tools/runtimesecretevidence -check-doc
