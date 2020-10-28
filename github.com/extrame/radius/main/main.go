package main

import (
	"github.com/jessta/radius"
	"log"
)

func main() {
	s := radius.NewServer(":1812", "sEcReT")
	s.RegisterService("127.0.0.1", &radius.PasswordService{})
	log.Println("waiting for packets...")
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
