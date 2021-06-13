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

func (k *Knock) Validate() {
	if k.Submission.Memory < 256 || k.Submission.Memory > 512 {
		k.Submission.Memory = 512
	}
	if k.Submission.Time <= 0 || k.Submission.Time > 10 {
		k.Submission.Time = 10
	}
}
