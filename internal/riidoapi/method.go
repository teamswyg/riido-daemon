package riidoapi

type Method string

const (
	MethodStatus     Method = "status"
	MethodTasks      Method = "tasks"
	MethodTransition Method = "transition"
	MethodEvidence   Method = "evidence"
	MethodValidate   Method = "validate"
	MethodReviewDemo Method = "review-demo"
)
