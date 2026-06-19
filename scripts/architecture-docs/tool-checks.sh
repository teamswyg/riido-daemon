#!/usr/bin/env bash
set -euo pipefail

tool_tests=(
  ./tools/figmaboundary
  ./tools/providervalidation
  ./tools/agentexecutionevidence
  ./tools/loopevidence
  ./tools/redactiondrift
  ./tools/providerintegrationevidence
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
  ./tools/runtimeupgrade
  ./tools/localdaemoncontract
  ./tools/saasassignment
  ./tools/validationevidence
  ./tools/policybundleevidence
  ./tools/toolusegateevidence
  ./tools/nativeconfigmcp
  ./tools/fullaccessharness
  ./tools/assignmentfsm
  ./tools/privacymetadata
)

doc_tools=(
  repoverification
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
  localdaemoncontract
  saasassignment
  validationevidence
  policybundleevidence
  toolusegateevidence
  nativeconfigmcp
  fullaccessharness
  assignmentfsm
  privacymetadata
)

go test "${tool_tests[@]}" -count=1
go run ./tools/loopevidence -check
go run ./tools/docmap -check

for tool in "${doc_tools[@]}"; do
  go run "./tools/$tool" -check-doc
done

go run ./tools/branchgate -check-doc -check-script
go run ./tools/compatibilitygate -check-doc
go run ./tools/redactiondrift
go run ./tools/providerintegrationevidence -check-doc
go run ./tools/runtimesecretevidence -check-doc
