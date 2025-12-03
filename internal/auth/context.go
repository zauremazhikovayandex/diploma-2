package auth

import "github.com/gin-gonic/gin"

// UserLoginKey — ключ, под которым логин пользователя хранится в gin.Context.
const UserLoginKey = "user_login"

// GetLoginFromCtx извлекает логин пользователя из gin.Context, если он там есть.
func GetLoginFromCtx(c *gin.Context) (string, bool) {
	v, ok := c.Get(UserLoginKey)
	if !ok {
		return "", false
	}
	login, ok := v.(string)
	if !ok || login == "" {
		return "", false
	}
	return login, true
}
