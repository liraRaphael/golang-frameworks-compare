# ADR-011: Gestão de Configuração

**Status**: Aceito

**Data**: 2026-06-25

**Tags**: [config, env, infra]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)

## Contexto
Diferentes frameworks e integrações podem demandar configurações distintas, mas a aplicação precisa de uma fonte única de verdade para variáveis de ambiente. Sem um padrão claro, os benchmarks podem sofrer variações por diferenças invisíveis nas configurações carregadas em cada execução.

## Decisão
Utilizaremos a biblioteca `envconfig` para carregar configurações a partir de variáveis de ambiente. As configurações devem ser centralizadas na camada `infra/config` e injetadas durante o bootstrap da aplicação.

### Padrão adotado
- O projeto deve manter um arquivo `.env.example` na raiz, descrevendo todas as variáveis esperadas.
- A leitura de configuração deve ocorrer em um único ponto de entrada da aplicação.
- Os valores devem ser tipados e validados no início da inicialização, evitando falhas silenciosas durante a execução.
- Os componentes devem receber suas configurações por construtor, preservando o padrão de injeção de dependência.
- Cada trecho que necessitar de config, deverá ter seu arquivo com sua struct global inicializada com métodos proprios
  - Ex: `config/fiber.go`, `config/gin.go`, `config.client_grpc.go` e assim por seguinte
- Um arquivo `config.go` com a função `func init()` onde deverá ser carregado e registrado os valores

## Consequências
- O ambiente de benchmark passa a ser mais consistente e reproduzível entre os frameworks.
- A configuração fica explícita, documentada e fácil de trocar entre ambientes.
- O processo de setup da aplicação se torna mais simples para desenvolvedores e automações.

## Referências
1. https://github.com/kelseyhightower/envconfig
2. https://12factor.net/config
