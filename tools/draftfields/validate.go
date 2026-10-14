package main

func validate(repo string, manifest Manifest) (
	[]problem,
	[]AllowedCheck,
	[]ForbiddenCheck,
) {
	allowedProblems, allowed := validateAllowed(repo, manifest.AllowedFields)
	forbiddenProblems, forbidden := validateForbidden(repo, manifest)
	problems := allowedProblems
	problems = append(problems, forbiddenProblems...)
	return problems, allowed, forbidden
}
