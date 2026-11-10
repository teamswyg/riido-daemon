package main

// commandHelp is the curated, person-facing help for one user command. The
// complete legacy usage (agent modes and every flag) stays in `riido help --all`.
type commandHelp struct {
	name     string
	summary  string
	usage    string
	notes    []string
	subs     []subHelp
	flags    []helpEntry
	examples []string
}

type subHelp struct {
	name    string
	summary string
	usage   string
	flags   []helpEntry
	example string
}

func (c commandHelp) sub(name string) *subHelp {
	for i := range c.subs {
		if c.subs[i].name == name {
			return &c.subs[i]
		}
	}
	return nil
}

var (
	jsonFlag      = helpEntry{"--json", "스크립트용 JSON 출력"}
	workFlag      = helpEntry{"--work <RIID-1234|ID>", "대상 작업"}
	workspaceFlag = helpEntry{"--workspace <ID>", "워크스페이스 (기본값: riido workspace use)"}
	teamFlag      = helpEntry{"--team-id <ID>", "팀 (작업 ID로 조회할 때 필요)"}
	pageFlags     = []helpEntry{{"--page <N>", "페이지 (기본 1)"}, {"--limit <N>", "한 번에 볼 개수 (기본 20, 최대 100)"}}
	keyNote       = "작업 키(RIID-1234)로 조회하려면 워크스페이스나 팀이 필요해요. riido workspace use <ID>로 기본값을 저장하면 생략할 수 있어요."
)

