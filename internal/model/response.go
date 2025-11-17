package model


type ResponseApi struct {
    ApiMessage string `json:"api_message"`
	Data       interface{} `json:"data"`
	Count 	int         `json:"count"`
}