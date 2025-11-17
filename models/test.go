package models

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"email"`
	Age   int    `validate:"gt=0,lt=150"`
}
