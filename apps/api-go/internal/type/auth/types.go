package authtype

import "feishu-pipeline/apps/api-go/internal/model"

type UserResponse struct {
	ID           string     `json:"id"`
	FeishuOpenID string     `json:"feishuOpenID,omitempty"`
	Name         string     `json:"name"`
	Email        string     `json:"email,omitempty"`
	Role         model.Role `json:"role"`
	Departments  []string   `json:"departments"`
}

type FeishuSSOLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type FeishuSSOConfigResponse struct {
	Enabled bool   `json:"enabled"`
	AppID   string `json:"appId,omitempty"`
}

type LoginResponse struct {
	User UserResponse `json:"user"`
}

func NewUserResponse(user model.User) UserResponse {
	return UserResponse{
		ID:           user.ID,
		FeishuOpenID: user.FeishuOpenID,
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		Departments:  user.Departments,
	}
}

func NewLoginResponse(user model.User) LoginResponse {
	return LoginResponse{
		User: NewUserResponse(user),
	}
}
