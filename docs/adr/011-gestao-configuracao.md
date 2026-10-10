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
`infra/config.AppConfig` centraliza as configurações carregadas com `envconfig`. O pacote processa a configuração em `init`; `main.go` obtém uma cópia por `NewAppConfig`.

### Padrão adotado
- `envconfig.Process("APP", &appConfig)` usa o prefixo `APP_`; o erro do processamento é atualmente ignorado.
- As variáveis efetivas usam o prefixo `APP_`: `APP_PORT` (8080), `APP_PPROF_PORT` (8081), `APP_SWAGGER_PORT` (8082), `APP_OTEL_EXPORTER_OTLP_ENDPOINT` (`localhost:4318`), `APP_SERVICE_NAME` (`go-framework-bench`) e `APP_FRAMEWORK` (`nethttp`).
- A API não carrega arquivos `.env` nem declara `config` por framework/cliente. A configuração é lida globalmente no pacote e consumida do bootstrap.

## Consequências
- A centralização e os defaults atuais facilitam configurar a inicialização; como o erro do carregamento é descartado, a validação de configuração ainda não é garantida.
- A configuração fica explícita, documentada e fácil de trocar entre ambientes.
- O processo de setup da aplicação se torna mais simples para desenvolvedores e automações.

## Referências
1. https://github.com/kelseyhightower/envconfig
2. https://12factor.net/config
