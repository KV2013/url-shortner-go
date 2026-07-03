package other

import (
	"log"
	"os"
)

func someFunc() {
	os.Exit(1)        // want "call to os.Exit outside main function"
	log.Fatal("test") // want "call to log.Fatal outside main function"
	log.Fatalf("test %d", 1) // want "call to log.Fatalf outside main function"
	log.Fatalln("test") // want "call to log.Fatalln outside main function"
	panic("test") // want "use of built-in panic"
}
