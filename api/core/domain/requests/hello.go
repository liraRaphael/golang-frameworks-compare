package requests

type HelloRequest struct {
	Name string `json:"name,omitempty" title:"Nome" description:"Nome para o cumprimento" example:"Ada"`
}
