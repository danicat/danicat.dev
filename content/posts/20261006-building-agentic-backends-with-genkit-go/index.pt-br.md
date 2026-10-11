---
title: "Por Que Reescrevi Meu Agente RAG em Python Usando Go e Genkit"
date: 2026-10-06
draft: false
categories:
  - Agent Development
tags:
  - agent-skills
  - ai-agents
  - gemini
  - genkit
  - golang
  - osquery
  - rag
series:
  - Gemini for Go Developers
series_order: 4
heroStyle: big
slug: "building-agentic-backends-with-genkit-go"
description: "Veja como substituí um agente legado em Python com SQLite RAG por um binário único em Go usando Genkit, templates dotprompt tipados e Agent Skills sob demanda."
summary: "Quando sua base de código começa a te segurar, jogue tudo fora e comece do zero. Veja como substituí um pipeline RAG em Python por um agente compilado em Go com Genkit e Agent Skills."
proficiencyLevel: "Advanced"
dependencies:
  - "Go 1.25+"
  - "github.com/firebase/genkit/go"
  - "google.golang.org/genai"
---

Há pouco mais de um ano, construí a [AIDA, a Agente de Diagnóstico por IA]({{< ref "/posts/20250531-diagnostic-agent" >}}), o primeiro agente autônomo que criei. A missão dela é diagnosticar problemas no computador consultando o sistema operacional por meio do [osquery](https://osquery.io/). Embora a AIDA tenha passado por algumas atualizações ao longo dos meses, nunca havia questionado sua arquitetura de verdade até recentemente.

Quando construí a AIDA originalmente em Python, os modelos estavam em outro patamar de capacidade (`gemini-2.5-flash` na época) e técnicas como [Agent Skills](https://agentskills.io) ainda não haviam sido inventadas. Como o `osquery` expõe centenas de tabelas que variam entre macOS, Linux e Windows, construí uma [solução local de RAG baseada em SQLite]({{< ref "/posts/20251103-building-aida-part-2" >}}) para injetar definições de schemas sob demanda no contexto. Mesmo sendo um conjunto de dados pequeno e estático (uma tabela por chunk, com embeddings gerados uma única vez), montar um banco vetorial para isso era trabalhoso, e o modelo nem sempre formulava as melhores consultas para a ferramenta de RAG, às vezes recuperando as tabelas erradas ou deixando passar uma tabela relacionada.

Mas não era apenas o RAG. Toda a base de código da AIDA havia sido feita na base da gambiarra em vez de engenharia de verdade, usando a fase inicial do "vibe coding".

Venho dizendo isso há mais de uma década, e hoje é mais verdade do que nunca: se a sua base de código está te segurando, jogue tudo fora e comece do zero. Enfrentar as dores de crescimento da v1 ensina exatamente o que simplificar na v2. E enquanto reescritas costumavam exigir meses de esforço e capital político, hoje você consegue iterar por três arquiteturas limpas antes do almoço.

Depois de [comparar as três principais bibliotecas de agentes em Go]({{< ref "/posts/20260825-gemini-for-go-developers-part-3-building-agents" >}}) (o [GenAI SDK](https://pkg.go.dev/google.golang.org/genai), o [Genkit](https://genkit.dev) e o [ADK](https://adk.dev)), decidi reconstruir a AIDA do zero em Go usando o **Genkit**. Migrar de Python e RAG para Genkit Go me permitiu:

1. **Compilar a AIDA em um único binário Go autocontido** que roda tanto como ferramenta local de CLI quanto como serviço HTTP no Cloud Run.
2. **Substituir o pipeline de RAG em SQLite por Agent Skills sob demanda**, entregando ao modelo o catálogo completo de tabelas do `osquery` para cada sistema operacional sem depender de buscas difusas.
3. **Garantir segurança de tipos em tempo de compilação** desde as entradas do prompt até as saídas estruturadas usando `dotprompt` e structs Go.
4. **Evoluir de um fluxo único para delegação multiagente com estado**, isolando a telemetria ruidosa do sistema das pesquisas na web.

Vamos ver como cada parte dessa nova arquitetura funciona na prática.

## Configurando seu agente de programação e fluxo de trabalho local

Antes de escrever qualquer código de aplicação, precisamos de duas coisas: a CLI local `genkit` para a interface de desenvolvimento e utilitários de linha de comando, e as Agent Skills que ensinam ao seu agente de programação como o Genkit Go funciona.

### Genkit CLI e Dev UI

Para instalar a CLI `genkit`, execute o script oficial de instalação:

```sh
curl -sL cli.genkit.dev | bash
```

A CLI `genkit` é a espinha dorsal do desenvolvimento local no Genkit Go. Se você executar seu programa diretamente com `go run .`, ele funciona como um binário Go normal, mas ao prefixar seu comando de execução com `genkit start --` você terá acesso à **Genkit Developer UI** (por padrão em `http://localhost:4000`), onde pode inspecionar visualmente e testar os componentes do seu programa Genkit de forma individual:

```sh
genkit start -- go run .
```

Na Dev UI, você pode testar fluxos, prompts e agentes registrados de forma interativa, além de inspecionar traces de prompts, chamadas de ferramentas, uso de tokens e latência.

Além da Dev UI, a CLI `genkit` também fornece utilitários de linha de comando para executar fluxos, visualizar traces e navegar pela documentação. Esse último recurso é particularmente útil para evitar alucinações do modelo durante a fase de implementação.

Aqui estão alguns exemplos:

```sh
# Run a specific flow once from the terminal and exit (--stream or --wait optional)
genkit flow:run diagnose '{"userQuestion": "how is my disk usage?"}' --non-interactive -- go run .

# Browse and read the embedded Genkit Go documentation offline
genkit docs:list go
genkit docs:read go/middleware.md
```

### Skills recomendadas

O padrão ouro para SDKs hoje em dia é distribuí-los com Agent Skills oficiais, e com o Genkit não é diferente. Se você está desenvolvendo um programa em Genkit Go, esta é a skill que você procura:

- [`developing-genkit-go`](https://github.com/genkit-ai/skills): a skill oficial do Genkit Go cobrindo fluxos, `dotprompt`, middlewares nativos e agentes experimentais.

Como estamos desenvolvendo com modelos Gemini, também é útil instalar a skill `gemini-api-dev`:
- [`gemini-api-dev`](https://github.com/google-gemini/gemini-skills): a skill oficial do Gemini para IDs de modelos atuais e configuração de raciocínio (*thinking*).

Você pode instalá-las no seu workspace usando a [CLI `skills`](https://github.com/vercel-labs/skills) da Vercel:

```sh
npx skills add genkit-ai/skills --skill developing-genkit-go
npx skills add google-gemini/gemini-skills --skill gemini-api-dev
```

Ou registrar os catálogos com o [`kungfu`](https://github.com/danicat/kungfu) para carregamento just-in-time (JIT):

```sh
kungfu catalog add genkit-ai/skills
kungfu catalog add google-gemini/gemini-skills
```

Se você preferir usar um servidor MCP em vez de skills, a CLI também inclui um [servidor MCP](https://genkit.dev/docs/go/mcp-server/) (`genkit mcp`). Ultimamente, tenho preferido usar mais skills, mas os MCPs ainda têm valor para certos casos de uso específicos (como consulta de documentação).

## Conceitos principais

O Genkit é um framework open-source para aplicações de IA generativa criado originalmente pela equipe do Firebase antes de crescer e se tornar um projeto próprio. Embora sua linguagem original seja JS, o Genkit Go parece surpreendentemente idiomático para desenvolvedores Go experientes, mesmo que ainda tenha algumas arestas para aparar aqui e ali. Um programa típico em Genkit não será muito diferente de um serviço web tradicional, graças a adaptadores como `genkit.Handler` e construções familiares como middlewares e fluxos e prompts fortemente tipados.

### Fluxos (*Flows*)

A unidade fundamental no Genkit é um **fluxo** (*flow*): uma função fortemente tipada que encapsula chamadas de IA generativa junto com qualquer lógica de pré ou pós-processamento. Vamos começar a AIDA como um fluxo de diagnóstico de turno único equipado com uma ferramenta para consultar o sistema operacional via `osqueryi`:

```go
type AIDARequest struct {
	OS           string `json:"os,omitempty" jsonschema_description:"The target operating system (defaults to runtime.GOOS)"`
	UserQuestion string `json:"userQuestion" jsonschema_description:"The diagnostic question to investigate"`
}

type AIDAResponse struct {
	Message string `json:"message" jsonschema_description:"The diagnostic findings and recommendations"`
}

type OsqueryInput struct {
	Query string `json:"query" jsonschema_description:"The SQL query to execute in osquery."`
}

g := genkit.Init(ctx,
	genkit.WithPlugins(&googlegenai.VertexAI{}),
	genkit.WithDefaultModel("vertexai/gemini-3.8-flash"),
)

runOsquery := genkit.DefineTool(g, "runOsquery", "Runs a system inspection query using osquery.",
	func(ctx *ai.ToolContext, in OsqueryInput) (string, error) {
		out, err := exec.CommandContext(ctx, "osqueryi", "--json", in.Query).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("osqueryi failed: %w: %s", err, out)
		}
		return string(out), nil
	},
)

diagnoseFlow := genkit.DefineFlow(
	g,
	"diagnose",
	func(ctx context.Context, req AIDARequest) (AIDAResponse, error) {
		if req.OS == "" {
			req.OS = runtime.GOOS
		}
		resp, err := genkit.Generate(ctx, g,
			ai.WithSystem(fmt.Sprintf("You are the Emergency Diagnostic Agent for diagnosing %s operating system failures. Your name is AIDA.", req.OS)),
			ai.WithPrompt(req.UserQuestion),
			ai.WithTools(runOsquery),
		)
		if err != nil {
			return AIDAResponse{}, fmt.Errorf("diagnose flow failed: %w", err)
		}
		return AIDAResponse{Message: resp.Text()}, nil
	},
)
```

`genkit.DefineTool` registra a ferramenta `runOsquery` usando a struct `OsqueryInput` para gerar seu JSON schema, enquanto `genkit.DefineFlow` registra o fluxo pelo nome com uma função de callback. Se você observar a assinatura do callback, ela é muito parecida com um handler HTTP tipado — recebe um `context.Context` e uma struct de requisição fortemente tipada, e retorna uma resposta tipada ao lado de um `error` padrão do Go.

Repare na variável `g` passada como primeiro argumento para `DefineTool`, `DefineFlow` e `Generate`. `g` é uma instância de `*genkit.Genkit`, que atua como o registro central de todos os fluxos, prompts, ferramentas, modelos e middlewares do seu programa. Você inicializa `g` uma única vez na função `main` com `genkit.Init` e a passa explicitamente.

Vou ser sincera: não sou muito fã de `genkit.WithDefaultModel`, pois prefiro que o código seja explícito, mas ele ajuda a reduzir um pouco de código repetitivo (*boilerplate*) se você estiver definindo muitos fluxos. Veremos uma opção melhor do que fixar modelos diretamente no código dos fluxos quando falarmos sobre prompts.

`genkit.Init` configura tanto o registro em tempo de execução quanto a stack de observabilidade. Cada chamada de modelo e ferramenta dentro de um fluxo é registrada automaticamente como um span filho dentro do trace do fluxo. Se o seu fluxo também realizar pré ou pós-processamento customizado — como consultar um banco de dados ou chamar uma API externa —, encapsular essa lógica em `genkit.Run` registra a operação como uma subetapa nomeada própria na cascata do trace.

Para fins de desenvolvimento, você pode executar um fluxo via CLI ou na Dev UI. Em produção, você pode invocar fluxos como uma chamada de função Go normal (por exemplo, `myFlow.Run`) ou expô-los como um endpoint HTTP. Transformar um fluxo em um handler HTTP leva apenas uma linha com `genkit.Handler`:

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /diagnose", genkit.Handler(diagnoseFlow))

port := os.Getenv("PORT")
if port == "" {
	port = "3400"
}

// Note: "server" requires the server plugin:
// import "github.com/firebase/genkit/go/plugins/server"
log.Fatal(server.Start(ctx, "0.0.0.0:"+port, mux))
```

Se você quiser expor todos os fluxos registrados na sua aplicação de uma só vez, pode iterar sobre `genkit.ListFlows(g)`:

```go
mux := http.NewServeMux()
for _, flow := range genkit.ListFlows(g) {
	mux.HandleFunc("POST /"+flow.Name(), genkit.Handler(flow))
}
log.Fatal(server.Start(ctx, "0.0.0.0:"+port, mux))
```

Vale lembrar que `genkit.Handler` não realiza nenhuma autenticação por conta própria, portanto proteja endpoints públicos com o seu middleware de autenticação padrão do `net/http`.

### Prompts

Fixar prompts e configurações de modelo diretamente no código do fluxo funciona para projetos pequenos, mas pode ficar difícil de manter conforme o projeto cresce. Mesmo que os agentes de programação façam a maior parte do trabalho pesado hoje em dia, não devemos jogar fora anos de boas práticas de engenharia, e a organização de código é uma delas.

O Genkit oferece suporte a templates de prompt usando um formato chamado [dotprompt](https://genkit.dev/docs/go/dotprompt/) (arquivos `*.prompt`). Os prompts em `dotprompt` podem conter mais do que apenas as instruções para o modelo; por exemplo, podem incluir metadados para selecionar e configurar o modelo, definições de schema, instruções de sistema e histórico de conversa.

Extrair o prompt, as ferramentas e a configuração de modelo da AIDA de dentro do `diagnoseFlow` para o arquivo `prompts/aida.prompt` fica assim:

```markdown
---
model: vertexai/gemini-3.8-flash
tools:
  - runOsquery
config:
  thinkingConfig:
    thinkingLevel: low
input:
  schema:
    os: string
    userQuestion: string
output:
  format: json
  schema:
    message: string
---
{{role "system"}}
You are the Emergency Diagnostic Agent for diagnosing {{os}} operating system failures. Your name is AIDA.

{{role "user"}}
{{userQuestion}}
```

Por padrão, `genkit.Init` carrega automaticamente os arquivos `.prompt` do diretório `./prompts` no seu diretório de trabalho. Se você preferir distribuir seu agente como um único binário, pode embutir o diretório `prompts` diretamente no binário com `//go:embed` e passá-lo para `genkit.WithPromptFS`:

```go
import "embed"

//go:embed prompts/*
var promptsFS embed.FS

func main() {
	g := genkit.Init(context.Background(), genkit.WithPromptFS(promptsFS))
	// ...
}
```

Os prompts ficam acessíveis no código por meio do método `genkit.LookupPrompt` e são invocados com `myPrompt.Execute`:

```go
aida := genkit.LookupPrompt(g, "aida")
resp, err := aida.Execute(ctx, ai.WithInput(map[string]any{
	"os":           runtime.GOOS,
	"userQuestion": "how is my disk usage?",
}))
```

Se você programa em Go há algum tempo, aquele `map[string]any` provavelmente te incomoda tanto quanto me incomoda. Felizmente, o Genkit nos permite vincular as structs `AIDARequest` e `AIDAResponse` que definimos anteriormente diretamente ao `aida.prompt` usando **data prompts** (`genkit.LookupDataPrompt`) e `genkit.DefineSchemaFor`:

```go
// In prompts/aida.prompt, reference the Go types by name:
//
// input:
//   schema: AIDARequest
// output:
//   schema: AIDAResponse
//
// Register both schemas with Genkit so the prompt file can resolve them:
genkit.DefineSchemaFor[AIDARequest](g)
genkit.DefineSchemaFor[AIDAResponse](g)

aida := genkit.LookupDataPrompt[AIDARequest, *AIDAResponse](g, "aida")
resp, _, err := aida.Execute(ctx, AIDARequest{
	OS:           runtime.GOOS,
	UserQuestion: "how is my disk usage?",
})
// resp is of type *AIDAResponse
// Note: The second return value is the raw *ai.ModelResponse
```

Com `DefineSchemaFor` e `LookupDataPrompt`, você define seu schema uma única vez em Go e mantém total segurança de tipos em tempo de compilação, desde a entrada até a saída processada.

### Plugins

Você estende o Genkit com pacotes sob o namespace `plugins/`. Embora o caso mais comum para um plugin seja registrar uma família de modelos, também temos plugins para capacidades mais amplas, como `server` e `middleware`. O `server` é um caso especial no sentido de que não precisa ser registrado para ser útil, mas a maioria dos plugins será inicializada na chamada `genkit.Init`:

```go
import (
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/middleware"
)

g := genkit.Init(ctx,
	genkit.WithPlugins(
		&googlegenai.VertexAI{},  // Gemini via Vertex AI
		&middleware.Middleware{}, // Registers Retry, Fallback, Skills, etc.
	),
)
```

O conceito de middleware deve ser familiar para a maioria dos desenvolvedores Go, mas é importante destacar esse plugin, pois ele é a porta de entrada para recursos críticos como fallbacks, novas tentativas (*retries*) e Agent Skills. Vamos analisá-lo mais de perto a seguir.

### Substituindo o SQLite RAG pelo middleware de Agent Skills

A AIDA já consegue executar comandos `osquery` por meio da ferramenta `runOsquery` que definimos antes, mas ela ainda depende dos pesos gerais de treinamento do modelo para adivinhar quais tabelas e colunas existem em cada sistema operacional (`darwin`, `linux` e `windows`).

Na versão legada em Python da AIDA, resolvi isso com um banco local de RAG em SQLite onde cada chunk continha a definição de uma tabela. Embora o conjunto de dados fosse pequeno e estático, a abordagem com RAG era trabalhosa de montar e tinha uma falha sutil: o `gemini-2.5-flash` precisava adivinhar primeiro o que buscar e, quando a consulta de RAG saía um pouco fora do alvo, ele trazia as tabelas erradas ou deixava passar uma tabela relacionada completamente.

Na reescrita em Go, apaguei o pipeline de RAG e o substituí por [Agent Skills](https://agentskills.io). Apontando o Antigravity para o [repositório do osquery](https://github.com/osquery/osquery), gerei três skills específicas por sistema operacional (`darwin`, `linux` e `windows`) na pasta `./skills`, cada uma contendo o catálogo completo de tabelas e schemas daquela plataforma. Em vez de torcer para uma busca vetorial retornar a tabela certa, a AIDA simplesmente ativa a skill do sistema operacional atual e carrega a referência completa de schemas sob demanda.

O Genkit anexa middlewares às chamadas de geração usando `ai.WithUse`, e o pacote `plugins/middleware` fornece implementações nativas para retentativas, fallbacks de modelo e [Agent Skills](https://agentskills.io):

```go
resp, err := genkit.Generate(ctx, g,
	ai.WithSystem(fmt.Sprintf("You are AIDA, a %s operating system diagnostic assistant.", runtime.GOOS)),
	ai.WithPrompt("My machine is running out of file descriptors. How do I inspect open files with osquery?"),
	ai.WithTools(runOsquery),
	ai.WithUse(
		&middleware.Retry{MaxRetries: 3, InitialDelayMs: 1000, BackoffFactor: 2},
		&middleware.Fallback{
			Models: []ai.ModelRef{
				googlegenai.ModelRef("vertexai/gemini-3.5-flash-lite", nil),
			},
		},
		&middleware.Skills{
			SkillPaths: []string{"./skills"},
		},
	),
)
```

Assim como um middleware HTTP, `ai.WithUse` compõe os itens na ordem declarada: colocar `Retry` antes de `Fallback` tenta novamente o modelo principal antes de recorrer ao modelo secundário.

Quando você aponta `&middleware.Skills{}` para `./skills`, o Genkit examina cada subdiretório, injeta as descrições das skills no prompt de sistema e registra uma ferramenta `use_skill` para que a AIDA possa carregar o schema do sistema operacional e os padrões de consulta corretos sob demanda.

Vale notar que, diferentemente do `dotprompt`, no momento o middleware `Skills` não suporta sistemas de arquivos embutidos (*embedded filesystems*), então você ainda precisa incluir um diretório de skills junto com o seu deploy.

Como registramos `&middleware.Middleware{}` durante o `genkit.Init`, também podemos declarar esses middlewares diretamente dentro de `prompts/aida.prompt`:

```yaml
---
model: vertexai/gemini-3.8-flash
tools:
  - runOsquery
use:
  - name: genkit-middleware/retry
    config:
      maxRetries: 2
  - genkit-middleware/skills
---
```

Para casos de uso avançados, você também pode escrever middlewares customizados implementando a interface `ai.Middleware` para interceptar e encapsular chamadas de geração, modelo ou ferramenta com a sua própria lógica.

### A Developer UI

Você encontra o código completo até este ponto no repositório complementar em [`github.com/danicat/gemini-for-go-developers/part-4/01-flow`](https://github.com/danicat/gemini-for-go-developers/tree/main/part-4/01-flow).

Para testar a AIDA na Genkit Developer UI, inicie a aplicação a partir desse diretório com `genkit start`:

```sh
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_CLOUD_LOCATION=global # required for Gemini 3.x
genkit start -- go run .
```

Abra `http://localhost:4000`, selecione o fluxo **diagnose** na barra lateral e execute-o com o seguinte payload de teste:

```json
{
  "userQuestion": "what is my battery status?"
}
```

Aqui está a saída desse prompt na minha máquina (observe o trace no painel direito mostrando que a skill foi ativada):

![Genkit Developer UI executando o fluxo diagnose](image-1.png)

## Agentes com estado (experimental)

Normalmente prefiro manter distância de recursos experimentais, mas este é uma exceção. Embora seja perfeitamente possível implementar agentes baseados apenas em fluxos, desenvolver padrões mais complexos — como chatbots, agentes seriais e agentes paralelos — exige um certo esforço para construir os padrões certos usando goroutines, sincronização, tratamento de erros e afins.

A introdução da abstração de agente simplifica esse processo ao nos fornecer uma API muito mais ergonômica, focada no aspecto declarativo da criação de agentes.

### Definindo um agente

Como a API ainda é experimental, você precisa habilitá-la com `genkit.WithExperimental()` ao inicializar o Genkit. Veja como transformamos a AIDA em uma agente de terminal multi-turno com estado, apoiada por um armazenamento de sessão (*session store*):

```go
import (
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/ai/exp/localstore"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
)

g := genkit.Init(ctx,
	genkit.WithExperimental(),
	genkit.WithPlugins(&googlegenai.VertexAI{}, &middleware.Middleware{}),
)

store := localstore.NewInMemorySessionStore[struct{}]()

aida := genkitx.DefineAgent(g, "aida",
	aix.InlinePrompt{
		ai.WithModelName("vertexai/gemini-3.8-flash"),
		ai.WithSystem(fmt.Sprintf("You are AIDA, a %s operating system diagnostic assistant. Use runOsquery to inspect the host.", runtime.GOOS)),
		ai.WithTools(runOsquery),
		ai.WithUse(&middleware.Skills{SkillPaths: []string{"./skills"}}),
	},
	aix.WithSessionStore(store),
)

// Interactive terminal session maintaining state across turns
scanner := bufio.NewScanner(os.Stdin)
var sessionID string

fmt.Print("AIDA> ")
for scanner.Scan() {
	input := strings.TrimSpace(scanner.Text())
	if input == "" || input == "exit" {
		break
	}

	var opts []aix.InvocationOption[struct{}]
	if sessionID != "" {
		opts = append(opts, aix.WithSessionID[struct{}](sessionID))
	}

	out, err := aida.RunText(ctx, input, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Print("AIDA> ")
		continue
	}

	sessionID = out.SessionID
	fmt.Printf("\n%s\n\nAIDA> ", out.Message.Text())
}
```

Armazenar `out.SessionID` e passá-lo nos turnos seguintes preserva o histórico da conversa ao longo da sessão. O parâmetro de tipo `[struct{}]` simplesmente informa ao Genkit que queremos apenas persistir o histórico de mensagens sem anexar nenhum estado de sessão customizado.

As mesmas técnicas usadas para melhorar o código em torno dos fluxos podem ser aplicadas aos agentes. Para evitar fixar os prompts no código, você pode usar `genkitx.DefinePromptAgent` para vincular `prompts/aida.prompt` diretamente ao agente.

Para expor a AIDA via HTTP de modo que um frontend remoto possa se conectar a ela, `genkitx.AllAgentRoutes` monta endpoints de turno, snapshot e cancelamento (*abort*) em roteadores Go padrão:

```go
mux := http.NewServeMux()
for _, route := range genkitx.AllAgentRoutes(g) {
	mux.HandleFunc(route.Pattern(), route.Handler())
}
```

Você encontra a implementação completa do agente com estado no repositório complementar em [`github.com/danicat/gemini-for-go-developers/part-4/02-agent`](https://github.com/danicat/gemini-for-go-developers/tree/main/part-4/02-agent).

### Delegação multiagente

Existem casos em que você precisará de mais de um agente para realizar uma tarefa. Os benefícios de uma arquitetura multiagente incluem a separação de responsabilidades — para melhorar a atenção do agente e reduzir a degradação de contexto (*context rot*) — e a paralelização — para melhorar os tempos de resposta.

Usando o middleware experimental `middlewarex.Agents` do Genkit, podemos dividir o trabalho investigativo da AIDA entre dois especialistas, mantendo `aida` como a persona coordenadora principal:

- **`inspector`**: Foca exclusivamente em consultar a telemetria do host via `runOsquery` e as skills de `osquery` específicas do sistema operacional.
- **`researcher`**: Usa o aterramento com a Busca do Google (*Google Search grounding*) para pesquisar processos desconhecidos, binários, mensagens de erro e CVEs na web sem inflar o histórico da conversa principal da AIDA.

```go
// requires:
// import (
// 	"google.golang.org/genai"
// 	middlewarex "github.com/firebase/genkit/go/plugins/middleware/exp"
// )

inspector := genkitx.DefineAgent(g, "inspector",
	aix.InlinePrompt{
		ai.WithModelName("vertexai/gemini-3.8-flash"),
		ai.WithSystem(fmt.Sprintf("Inspect %s system telemetry using osquery and report diagnostic findings.", runtime.GOOS)),
		ai.WithTools(runOsquery),
		ai.WithUse(
			&middleware.Retry{MaxRetries: 5},
			&middleware.Skills{SkillPaths: []string{"./skills"}},
		),
	},
	aix.WithDescription[struct{}]("Inspects host metrics, running processes, and system configuration using osquery."),
)

researcher := genkitx.DefineAgent(g, "researcher",
	aix.InlinePrompt{
		ai.WithModelName("vertexai/gemini-3.8-flash"),
		ai.WithSystem("Research unfamiliar processes, binaries, error messages, and CVEs using Google Search and report concise findings."),
		ai.WithConfig(&genai.GenerateContentConfig{
			Tools: []*genai.Tool{
				{GoogleSearch: &genai.GoogleSearch{}},
			},
		}),
	},
	aix.WithDescription[struct{}]("Searches the web for information on unfamiliar processes, binaries, error codes, and CVEs."),
)

aida := genkitx.DefineAgent(g, "aida",
	aix.InlinePrompt{
		ai.WithModelName("vertexai/gemini-3.8-flash"),
		ai.WithSystem(fmt.Sprintf("You are AIDA, a %s operating system diagnostic assistant. Delegate host inspection to inspector and web research to researcher, then synthesize findings for the user.", runtime.GOOS)),
		ai.WithUse(
			&middleware.Retry{MaxRetries: 5},
			&middlewarex.Agents{
				Agents:         []aix.AgentRef{inspector.Ref(), researcher.Ref()},
				MaxDelegations: 5,
			},
		),
	},
	aix.WithSessionStore(store),
)
```

A delegação é executada de forma síncrona por padrão, embora você possa definir `Async: true` em `middlewarex.Agents` caso queira que a AIDA dispare tarefas especializadas de longa duração em segundo plano e reúna os resultados entre os turnos.

Um ponto importante para ter em mente é o trade-off de latência. Quando cronometrei exatamente a mesma consulta `"what is my battery status?"` nas duas versões (com `thinkingLevel` definido como `low`), a versão com um único agente (`02-agent`) terminou em **9,56s** ao longo de 3 chamadas de LLM, enquanto esta versão multiagente (`03-multi-agent`) levou **17,02s** ao longo de 5 chamadas de LLM. A latência por chamada é basicamente idêntica (~3,3s), então os ~7,5 segundos extras vêm puramente das duas viagens de ida e volta adicionais da coordenadora quando `aida` delega para `inspector` e depois resume o relatório dele.

Assim como na programação paralela, o custo de coordenação para um *N* pequeno pode deixar uma tarefa simples mais lenta. Para uma consulta simples de exemplo, um único agente é mais rápido; em um sistema do mundo real, contudo, dividir um "superagente" em subagentes focados geralmente compensa essas viagens extras para isolar saídas ruidosas de ferramentas, otimizar o uso de contexto e melhorar a qualidade geral das respostas.

Você encontra o exemplo multiagente completo no repositório complementar em [`github.com/danicat/gemini-for-go-developers/part-4/03-multi-agent`](https://github.com/danicat/gemini-for-go-developers/tree/main/part-4/03-multi-agent).

## Fazendo o deploy em produção

### Runtimes

Como uma aplicação Genkit Go compila para um binário Go padrão usando `net/http`, não há dependência (*lock-in*) de nenhum runtime proprietário. Isso significa que praticamente qualquer plataforma que você já usa para hospedar seus servidores HTTP serve para aplicações Genkit.

Minha primeira opção é sempre um runtime de containers serverless. No Google, uso principalmente o **Google Cloud Run**, escolhendo plataformas mais complexas (como o **Google Kubernetes Engine (GKE)** ou o **Gemini Enterprise**) apenas quando esbarro em algum bloqueio crítico que exige um recurso que o Cloud Run não me oferece.

Também vale dizer que, no passado, esse tipo de decisão costumava ser mais importante do que é hoje. Mudar de plataforma (*replatforming*) dava muito trabalho, então as pessoas tinham a tendência de querer "acertar de primeira" logo no início. Com agentes de programação, conseguimos adaptar uma aplicação para uma nova plataforma em questão de dias (se não horas), então não há mais desculpas para adicionar complexidade cedo demais a um projeto.

Essa é a teoria. Na prática, ao longo do último ano não encontrei nenhum caso que eu não conseguisse resolver com o Cloud Run. Ele é simplesmente conveniente demais! A principal exceção é quando você está construindo fortemente em torno do [ADK](https://adk.dev) e quer persistência de sessão gerenciada e conectores corporativos prontos para uso, onde o Gemini Enterprise Agent Platform (que cobrimos em [Construindo Agentes de IA em Go]({{< ref "/posts/20260825-gemini-for-go-developers-part-3-building-agents" >}})) oferece fortes sinergias nativas.

Normalmente, os agentes — especialmente os baseados em Gemini, como o Antigravity — não precisam de ajuda para fazer deploy no Cloud Run. Mas, se o seu agente estiver de mau humor, você sempre pode instalar a skill do Cloud Run disponível em [github.com/google/skills](https://github.com/google/skills/tree/main/skills/cloud/cloud-run-basics) para dar um empurrãozinho nele.

### Observabilidade

Durante o desenvolvimento local, o comando `genkit start` grava traces completos de execução e logs estruturados em `.genkit/traces` (que você deve adicionar ao `.gitignore`) para que você possa inspecioná-los na Dev UI. Em produção, você quer ter essa mesma visibilidade sobre o uso de tokens, latência e traces passo a passo no seu backend de monitoramento.

Como o Genkit é construído sobre o [OpenTelemetry](https://opentelemetry.io/), você tem duas opções principais: se estiver fazendo deploy no Google Cloud, o plugin `googlecloud` (ou o plugin `firebase`, que executa o mesmo exportador por baixo dos panos) envia tudo para o Cloud Logging, Cloud Trace e Cloud Monitoring. Se a sua equipe usa outro backend como Datadog, Grafana ou Honeycomb, você pode anexar um exportador OTLP padrão diretamente ao tracer provider do Genkit.

Para um deploy típico no Cloud Run, habilite o exportador do Google Cloud antes de chamar `genkit.Init`:

```go
import "github.com/firebase/genkit/go/plugins/googlecloud"

func main() {
	ctx := context.Background()

	googlecloud.EnableGoogleCloudTelemetry(&googlecloud.GoogleCloudTelemetryOptions{
		// Raw prompts and responses are logged by default; disable this for sensitive data:
		DisableLoggingInputAndOutput: true,
		// Set to true if you want to test Cloud export locally under `genkit start`:
		ForceDevExport: false,
	})

	g := genkit.Init(ctx,
		// if you are deploying to GCP, auth via Gemini Enterprise (aka Vertex AI) is preferred
		genkit.WithPlugins(&googlegenai.VertexAI{}),
	)
	// ...
}
```

Se você preferir um backend OTLP, registre um processador de spans padrão do OpenTelemetry antes de `genkit.Init` e use `defer` na sua função de encerramento (*shutdown*):

```go
import (
	"context"
	"log"

	"github.com/firebase/genkit/go/core/tracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func registerOTLP(ctx context.Context) func(context.Context) error {
	// Configured via standard OTEL_EXPORTER_OTLP_* environment variables.
	exp, err := otlptracegrpc.New(ctx)
	if err != nil {
		log.Fatalf("failed to build OTLP exporter: %v", err)
	}

	tp := tracing.TracerProvider()
	tp.RegisterSpanProcessor(sdktrace.NewBatchSpanProcessor(exp))
	return tp.Shutdown
}
```

## Conclusão

Reconstruir a AIDA do zero é um ótimo lembrete de quão rápido a engenharia de IA amadureceu nos últimos doze meses. O que antes exigia um pipeline de RAG customizado em SQLite, um modelo de embedding local e centenas de linhas de código Python frágil (uma crítica ao meu próprio vibe coding inicial, e não ao Python em si) agora é compilado em um único binário Go fortemente tipado, impulsionado pelo Genkit e por Agent Skills sob demanda.

Migrar para o Genkit Go trouxe para a AIDA validação de schemas em tempo de compilação com `dotprompt`, retentativas nativas e injeção de skills por middleware, além de uma delegação multiagente limpa, mantendo a inicialização rápida e a simplicidade de deploy de um serviço HTTP padrão em Go. Combinado com um modelo rápido como o Gemini 3.8 Flash, a diferença de agilidade em comparação com a versão legada é da água para o vinho.

Neste artigo focamos no runtime principal e na arquitetura de backend da AIDA, reservando o frontend (usando A2UI e design de interação generativa) para um artigo de continuação. Se você planeja construir agentes em Go, ou se tem um protótipo inicial com RAG começando a mostrar a idade, recomendo fortemente dar uma chance ao Genkit.
