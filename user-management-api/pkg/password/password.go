package password

import "golang.org/x/crypto/bcrypt"

// cost=10: ช้าพอกัน brute force เร็วพอไม่กระทบ UX (~100ms)
const cost = 10

func Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Compare(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
