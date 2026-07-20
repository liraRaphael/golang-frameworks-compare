# Sugestões de Melhorias de Arquitetura e Performance

Com base na refatoração profunda dos contratos de `Request`, `Response` e `HttpParam` no Core, identificamos excelentes oportunidades para melhorar a performance de execução do benchmark, a robustez da tipagem e o isolamento de camadas do projeto.

---

## 1. Cache de Reflexão na Extração de Tag de Parâmetros (Alta Performance)
### Problema
Atualmente, as funções `extractParamsFromTag` e `ParamsValuesToHttpParamsType` executam análises reflexivas (`reflect.TypeOf`, `NumField`, `Field(i).Tag`) em **cada requisição recebida** por qualquer adaptador de framework. A reflexão em Go é computacionalmente cara e gera alocações de memória desnecessárias no hot-path de rotas sob estresse.

### Sugestão
Implementar um cache global thread-safe (`sync.Map`) indexado pelo `reflect.Type` do parâmetro para armazenar os índices dos campos mapeados por tipo de tag.

### Exemplo de Código Conceitual
```go
package domain

import (
	"reflect"
	"sync"
)

type CachedFields struct {
	Headers []int
	Query   []int
	Path    []int
	Cookies []int
}

var tagCache sync.Map // map[reflect.Type]*CachedFields

func ExtractParamsFromTagWithCache(typ reflect.Type) (headersFields, queryFields, pathFields, cookieFields []int) {
	if typ == nil {
		return
	}
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if val, ok := tagCache.Load(typ); ok {
		cf := val.(*CachedFields)
		return cf.Headers, cf.Query, cf.Path, cf.Cookies
	}

	// Executa reflexão apenas na primeira vez
	headersFields, queryFields, pathFields, cookieFields = ExtractParamsFromTag(typ)
	
	tagCache.Store(typ, &CachedFields{
		Headers: headersFields,
		Query:   queryFields,
		Path:    pathFields,
		Cookies: cookieFields,
	})
	return
}
```

---

## 2. Redução de Alocações em `MarshalJSON` de Respostas (Zero-Allocation)
### Problema
O método `MarshalJSON` implementado para o `Responser` genérico instancia uma `struct` anônima temporária a cada serialização de resposta:
```go
func (r Responser[BodyType, ParamType]) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct { ... }{ ... })
}
```
Isso gera alocações na Heap para cada requisição HTTP de sucesso, impactando diretamente o benchmark de latência e consumo de memória (GC pressure).

### Sugestão
Usar um pool de buffers (`sync.Pool`) ou realizar a serialização do corpo diretamente se não houver necessidade de envelopar os cabeçalhos HTTP dentro do corpo JSON, uma vez que cabeçalhos e status codes já são entregues nativamente via protocolo HTTP.

---

## 3. Extensão do Router Abstrato para Agrupamento de Rotas (`Route Groups`)
### Problema
A interface de roteamento abstrata (`ports.Router`) atual é plana e só permite registrar rotas individualmente. Frameworks modernos como Gin, Echo e Fiber têm excelente suporte e otimizações nativas para agrupamento de rotas e aplicação de middlewares escopados por grupo (ex: `/api/v1/hello`).

### Sugestão
Adicionar suporte a sub-roteadores ou grupos na interface `ports.Router`:
```go
type Router interface {
	AddRoute(method string, path string, handler Controller[any, any], meta any)
	Group(prefix string, middlewares ...Middleware) Router
	Use(middleware Middleware)
	Start(addr string) error
}
```
Isso permitirá comparar a eficiência dos algoritmos de roteamento de sub-árvores (`trie routers`) de cada framework Go sob múltiplos níveis de prefixo.

---

## 4. Tipagem Avançada de Erros com Mapeamento Automatizado de Status HTTP
### Problema
Atualmente, o mapeamento de exceções no Standard Handler é feito de forma manual baseando-se em comparação de strings (`RegisterByMessage("user not found", ...)`) ou tipos estáticos de erro. Isso exige muita configuração manual e duplicação de boilerplate no bootstrap.

### Sugestão
Definir uma interface `DomainError` em `core/domain/errors` com suporte a código de erro estruturado e status HTTP semântico:
```go
type DomainError interface {
	error
	ErrorCode() string
	HTTPStatus() int
}
```
O handler padrão pode então resolver dinamicamente e de forma genérica qualquer erro que satisfaça a interface `DomainError` de forma automática, economizando linhas de código e aumentando a confiabilidade.
