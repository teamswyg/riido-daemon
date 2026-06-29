package main

//go:generate go run ../qadslsync -spec ../../docs/30-architecture/figma-ai-agent-daemon-boundary/feature-ui.dsl.json -out feature_ui.generated.json
//go:generate go run ../qadslsync -spec ../../docs/30-architecture/qa-i18n.dsl.json -out qa_i18n.generated.json
//go:generate go run ../qadslsync -spec ../../docs/30-architecture/qa-system.dsl.json -out qa_system.generated.json
//go:generate go run ../qadslsync -spec ../../docs/30-architecture/domain-fixture-journey.dsl.json -out domain_fixture_journey.generated.json
//go:generate go run ../qadslsync -spec ../../docs/30-architecture/local-qa-daily-trigger.dsl.json -out local_qa_daily_trigger.generated.json
//go:generate go run ../qadslsync -spec ../../docs/30-architecture/local-acceptance-coverage.riido.json -out local_acceptance_coverage.generated.json
//go:generate go run ../qadslsync -spec ../../docs/30-architecture/closed-loop-maturity.dsl.json -out closed_loop_maturity.generated.json
