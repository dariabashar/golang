package entity

type CreateUserDTO struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type LoginUserDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type VerifyEmailDTO struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=4"`
}
