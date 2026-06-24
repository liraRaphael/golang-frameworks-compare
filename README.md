# Comparação de frameworks Golang
Sabendo que o *Golang* é uma ferramenta conhecida pela sua performace, uma das coisas para se decidir no momento do desenvolvimento é qual framework escolher para o servidor *HTTP* que rodará por trás.
O intuito aqui é comparar os principais *frameworks* *Go* em questão de **performace**, **consumo de recurso**, **complexidade de integração**, **consistência** e **recursos nativamente disponíveis**.

## Frameworks
Os frameworks usados para os testes são *Gin*, *Fiber*, *Echo*, *Chi*, and *Beego*.

### Gin
```ToDo
faça uma intro falando do framework, como foi implementado (no caso fasthttp ou net/http) e dados como:
Best for: REST APIs, microservices, and general backend application development.Core Advantage: It is the most popular framework in the ecosystem, capturing nearly 48% of Go developers. It utilizes a custom version of HttpRouter for lightning-fast routing and boasts minimal memory allocations.Key Features: Built-in JSON validation, easy routing groups, and robust crash-free middleware management.
deixe em portugues e textual
```

### Fiber
```ToDo
faça uma intro falando do framework, como foi implementado (no caso fasthttp ou net/http) e dados como:
Best for: Node.js developers transitioning to Go, and extreme high-throughput tasks.Core Advantage: It is explicitly designed to mimic the minimalist API design of Express.js (JavaScript), easing the learning curve for frontend and full-stack engineers.Key Features: Built on top of the ultra-fast fasthttp engine rather than Go's native net/http, making it exceptionally quick in synthetic benchmarks.
deixe em portugues e textual
```

### Echo
```ToDo
faça uma intro falando do framework, como foi implementado (no caso fasthttp ou net/http) e dados como:
Best for: Clean architecture, scalable REST endpoints, and complex enterprise structures.Core Advantage: Highly extensible and reliable, Echo strikes a middle ground by being lightweight while including excellent out-of-the-box data utilities.Key Features: Highly optimized HTTP router, robust data binding, centralized error handling, and native support for various template engines.
deixe em portugues e textual
```

### Chi
```ToDo
faça uma intro falando do framework, como foi implementado (no caso fasthttp ou net/http) e dados como:
Best for: Developers who want to stay as close to the native Go ecosystem as possible.Core Advantage: It is a lightweight, composable router completely compatible with the standard net/http library. You can easily pull native Go code straight into Chi without rewriting context variables.Key Features: Sub-routing structures, zero external dependencies, and context control designed for large, long-term codebases.
deixe em portugues e textual
```

### Beego
```ToDo
faça uma intro falando do framework, como foi implementado (no caso fasthttp ou net/http) e dados como:
Best for: Large corporate applications requiring a full-stack, "batteries-included" toolkit.Core Advantage: Unlike the micro-routers above, Beego is a heavyweight Model-View-Controller (MVC) framework modeled after traditional tools like Rails or Django.Key Features: Built-in Object-Relational Mapper (ORM), automated project generation via its CLI tool, integrated cache handling, and session logging systems.
deixe em portugues e textual
```

## A API
A *API* será desenvolvida com os *frameworks* aqui listado com **Clean Arch**, além do uso de *Postsgres*, *Otel*, *Grafana*, *JMeter* rodando sobre *Docker*.
```ToDo
falar mais sobre o que e como implementado
```

## Testes
```ToDo
Breve resumo do que será testado
```

### Performace
O teste de performace será realizado com o *JMter* e além do tradicional protocolo *Http/1.1*, também com o *Http/2* e *gRPC*, bem como os *mocks* nesses protocolos desde exemplos mais básicos como um simples **Hello, World**, até execuções complexas envolvendo *Goroutines*.
```ToDo
Colocar mais informações
```

### Consumo de recursos
```ToDo
Colocar mais informações; e dê ideias de como monitorar o uso de memória ram, cpu, rede e etc
```

### Complexidade de integração
```ToDo
Colocar mais informações; falar de cada framework, o que vem nativamente e o que irá facilitar ou dificultar a integração
```

### Consistência
```ToDo
Colocar mais informações; teste que dado uma entrada terá saida esperada sem corromper; de ideia de como fazer esse teste
```

### Recursos nativamente disponíveis
```ToDo
Colocar mais informações; observar o que cada framework tem e o que facilitará na hora de integrar
```

## Problemas
Para termos alguns cenários de testes, iremos desde os mais básicos até aos mais complexos como:

### Basicos
```ToDo
Falar sobre
```

#### Hello, World
```ToDo
Falar sobre Hello World
```


#### Conversor de Moedas

Este cenário foca no custo que o framework cobra para fazer o *binding* (mapeamento) de dados enviados pelo usuário e a serialização do JSON de resposta, sem tocar no banco de dados.

