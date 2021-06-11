package models

type Knock struct {
	SubmissionRoom string
	Submission     struct {
		Memory   float32
		Time     float32
		Input    string
		Lang     string
		Compiler string
		Src      string
	}
}
