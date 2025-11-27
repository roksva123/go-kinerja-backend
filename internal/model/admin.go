package model

type Admin struct {
    ID           string `json:"id"`
    Username     string `json:"username"`
    PasswordHash string `json:"password_hash"`
    CreatedAt    int16  `json:"createdat"`
}
