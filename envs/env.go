package envs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}
}

func Get(key, def string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	// log
	log.Println(key, ": Valor default")

	return def
}
