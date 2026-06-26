package responses

type HelloOutput struct {
	Message string `json:"message" title:"Mensagem" description:"Mensagem de cumprimento" example:"Hello, World"`
}
