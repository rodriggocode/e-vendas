package security 

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) ([]byte, error) {
  return  bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func CheckPassword(passwordHash, passwordString string) error{
  return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(passwordString))
}