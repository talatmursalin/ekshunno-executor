package models

type Config struct {
	Rabbitmq struct {
		Host     string `rmq:"host"`
		Port     string `rmq:"port"`
		Username string `rmq:"username"`
		Password string `rmq:"password"`
		Queue    string `rmq:"queue"`
	}
	General struct {
		Concurrency int
	}
}
