package models

import "github.com/talatmursalin/ekshunno-executor/xcore/utils"

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

func (k *Knock) Validate() error {
	if k.Submission.Memory < 64 || k.Submission.Memory > 512 {
		k.Submission.Memory = 256
	}
	if k.Submission.Time <= 0 || k.Submission.Time > 3 {
		k.Submission.Time = 3
	}
	if _, err := utils.StringToLangId(k.Submission.Lang); err != nil {
		return err
	}
	// validate compiler
	return nil
}
