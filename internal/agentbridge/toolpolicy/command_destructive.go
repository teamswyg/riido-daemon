package toolpolicy

import "strings"

func commandIsDestructive(command string) bool {
	normalized := strings.ToLower(strings.TrimSpace(command))
	for _, marker := range destructiveCommandMarkers() {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func destructiveCommandMarkers() []string {
	return []string{
		"rm -rf", "rm -fr", "sudo ", "chmod 777", "chown ",
		"dd if=", "dd of=", "mkfs", "git reset --hard", "git clean -fd",
		"git push", "terraform apply", "terraform destroy", "kubectl delete",
		"aws cloudformation delete", "aws dynamodb delete", "aws ecr delete",
		"aws iam delete", "aws s3 rm", "aws secretsmanager delete",
	}
}