var commandHelpCatalog = map[string]commandHelp{
	"auth": {
		name: "auth", summary: "Riido 계정 로그인과 이 기기 연결을 관리합니다.",
		usage: "riido auth <명령> [옵션]",
		notes: []string{"로그인 정보는 macOS 키체인 같은 OS 보호 저장소에만 보관해요."},
		subs: []subHelp{
			{"login", "브라우저로 로그인합니다.", "riido auth login [--browser | --desktop]", []helpEntry{{"--browser", "브라우저에서 로그인 (기본)"}, {"--desktop", "Riido 앱에 로그인된 계정으로 로그인"}}, "riido auth login"},
			{"status", "로그인 상태를 확인합니다.", "riido auth status", nil, "riido auth status"},
			{"logout", "로그아웃하고 저장된 로그인을 지웁니다.", "riido auth logout", nil, "riido auth logout"},
			{"link", "이 CLI를 Riido 앱(이 기기)과 연결합니다.", "riido auth link [--restart] [--no-input]", []helpEntry{{"--restart", "만료된 미승인 요청을 새 승인 요청으로 교체"}, {"--no-input", "확인 창 없이 이전 승인만 이어받기"}}, "riido auth link"},
			{"repair", "보호 저장소를 복구합니다. 브라우저에서 다시 동의해요.", "riido auth repair [--restart] [--no-input]", []helpEntry{{"--restart", "끝난 요청 대신 새로 동의 시작"}, {"--no-input", "확인 창 없이 기존 영수증만 복구"}}, "riido auth repair"},
			{"transfer", "현재 기기로 연결을 전환하거나 진행 중인 전환을 복구합니다.", "riido auth transfer [--repair] [--restart] [--no-input]", []helpEntry{{"--repair", "기존 설치 키와 새 사용자 승인으로 전환 복구"}, {"--restart", "만료되거나 거절된 복구 승인 다시 요청"}, {"--no-input", "기존 요청 상태 확인"}}, "riido auth transfer"},
		},
		flags:    []helpEntry{{"--auth-origin <browser|desktop>", "이번 실행만 로그인 방식 지정"}, jsonFlag},
		examples: []string{"riido auth login", "riido auth status"},
	},
	"whoami": {
		name: "whoami", summary: "지금 로그인한 계정을 보여줍니다.", usage: "riido whoami [옵션]",
		flags: []helpEntry{jsonFlag}, examples: []string{"riido whoami"},
	},
	"update": {
		name: "update", summary: "CLI를 설치한 채널(정식·staging)의 최신 버전으로 업데이트합니다.", usage: "riido update [옵션]",
		notes:    []string{"설치 스크립트로 설치한 CLI만 업데이트해요. Riido 앱에 포함된 CLI는 앱을 업데이트하세요.", "riido는 정식, riido-staging은 staging 채널을 따라요."},
		flags:    []helpEntry{{"--check", "설치하지 않고 새 버전이 있는지만 확인"}, {"--channel <staging|latest>", "받을 채널 (기본: 설치한 채널)"}, jsonFlag},
		examples: []string{"riido update", "riido update --check"},
	},
	"doctor": {
		name: "doctor", summary: "설치·연결 상태를 진단합니다.", usage: "riido doctor [옵션]",
		flags:    []helpEntry{{"--check-api", "로그인과 API 호출까지 확인"}, jsonFlag},
		examples: []string{"riido doctor", "riido doctor --check-api"},
	},
	"work": {
		name: "work", summary: "작업과 댓글·활동·연관 작업을 조회합니다.", usage: "riido work <명령> <RIID-1234|ID> [옵션]",
		notes: []string{keyNote},
		subs: []subHelp{
			{"get", "작업 상세를 봅니다.", "riido work get <RIID-1234|ID>", nil, "riido work get RIID-1234"},
			{"search", "제목·키·본문으로 작업을 찾습니다.", "riido work search --query <TEXT>", []helpEntry{{"--query <TEXT>", "검색어"}, {"--limit <N>", "결과 개수"}, {"--cursor <C>", "다음 페이지"}}, "riido work search --query \"로그인\""},
			{"tree", "상위·하위 작업을 봅니다.", "riido work tree <RIID-1234|ID>", nil, "riido work tree RIID-1234"},
			{"relations", "연관 작업을 봅니다.", "riido work relations <RIID-1234|ID>", nil, "riido work relations RIID-1234"},
			{"comments", "댓글을 봅니다.", "riido work comments <RIID-1234|ID>", pageFlags, "riido work comments RIID-1234 --limit 50"},
			{"replies", "댓글의 답글을 봅니다.", "riido work replies <RIID-1234|ID> --comment-id <ID>", []helpEntry{{"--comment-id <ID>", "답글을 볼 댓글"}}, "riido work replies RIID-1234 --comment-id <COMMENT-ID>"},
			{"activity", "활동 기록을 봅니다.", "riido work activity <RIID-1234|ID>", pageFlags, "riido work activity RIID-1234"},
			{"body", "긴 본문을 나눠 읽습니다.", "riido work body <RIID-1234|ID> --version <VERSION>", []helpEntry{{"--version <VERSION>", "work get 결과의 본문 버전"}, {"--offset <N>", "읽기 시작 위치"}}, "riido work body RIID-1234 --version <VERSION>"},
		},
		flags:    []helpEntry{workspaceFlag, teamFlag, jsonFlag},
		examples: []string{"riido work get RIID-1234", "riido work search --query \"로그인\"", "riido work comments RIID-1234"},
	},
	"context": {
		name: "context", summary: "작업과 직접 연결된 맥락(작업·미팅·관계)을 한 번에 가져옵니다.", usage: "riido context get --work <RIID-1234|ID> [옵션]",
		notes: []string{keyNote, "part와 changes는 Agent 실행 전용이에요."},
		subs: []subHelp{
			{"get", "작업 맥락 묶음을 가져옵니다.", "riido context get --work <RIID-1234|ID>", nil, "riido context get --work RIID-1234"},
			{"part", "(Agent 실행 전용) 수집한 자료의 다음 부분", "riido context part <SOURCE-ID> --cursor <CURSOR>", nil, ""},
			{"changes", "(Agent 실행 전용) 이전 수집 이후 바뀐 점", "riido context changes --since <SNAPSHOT>", nil, ""},
		},
		flags:    []helpEntry{workspaceFlag, teamFlag, jsonFlag},
		examples: []string{"riido context get --work RIID-1234"},
	},
	"meeting": {
		name: "meeting", summary: "작업에 연결된 미팅과 본문·보드·댓글을 조회합니다.", usage: "riido meeting <명령> [옵션]",
		subs: []subHelp{
			{"list", "작업에 연결된 미팅을 봅니다 (최대 5개).", "riido meeting list --work <RIID-1234|ID>", []helpEntry{workFlag}, "riido meeting list --work RIID-1234"},
			{"get", "미팅 상세·본문·보드를 봅니다.", "riido meeting get <MEETING-ID> --team-id <ID>", []helpEntry{{"--limit <N>", "보드 카드 개수 (기본 100, 최대 200)"}, {"--cursor <C>", "다음 보드 페이지"}}, "riido meeting get <MEETING-ID> --team-id <TEAM-ID>"},
			{"comments", "미팅 댓글을 봅니다.", "riido meeting comments <MEETING-ID> --team-id <ID>", pageFlags, "riido meeting comments <MEETING-ID> --team-id <TEAM-ID>"},
			{"body", "긴 본문을 나눠 읽습니다.", "riido meeting body <MEETING-ID> --team-id <ID> --version <VERSION>", []helpEntry{{"--version <VERSION>", "meeting get 결과의 본문 버전"}}, ""},
		},
		flags:    []helpEntry{workspaceFlag, teamFlag, jsonFlag},
		examples: []string{"riido meeting list --work RIID-1234"},
	},
	"agents": {
		name: "agents", summary: "작업에서 볼 수 있는 Agent를 조회합니다.", usage: "riido agents <명령> --work <RIID-1234|ID> [옵션]",
		notes: []string{"연결 상태는 멘션·실행 허가를 뜻하지 않아요."},
		subs: []subHelp{
			{"list", "Agent 목록을 봅니다.", "riido agents list --work <RIID-1234|ID>", []helpEntry{{"--cursor <C>", "다음 페이지"}}, "riido agents list --work RIID-1234"},
			{"show", "Agent 상세를 봅니다.", "riido agents show <AGENT-ID> --work <RIID-1234|ID>", nil, "riido agents show <AGENT-ID> --work RIID-1234"},
		},
		flags:    []helpEntry{workFlag, workspaceFlag, teamFlag, jsonFlag},
		examples: []string{"riido agents list --work RIID-1234"},
	},
	"thread": {
		name: "thread", summary: "작업의 Agent 대화를 조회합니다.", usage: "riido thread <명령> --work <RIID-1234|ID> [옵션]",
		subs: []subHelp{
			{"list", "대화 목록을 봅니다.", "riido thread list --work <RIID-1234|ID>", nil, "riido thread list --work RIID-1234"},
			{"show", "대화 내용을 봅니다.", "riido thread show <THREAD-ID> --work <RIID-1234|ID>", nil, "riido thread show <THREAD-ID> --work RIID-1234"},
		},
		flags:    []helpEntry{workFlag, workspaceFlag, teamFlag, jsonFlag},
		examples: []string{"riido thread list --work RIID-1234"},
	},
	"resource": {
		name: "resource", summary: "작업에 연결된 링크·첨부·Figma를 봅니다.", usage: "riido resource list --work <RIID-1234|ID> [옵션]",
		subs:     []subHelp{{"list", "연결된 자료 목록을 봅니다.", "riido resource list --work <RIID-1234|ID>", nil, "riido resource list --work RIID-1234"}},
		flags:    []helpEntry{workFlag, workspaceFlag, teamFlag, jsonFlag},
		examples: []string{"riido resource list --work RIID-1234"},
	},
	"repo": {
		name: "repo", summary: "작업에 연결된 저장소·PR을 봅니다.", usage: "riido repo list --work <RIID-1234|ID> [옵션]",
		subs:     []subHelp{{"list", "연결된 저장소·PR 목록을 봅니다.", "riido repo list --work <RIID-1234|ID>", nil, "riido repo list --work RIID-1234"}},
		flags:    []helpEntry{workFlag, workspaceFlag, teamFlag, jsonFlag},
		examples: []string{"riido repo list --work RIID-1234"},
	},
	"workspace": {
		name: "workspace", summary: "참여 중인 워크스페이스를 보고 기본값을 정합니다.", usage: "riido workspace <명령> [옵션]",
		subs: []subHelp{
			{"list", "참여 중인 워크스페이스를 봅니다.", "riido workspace list", []helpEntry{{"--limit <N>", "한 번에 볼 개수 (기본 50, 최대 100)"}, {"--cursor <C>", "다음 페이지"}}, "riido workspace list"},
			{"use", "기본 워크스페이스로 저장합니다. 지금 멤버인지 확인한 뒤 저장해요.", "riido workspace use <WORKSPACE-ID>", nil, "riido workspace use <WORKSPACE-ID>"},
		},
		flags:    []helpEntry{{"--profile <NAME>", "저장할 프로필"}, jsonFlag},
		examples: []string{"riido workspace list", "riido workspace use <WORKSPACE-ID>"},
	},
	"config": {
		name: "config", summary: "이 기기의 CLI 기본 설정을 보고 바꿉니다.", usage: "riido config <명령> [옵션]",
		notes: []string{"키: profile, api_origin, auth_origin, workspace, team, format, timeout · 빈 값을 넣으면 설정이 지워져요. 비밀값은 저장하지 않아요."},
		subs: []subHelp{
			{"list", "현재 설정과 출처를 봅니다.", "riido config list", nil, "riido config list"},
			{"get", "설정 하나를 봅니다.", "riido config get <KEY>", nil, "riido config get workspace"},
			{"set", "설정을 저장합니다.", "riido config set <KEY> <VALUE>", nil, "riido config set profile staging"},
		},
		flags:    []helpEntry{{"--profile <NAME>", "대상 프로필 (local | staging | production)"}, jsonFlag},
		examples: []string{"riido config list", "riido config set profile staging"},
	},
	"completion": {
		name: "completion", summary: "셸 자동완성 스크립트를 출력합니다.", usage: "riido completion <bash|zsh|fish|powershell>",
		notes:    []string{"출력을 파일로 저장한 뒤 셸 설정에서 불러오세요 (zsh는 compinit 다음). 설정 파일을 자동으로 바꾸지 않아요."},
		examples: []string{"riido completion zsh > ~/.riido/completion.zsh", "echo 'source ~/.riido/completion.zsh' >> ~/.zshrc"},
	},
	"capabilities": {
		name: "capabilities", summary: "이 CLI가 지원하는 명령을 보여줍니다.", usage: "riido capabilities [옵션]",
		flags: []helpEntry{jsonFlag}, examples: []string{"riido capabilities"},
	},
	"version": {
		name: "version", summary: "CLI 버전을 보여줍니다.", usage: "riido version",
		examples: []string{"riido version"},
	},
}
