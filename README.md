# Comparação de frameworks Golang

Este repositório tem como objetivo comparar o comportamento de diferentes frameworks web em Go em cenários realistas de API, com foco em desempenho, consumo de recursos, integração, consistência e recursos nativos.

## Objetivo
A ideia é avaliar como cada framework lida com:
- roteamento e binding de dados;
- serialização de respostas;
- concorrência e processamento assíncrono;
- integração com banco de dados, clientes HTTP e observabilidade;
- impacto em CPU, memória e latência sob carga.

## Frameworks avaliados
Os frameworks contemplados nos testes são Gin, Fiber, Echo, Chi e Beego.

### Gin
Gin é um framework leve e muito popular na comunidade Go, conhecido por seu roteamento rápido e por oferecer uma API simples para construir APIs REST. Ele usa um router próprio otimizado e é frequentemente escolhido para projetos que precisam de boa performance sem abrir mão de ergonomia.

### Fiber
Fiber foi criado para oferecer uma experiência semelhante à do Express no ecossistema Node.js, mas com a performance de Go. Sua proposta é reduzir a curva de aprendizado para quem vem de JavaScript e, ao mesmo tempo, entregar alto throughput em cenários de alta carga.

### Echo
Echo destaca-se por seu equilíbrio entre simplicidade, extensibilidade e recursos prontos para APIs de negócios. Ele é bastante usado em projetos que precisam de estrutura limpa, middlewares bem organizados e uma boa experiência de desenvolvimento.

### Chi
Chi é uma opção minimalista e composável, mantendo forte proximidade com a biblioteca padrão do Go. Ele é interessante para quem quer um router enxuto, sem muita abstração, e com boa integração com o ecossistema nativo.

### Beego
Beego é o framework mais “batteries-included” da lista, trazendo uma proposta mais completa e mais opinativa. Ele se encaixa melhor em projetos que querem uma solução mais estruturada, com recursos prontos para MVC, ORM e camada de infraestrutura.

## Arquitetura da API
A aplicação será construída com uma abordagem baseada em Clean Architecture, com separação entre:
- `core`: regras de negócio, contratos e tipos de domínio;
- `adapter`: adaptações de entrada/saída, handlers e validações;
- `infra`: implementações concretas para banco, observabilidade, roteamento e clientes externos.

O projeto também utiliza Docker para orquestrar PostgreSQL, OpenTelemetry, Prometheus/Grafana e JMeter, permitindo testes repetíveis e comparação mais confiável entre frameworks.

## Estratégia de testes
Os testes serão divididos em três níveis:
1. Testes unitários do domínio e dos casos de uso.
2. Testes de integração para validar handler, roteamento e serialização.
3. Benchmarks end-to-end com JMeter e observabilidade integrada.

### Desempenho
A análise de desempenho será feita com JMeter, considerando HTTP/1.1, HTTP/2 e, quando aplicável, gRPC. Os cenários incluirão cargas leves e cargas pesadas, com foco em latência, throughput e uso de recursos.

### Consumo de recursos
Serão coletadas métricas de CPU, memória, rede e vazão para comparar o impacto de cada framework em condições similares. O monitoramento será feito com Prometheus, Grafana e métricas expostas pela aplicação.

### Complexidade de integração
A comparação também levará em conta o esforço necessário para integrar o framework com middleware, validação, banco de dados, tracing e rotas. Isso ajuda a medir não só desempenho, mas também a produtividade do desenvolvimento.

### Consistência
Os testes devem validar se uma mesma entrada gera uma saída previsível e estável, sem corrupção de dados, sem erros inesperados e sem comportamento divergente entre as implementações.

### Recursos nativamente disponíveis
Cada framework será avaliado por recursos prontos que ele oferece, como roteamento, binding, validação, middlewares, suporte a templates, documentação e integração com observabilidade.

## Cenários de benchmark
Os cenários foram escolhidos para cobrir diferentes tipos de custo operacional e de concorrência.

### Básicos
- Health Check: devolver um sucesso visando que a aplicação está saudável
- Hello World: mede o custo mínimo de uma requisição simples.
- Conversor de Moedas: avalia binding, validação e serialização sem acesso a banco.
- Buscador de CEP: testa leitura de dados em cache e uso de parâmetros de rota.

### Complexos
- Ingresso Esgotado: verifica contenção, transações e comportamento sob alta concorrência.
- Pix Black Friday: simula fila assíncrona, processamento em background e resposta rápida ao cliente.

## Próximos passos
- Definir a estrutura inicial da API com base nas ADRs.
- Implementar os contratos de domínio e as portas de uso.
- Criar o primeiro endpoint funcional com um framework inicial.
- Montar a infraestrutura de observabilidade e benchmark com Docker.
- Rodar a primeira bateria de testes e consolidar os resultados.

