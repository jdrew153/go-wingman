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

