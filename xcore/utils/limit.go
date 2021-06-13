package utils

type Limit struct {
	TimeLimit   float32
	MemoryLimit float32
	OutputLimit float32
}

func NewLimit(timeLimit, memoryLimit, outputLimit float32) *Limit {
	return &Limit{
		TimeLimit:   timeLimit,
		MemoryLimit: memoryLimit,
		OutputLimit: outputLimit,
	}
}
