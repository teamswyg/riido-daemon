package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

const (
	integrationArtifactName = "riido-cursor-side-effect.txt"
	integrationArtifactBody = "RIIDO_CURSOR_FILESYSTEM_SIDE_EFFECT_OK"
)

func integrationStartRequest(workdir string) agentbridge.StartRequest {
	return agentbridge.StartRequest{
		Prompt: `In the current working directory, create a file named ` + integrationArtifactName + ` with exactly this content and no trailing commentary in the file: ` + integrationArtifactBody + `

After the file is written, respond with exactly "ok".`,
		Cwd: workdir,
	}
}
