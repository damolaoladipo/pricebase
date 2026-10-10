package handlers

import (
	"github.com/gin-gonic/gin"
)

// Register creates a new user account and sends a registration OTP.
func Register(ctx *gin.Context) {

	// var req dtos.RegisterUserDTO

	// // Parse request body
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	ctx.JSON(http.StatusBadRequest, dtos.Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: "Invalid request body",
	// 		Status:  http.StatusBadRequest,
	// 	})
	// 	return
	// }

	//	return utils.NewErrorResponse("Invalid request body", http.StatusBadRequest, []any{}, nil)

	// // Validate registration data
	// validation := AuthService.ValidateRegister(req)

	// if validation.Error {
	// 	c.JSON(validation.Code, Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: validation.Message,
	// 		Status:  validation.Code,
	// 	})
	// 	return
	// }

	// // Validate email
	// email := strings.ToLower(strings.TrimSpace(req.Email))

	// if !AuthService.CheckEmail(email) {
	// 	c.JSON(http.StatusBadRequest, Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: "A valid email is required",
	// 		Status:  http.StatusBadRequest,
	// 	})
	// 	return
	// }

	// // Check if user already exists
	// user, err := UserRepository.FindByEmail(email)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: "Failed to check user",
	// 		Status:  http.StatusInternalServerError,
	// 	})
	// 	return
	// }

	// if user != nil {
	// 	if user.UserType == UserTypeSuperAdmin {
	// 		c.JSON(http.StatusBadRequest, Response{
	// 			Error:   true,
	// 			Errors:  []any{},
	// 			Message: "Forbidden!, use another email",
	// 			Status:  http.StatusBadRequest,
	// 		})
	// 		return
	// 	}

	// 	c.JSON(http.StatusBadRequest, Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: "User already exist, use another email",
	// 		Status:  http.StatusBadRequest,
	// 	})
	// 	return
	// }

	// // Validate password
	// if !AuthService.CheckPassword(req.Password) {
	// 	c.JSON(http.StatusBadRequest, Response{
	// 		Error:  true,
	// 		Errors: []any{},
	// 		Message: "password must contain at least 8 characters, 1 lowercase letter, 1 uppercase letter, 1 special character and 1 number",
	// 		Status: http.StatusBadRequest,
	// 	})
	// 	return
	// }

	// // Create user
	// user, err = UserService.CreateUser(CreateUserInput{
	// 	FirstName:    req.FirstName,
	// 	LastName:     req.LastName,
	// 	Email:        email,
	// 	Password:     req.Password,
	// 	PasswordType: PasswordTypeUserGenerated,
	// 	UserType:     req.UserType,
	// })

	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: "user not created",
	// 		Status:  http.StatusInternalServerError,
	// 	})
	// 	return
	// }

	// if user == nil {
	// 	c.JSON(http.StatusNotFound, Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: "user not created",
	// 		Status:  http.StatusNotFound,
	// 	})
	// 	return
	// }

	// // Update user type
	// if err := AuthService.UpdateUserType(user, req.UserType); err != nil {
	// 	c.JSON(http.StatusInternalServerError, Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: "failed to update user type",
	// 		Status:  http.StatusInternalServerError,
	// 	})
	// 	return
	// }

	// // Generate OTP
	// otp, err := AuthService.GenerateOTPCode(user, OtpTypeRegister)

	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, Response{
	// 		Error:   true,
	// 		Errors:  []any{},
	// 		Message: "failed to generate OTP",
	// 		Status:  http.StatusInternalServerError,
	// 	})
	// 	return
	// }

	// // Send OTP
	// if otp != "" {
	// 	sendOTP := EmailService.SendOTPEmail(SendOTPEmailInput{
	// 		User:    user,
	// 		Code:    otp,
	// 		OTPType: OtpTypeRegister,
	// 	})

	// 	if sendOTP.Error {
	// 		c.JSON(sendOTP.Code, Response{
	// 			Error:   true,
	// 			Errors:  []any{},
	// 			Message: sendOTP.Message,
	// 			Status:  sendOTP.Code,
	// 		})
	// 		return
	// 	}
	// }

	// // Map response
	// responseData := AuthMapper.MapRegisteredUser(user)

	// c.JSON(http.StatusOK, Response{
	// 	Error:   false,
	// 	Errors:  []any{},
	// 	Data:    responseData,
	// 	Message: "OTP has been sent to your email!",
	// 	Status:  http.StatusOK,
	// })

}

// Login authenticates a user and returns a JWT token.
func Login(ctx *gin.Context) {

}

// VerifyOTP verifies a registration OTP.
func VerifyOTP(ctx *gin.Context) {

}

// ResendOTP resends a registration OTP.
func ResendOTP(ctx *gin.Context) {

}

// ActivateAccount activates a user account.
func ActivateAccount(ctx *gin.Context) {

}

// ForgotPassword sends a password reset OTP.
func ForgotPassword(ctx *gin.Context) {

}

// ResetPassword resets a user password.
func ResetPassword(ctx *gin.Context) {

}