* **O Cenário:** Um endpoint POST `/v1/converte-moeda` recebe um JSON contendo um valor em Reais e a moeda de destino (ex: Dólar, Euro).
* **O Fluxo na API (Clean Arch):**
  1. O Handler faz o *parse* do JSON de entrada (ex: `{"valor": 150.00, "moeda_destino": "USD"}`).
  2. O UseCase aplica uma fórmula matemática simples baseada em uma taxa fixa (ex: divide o valor por 5.0).
  3. O Handler devolve o JSON calculado.


* **O que o Wiremock faz?** Pode ser usado para simular a busca da taxa de câmbio do dia, **mas sem delay** (resposta instantânea 200 OK com taxa fixa no JSON), apenas para testar o client HTTP do Go dentro do framework.
* **O banco de dados (Postgres):** Não é utilizado.
* **O que o Benchmark vai avaliar:** Alocação de memória por request. Fazer o parse de strings e structs gera lixo para o Garbage Collector. Como o banco de dados não está travando nada, o Grafana vai mostrar qual framework gerencia melhor a memória do Heap quando precisa processar milhares de JSONs por segundo.

---

#### O Buscador de CEP
Este cenário simula uma leitura de dados estáticos ou tabelas de referência (como países, estados ou categorias), onde o dado quase nunca muda e o banco responde imediatamente.

* **O Cenário:** Uma requisição GET `/v1/localidades/{id}` busca o nome de uma cidade baseada no ID.
* **O Fluxo na API (Clean Arch):**
1. O Handler extrai o parâmetro da URL (`id`).
2. O UseCase verifica se o ID já existe em um mapa em memória do Go (`map[int]string` protegido por um `sync.RWMutex` simples).
3. Se não estiver no mapa (primeira vez), faz um `SELECT nome FROM cidades WHERE id = $1` no Postgres, salva no mapa e retorna.
4. Nas requisições seguintes do JMeter, ele lê direto do mapa em memória.


* **O que o Benchmark vai avaliar:** A eficiência do **Roteador (Router)** de cada framework para extrair parâmetros de rota (`/localidades/:id`). Alguns frameworks usam algoritmos de busca em árvore (como o Radix Tree do Gin/Echo) que resolvem isso gastando quase zero de CPU, enquanto frameworks mais pesados podem ter um custo maior por requisição.


### Complexos
```ToDo
Falar sobre
```

#### O Problema do "Ingresso Esgotado"
Para garantir que o cenário do **"Ingresso Esgotado"** esteja no mesmo nível técnico, estrutural e analítico do cenário do Pix Black Friday, vamos detalhar minuciosamente a mecânica de concorrência, o comportamento do banco de dados, os pontos de estresse dos frameworks e o que medir especificamente na infraestrutura.


O Fluxo na API Go:

1. **Recepção de Carga:** O framework recebe uma avalanche de requisições HTTP (ou chamadas gRPC) enviadas pelo JMeter. Cada requisição instancia uma nova Goroutine de forma nativa.
2. **Abertura de Transação:** O UseCase invoca o repositório, que abre uma transação (`BEGIN`) no **Postgres** usando o driver `pgx`.
3. **Bloqueio de Linha (Contenção):** A API executa um `SELECT quantidade FROM ingressos WHERE id = $1 FOR UPDATE`.
* *Apenas uma Goroutine ganha a trava por vez.* As outras centenas de Goroutines que baterem na mesma linha ficam em estado de espera compulsória (*Waiting*), travando a execução do framework naquela requisição específica.


4. **Validação e Chamada Externa:** A Goroutine que conseguiu a trava valida se há estoque disponível. Se sim, ela faz uma chamada HTTP síncrona para o **Wiremock** (simulando a autorização do cartão de crédito), configurado com uma latência variável.
5. **Persistência e Liberação:** Após o retorno do Wiremock, a API executa o `UPDATE` decrementando o estoque e finaliza com um `COMMIT`. Somente neste momento a linha do banco é liberada para a próxima Goroutine da fila.
6. **Esgotamento:** Quando o estoque chega a 0, as Goroutines subsequentes passam direto pela trava, validam que o estoque acabou e retornam imediatamente `422 Unprocessable Entity` para o JMeter, sem efetuar chamadas ao Wiremock ou updates no banco.


#### O Pix Black Friday (Fila de Processamento Assíncrono / Event-Driven)
Enquanto o **"Ingresso Esgotado"** testa o bloqueio síncrono no banco de dados, este cenário testa a capacidade do framework de receber uma avalanche de dados, responder instantaneamente ao cliente e processar a lógica pesada em background usando Goroutines e canais (Worker Pool).

O Cenário: Uma API de e-commerce recebe ***X*** requisições por segundo de notificações de pagamento Pix. A API precisa apenas validar o formato do payload, responder 202 Accepted para o cliente e processar a baixa do pedido em background.

O Fluxo na *API*: 
1. O Framework recebe a requisição.
2. Valida o JSON e joga o evento para um canal interno do Go (chan).
3. Responde imediatamente 202 Accepted (o JMeter mede tempos de resposta baixíssimos aqui).
4. Um pool de Goroutines (Workers) consome desse canal, faz um SELECT/UPDATE no Postgres para mudar o status do pedido e chama o Wiremock para notificar o sistema de logística.
