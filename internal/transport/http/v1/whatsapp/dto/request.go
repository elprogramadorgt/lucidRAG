package dto

// TODO Validate data types
type HookRequest struct {
	VerifyToken string `form:"hub.verify_token" binding:"required"`
	Mode        string `form:"hub.mode" binding:"required"`
	Challenge   string `form:"hub.challenge" binding:"required"`
}
