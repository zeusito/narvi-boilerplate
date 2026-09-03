package terrors

import "net/http"

func PreconditionFailed(message string) *Terror {
	return &Terror{
		ErrCode:        "PreconditionFailed",
		ErrMessage:     message,
		HttpStatusCode: http.StatusBadRequest,
	}
}

func Forbidden(message string) *Terror {
	return &Terror{
		ErrCode:        "ActionForbidden",
		ErrMessage:     message,
		HttpStatusCode: http.StatusForbidden,
	}
}

func RecordNotFound(message string) *Terror {
	return &Terror{
		ErrCode:        "RecordNotFound",
		ErrMessage:     message,
		HttpStatusCode: http.StatusBadRequest,
	}
}

func RecordAlreadyExists(message string) *Terror {
	return &Terror{
		ErrCode:        "RecordAlreadyExists",
		ErrMessage:     message,
		HttpStatusCode: http.StatusConflict,
	}
}

func Unknown(message string) *Terror {
	return &Terror{
		ErrCode:        "UnknownError",
		ErrMessage:     message,
		HttpStatusCode: http.StatusInternalServerError,
	}
}

func UnsupportedMediaType(message string) *Terror {
	return &Terror{
		ErrCode:        "UnsupportedContentType",
		ErrMessage:     message,
		HttpStatusCode: http.StatusUnsupportedMediaType,
	}
}

func UnAuthorized(message string) *Terror {
	return &Terror{
		ErrCode:        "UnAuthorized",
		ErrMessage:     message,
		HttpStatusCode: http.StatusUnauthorized,
	}
}

func OperationFailed(message string) *Terror {
	return &Terror{
		ErrCode:        "OperationFailed",
		ErrMessage:     message,
		HttpStatusCode: http.StatusInternalServerError,
	}
}

func PrincipalLocked(message string) *Terror {
	return &Terror{
		ErrCode:        "PrincipalLocked",
		ErrMessage:     message,
		HttpStatusCode: http.StatusUnauthorized,
	}
}
