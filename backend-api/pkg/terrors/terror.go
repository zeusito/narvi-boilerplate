package terrors

import "github.com/labstack/echo/v5"

type Terror struct {
	ErrCode        string `json:"code"`
	HttpStatusCode int    `json:"-"`
	ErrMessage     string `json:"message"`
}

func (e *Terror) Error() string {
	return e.ErrMessage
}

func (e *Terror) ToMap() map[string]string {
	return map[string]string{
		"code":    e.ErrCode,
		"message": e.ErrMessage,
	}
}

func (e *Terror) ToEchoHttpError() error {
	return echo.NewHTTPError(e.HttpStatusCode, e.ErrMessage)
}
