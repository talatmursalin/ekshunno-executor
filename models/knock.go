package models

import "github.com/talatmursalin/ekshunno-executor/customenums"

type Submission struct {
	Memory   float32 `json:"memory"`
	Time     float32 `json:"time"`
	Input    string  `json:"input"`
	Lang     string  `json:"lang"`
	Compiler string  `json:"compiler"`
	Src      string  `json:"src"`
}

type Knock struct {
	SubmissionRoom string     `json:"submission_room"`
	Submission     Submission `json:"submission"`
}

func (k *Knock) Validate() error {
	if k.Submission.Memory < 64 || k.Submission.Memory > 512 {
		k.Submission.Memory = 256
	}
	if k.Submission.Time <= 0 || k.Submission.Time > 3 {
		k.Submission.Time = 3
	}
	if _, err := customenums.StringToLangId(k.Submission.Lang); err != nil {
		return err
	}
	// validate compiler
	return nil
}
