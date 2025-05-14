package model

var MethodNotAllowed = Resp{
	Code:    405,
	Msg:     "Method not allowed",
	Content: "",
}

var DocTypeNotAllowed = Resp{
	Code:    405,
	Msg:     "Document type not allowed",
	Content: "",
}

var InvalidRequestBody = Resp{
	Code:    400,
	Msg:     "Invalid request body",
	Content: "",
}

var InvalidAccessKey = Resp{
	Code:    401,
	Msg:     "Invalid access_key",
	Content: "",
}

var TransformError = Resp{
	Code:    500,
	Msg:     "Failed to transform",
	Content: "",
}

var InvalidUserAccessToken = Resp{
	Code:    403,
	Msg:     "Invalid user_access_token",
	Content: "",
}
