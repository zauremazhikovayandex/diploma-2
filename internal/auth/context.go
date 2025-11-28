package auth

import "github.com/gin-gonic/gin"

const UserLoginKey = "user_login"

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
