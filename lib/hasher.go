package lib

import "github.com/matthewhartstonge/argon2"



func Hash(password string) (string, error) {
	argon := argon2.DefaultConfig()

	encoded, err := argon.HashEncoded([]byte(password))

	if err != nil {
		panic(err)
	}

	return string(encoded), nil
}


func VerifyHash(hashedPass string, plainPass string) (bool, error) {
	
	ok, err := argon2.VerifyEncoded([]byte(plainPass), []byte(hashedPass))

	if err != nil {
		return ok, err
	}

	return ok, nil
}

