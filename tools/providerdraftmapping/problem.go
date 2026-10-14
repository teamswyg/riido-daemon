package main

type problem struct {
	msg string
}

func (p problem) Error() string {
	return p.msg
}

func problemError(problems []problem) error {
	if len(problems) == 0 {
		return nil
	}
	return problems[0]
}

func problemMessages(problems []problem) []string {
	out := make([]string, 0, len(problems))
	for _, p := range problems {
		out = append(out, p.msg)
	}
	return out
}
