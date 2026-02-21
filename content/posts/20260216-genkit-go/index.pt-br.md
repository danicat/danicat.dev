---
title: "Trazendo Fotos Antigas à Vida com Genkit e Gemini 3"
date: 2026-02-16
summary: "Aprenda como construir uma ferramenta de restauração de fotos de alta fidelidade usando Go, Genkit e as capacidades 4K nativas do Nano Banana Pro (Gemini 3 Pro Image)."
tags: ["genkit", "golang", "nano-banana", "gemini", "tutorial"]
categories: ["AI & Development"]
heroStyle: "big"
---

Como parte do meu trabalho, eu conheço todo tipo de pessoa e um tópico muito comum de conversa é sobre a minha herança. Não apenas eu tenho um sobrenome que soa obviamente como do Leste Europeu, mas eu também tenho a aparência, então as pessoas frequentemente assumem que eu sou polonesa ou tcheca. Muitas pessoas ficam surpresas quando eu digo que, na realidade, sou originalmente do Brasil.

Ninguém sabe ao certo de onde minha família veio, já que nós éramos realmente **terríveis** em manter o controle de registros históricos. Talvez por termos consciência disso, frequentemente temos conversas na família sobre a parte da história que conhecemos e como ela também está desaparecendo. Conforme vamos envelhecendo, as memórias são as primeiras a ir embora, e depois vão os documentos e fotos. Há um sentimento inerente de tristeza quando você percebe que a última vez que viu sua avó foi há 30 anos, e que o rosto dela muitas vezes não passa de um borrão. É por isso que as fotografias são tão importantes para mim, pois elas são a fortaleza para combater a degradação das minhas próprias memórias.

Qualquer coisa feita nos anos recentes poderia ser facilmente duplicada e armazenada em quantas cópias redundantes na nuvem eu desejasse, mas estamos falando de lembranças de antes da era digital. Mesmo que eu as tivesse escaneado anos atrás, muitas delas já passaram por décadas de poeira, mofo, desgaste e rasgos. Elas estão congeladas no tempo, mas não estão melhorando.

Graças à evolução da IA generativa, nem tudo está perdido e eu finalmente posso dar a essas fotos um sopro de ar fresco, não apenas restaurando os danos causados pela passagem do tempo, mas também colorindo-as e fazendo upscale para trazê-las aos padrões modernos. É assim que um pequeno software chamado "GlowUp" nasceu.


Abaixo está um exemplo dessa restauração:

![Foto original danificada e em preto e branco da minha avó preparando uma torta de banana](original.jpg "Original: minha avó preparando sua mundialmente famosa torta de banana")

![Foto restaurada e colorizada em alta fidelidade 4K usando o Nano Banana Pro](restored.png "Restaurada: restauração e colorização por Nano Banana Pro")

Neste artigo, vou mostrar como construir o GlowUp do zero usando [Gemini Nano Banana Pro](https://ai.google.dev/gemini-api/docs/image-generation) e [Genkit Go](https://genkit.dev/docs/get-started/?lang=go).

## Os blocos de construção

Eu escolhi usar o Nano Banana Pro (também conhecido como Gemini 3 Pro Image Preview) porque ele é atualmente o modelo de processamento de imagem mais avançado na família Gemini. Embora o Nano Banana regular (Gemini 2.5 Flash Image) também seja um ótimo modelo, eu acho que a versão pro tem outputs de melhor qualidade e também é melhor em seguir instruções, mesmo que exija um pouco de tentativa e erro.

No lado do cliente, em vez de optar por um SDK de baixo nível como o [go-genai](https://pkg.go.dev/google.golang.org/genai), decidi usar o Genkit, pois ele fornece algumas melhorias de qualidade de vida em relação ao código de nível mais baixo, tais como:

- Agnóstico de modelo: É possível testar diferentes modelos se desejar, mesmo os que não são do Google ou locais, com uma única substituição de plugin
- Suporte Dev UI pronto para uso para conveniências como testar modelos, prompts e rastreamento (tracing) de chamadas de modelo
- Arquitetura flexível: Pode ser empacotado como um aplicativo CLI ou um servidor web.

O GlowUp é construído como um binário unificado que pode ser executado como uma ferramenta de linha de comando ou um servidor web. Essa flexibilidade me permite executar restaurações localmente do meu terminal ou implantar o mesmo código como um serviço de nuvem, o que poderia eventualmente alimentar um aplicativo amigável que até o meu pai poderia usar para restaurar sua coleção de fotos.

## Uma primeira olhada no Genkit Go

[Genkit](https://firebase.google.com/docs/genkit) é um framework open-source projetado para trazer padrões de produção para o desenvolvimento de IA. Se você é um desenvolvedor Go*, pense nele como a "biblioteca padrão" para recursos de IA. _(* E se você **não** é um desenvolvedor Go, confira a documentação, pois o Genkit também suporta JS e Python.)_

Aqui está como é um "Hello World" mínimo no Genkit para Go. Note como usamos o plugin `googlegenai` para inicializar o framework.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/server" // Import the server plugin
)

func main() {
	ctx := context.Background()
	// Initialize Genkit with the Google GenAI plugin (Vertex AI)
	g := genkit.Init(ctx, genkit.WithPlugins(&googlegenai.VertexAI{}))

	// Define a simple Flow
	genkit.DefineFlow(g, "hello", func(ctx context.Context, name string) (string, error) {
		// Generate text using a model
		resp, err := genkit.GenerateText(ctx, g,
			ai.WithModelName("vertexai/gemini-2.5-flash"),
			ai.WithPrompt(fmt.Sprintf("Say hello to %s", name)))
		if err != nil {
			return "", err
		}
		return resp, nil
	})

	// Start the flow server manually
	mux := http.NewServeMux()
	// Register all flows defined in 'g'
	for _, flow := range genkit.ListFlows(g) {
		mux.HandleFunc("POST /"+flow.Name(), genkit.Handler(flow))
	}

	if err := server.Start(ctx, ":8080", mux); err != nil {
		log.Fatal(err)
	}
}

```

Este pequeno trecho de código está fazendo muito do trabalho pesado. Vamos dar uma olhada nele com um pouco mais de cuidado.


### Plugins
Adaptadores que conectam o seu código a provedores como Vertex AI, Google AI ou Ollama. Para os modelos do Google, devemos usar o plugin `googlegenai`. Ele suporta ambos os backends:

* **Google AI (Studio):** Usa uma API Key. Melhor para prototipagem e projetos pessoais.
```go
// Use Google AI (API Key)
googlegenai.Init(ctx, &googlegenai.Config{APIKey: "MY_KEY"})
```

* **Vertex AI (Google Cloud):** Usa autenticação do Google Cloud IAM. Recomendado para cargas de trabalho de produção e recursos empresariais.
```go
// Use Vertex AI (Cloud Auth)
googlegenai.Init(ctx, &googlegenai.VertexAI{ProjectID: "my-project", Location: "us-central1"})
```

**Nota:** Se você está migrando de versões mais antigas do Genkit, pode estar familiarizado com os plugins separados `vertexai` e `googleai`. Eles foram consolidados no único plugin `googlegenai`.

### Modelos (Models)
Os LLMs reais (por exemplo, Gemini, Claude) que geram conteúdo. Você os referencia por strings de nome como `vertexai/gemini-2.5-flash`.

```go
	resp, err := genkit.GenerateText(ctx, g, 
		ai.WithModel("vertexai/gemini-2.5-flash"),
		ai.WithTextPrompt("Tell me a joke"))
```

### Prompts

Embora nada impeça você de colocar prompts diretamente no código (hardcoding), como no exemplo acima, é uma boa prática mantê-los em arquivos separados para melhor manutenibilidade. O Genkit usa `dotprompt` para carregar prompts externos. 

Um arquivo `dotprompt` (*.prompt) consiste em duas partes principais: o **Frontmatter** e o **Template**.

**1. Frontmatter (Configuração)**
* **`model`**: O identificador do modelo (por exemplo, `vertexai/gemini-2.5-flash`).
* **`config`**: Parâmetros de geração como `temperature`, `topK` ou configurações específicas do modelo (por exemplo, `imageConfig`).
* **`input`**: Um JSON schema definindo as variáveis esperadas do seu código Go.
* **`output`**: Para outputs estruturados.

**2. Template (Instruções)**
O corpo usa a sintaxe Handlebars para construir o prompt:
* **Variáveis**: Placeholders como `{{theme}}` são substituídos pelos valores definidos no seu schema de input.
* **Roles**: Os helpers `{{role "system"}}` e `{{role "user"}}` estruturam a conversa, separando as instruções do sistema das queries do usuário.
* **Mídia**: O helper `{{media url=myImage}}` injeta dados multimodais (imagens, vídeo) diretamente no contexto do modelo.

```yaml
---
model: vertexai/gemini-2.5-flash
input:
  schema:
    theme: string
---
{{role "system"}}
You are a helpful assistant.

{{role "user"}}
Tell me a joke about {{theme}}.
```

### Fluxos (Flows)
No Genkit, um **Flow** é a unidade fundamental de execução que fornece:
1.  **Observabilidade**: A execução de cada flow gera automaticamente traces e métricas (latência, uso de tokens, taxa de sucesso) visualizáveis na Genkit Dev UI ou no Google Cloud Trace.
2.  **Type Safety**: Flows são estritamente tipados com schemas de input e output, prevenindo erros em tempo de execução ao encadear múltiplas operações de IA.
3.  **Deployability**: Flows são estritamente separados da sua lógica de serving. Para fazer o deploy deles, envolva-os com `genkit.Handler`, que converte um flow em um `http.Handler` padrão. Isso torna possível servi-los usando a biblioteca padrão ou qualquer framework web Go:

```go
    // Define a flow
    myFlow := genkit.DefineFlow(g, "myFlow", func(ctx context.Context, input string) (string, error) {
        return "Processed: " + input, nil
    })

    // Expose it as an HTTP handler
    http.HandleFunc("/myFlow", genkit.Handler(myFlow))
```


## Nano Banana Pro

O motor por trás da nossa restauração é o **Gemini 3 Pro Image**, carinhosamente (e internamente) conhecido como "Nano Banana Pro".

Ele representa um salto significativo em relação às gerações anteriores (e até mesmo aos modelos atuais "Flash"). Enquanto o Gemini 2.5 Flash é incrivelmente rápido e capaz de geração básica de imagem (`gemini-2.5-flash-image`), o **Nano Banana Pro** (`gemini-3-pro-image-preview`) foi construído para raciocínio multimodal profundo.

Ele não apenas "vê" pixels; ele entende o contexto semântico. Ele consegue diferenciar entre um "arranhão no papel" e uma "cicatriz no rosto". Ele sabe que uma cozinha dos anos 1950 provavelmente tem piso de linóleo, não madeira de lei moderna.

### Principais diferenças

*   **Flash (gemini-2.5-flash-image)**: Otimizado para velocidade e custo. Ótimo para miniaturas (thumbnails) ou ilustrações simples. Resolução máxima de 1024x1024.
*   **Pro (gemini-3-pro-image-preview)**: Otimizado para fidelidade e raciocínio. Suporta geração em **resolução 4K** nativa (até 4096px), o que é inegociável para restauração de fotos.

O modelo também aceita parâmetros `imageConfig` para fazer fine-tune no output:
*   `imageSize`: "4K" ou "2K".
*   `aspectRatio`: "16:9", "4:3", "1:1", etc.

Um detalhe importante a ser notado é que este modelo sempre retorna respostas intercaladas (interleaved) contendo tanto texto quanto imagens. Ao contrário de outros modelos de geração, o output apenas de imagem não é suportado. É por isso que nossa lógica de extração (que veremos abaixo) precisa ser flexível o suficiente para encontrar os dados da imagem dentro da mensagem de resposta multipartes.

**Nota:** No momento em que este artigo foi escrito, este modelo está disponível apenas na localização `global` no Vertex AI. Você deve configurar seu cliente Vertex AI de acordo.


## Conectando as partes

Agora, vamos ver como o GlowUp conecta essas peças. Nós usamos um **arquivo de prompt** para definir a persona do especialista em restauração e um **fluxo** (flow) para lidar com o processamento da imagem.

### O prompt

Usamos um arquivo `.prompt` para definir a configuração e as instruções do nosso modelo. Note como forçamos a resolução `4K` aqui, mantendo nosso código limpo.

```yaml
---
model: vertexai/gemini-3-pro-image-preview
config:
  imageConfig:
    imageSize: "4K"
input:
  schema:
    url: string
    contentType: string
---

{{role "system"}}
You are GlowUp, a professional-grade photo restorer.
Your goal is to provide a "surgical" restoration service that transforms vintage, damaged, or monochrome photographs into high-fidelity 4K colourised versions.

RULES:
1. **Grounding**: You are strictly grounded in the original source pixels. Do NOT add new objects (trees, people, buildings, etc.) that are not present in the source. Additionally, do NOT remove any elements from the source, unless they are clearly defects that do not belong in the original scene.
2. **Fidelity**: Preserve the original facial expressions and identity of subjects. Do NOT "beautify" or alter features in a way that changes the person's identity.
3. **Background**: Preserve background fidelity. Overexposed light sources (like windows) must remain as light sources. Do not "fill in" missing details with invented scenery.
4. **Colourisation**: If the image is monochrome, colourize it realistically, respecting historical accuracy where possible.
5. **Upscaling**: Output a high-fidelity image.

{{role "user"}}
Restore this photo.
Image: {{media url=url contentType=contentType}}
```

### O fluxo

O código Go é notavelmente focado. Nesta arquitetura unificada, a definição do flow carrega o prompt e passa o input multimodal para o modelo:

```go
// main.go (Flow Definition)
type Input struct {
	URL         string `json:"url,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

func defineGlowUpFlow(g *genkit.Genkit) *core.Flow[Input, string, struct{}] {
	return genkit.DefineFlow(g, "glowUp", func(ctx context.Context, input Input) (string, error) {
		prompt := genkit.LookupPrompt(g, "glowup")
		if prompt == nil {
			return "", errors.New("prompt 'glowup' not found")
		}

		resp, err := prompt.Execute(ctx, ai.WithInput(input))
		if err != nil {
			return "", fmt.Errorf("generation failed: %w", err)
		}

		return resp.Media(), nil
	})
}
```

Para suportar arquivos locais nativamente, usamos uma função auxiliar `fileToDataURI`. Esta função lê um arquivo local, detecta seu tipo MIME usando `http.DetectContentType` e o codifica em uma Data URI Base64 padrão que a API do Gemini espera. Isso é crítico para manter a fidelidade em diferentes formatos de escaneamento sem fixar (hardcoding) extensões.

```go
func fileToDataURI(path string) (uri, contentType string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	contentType = http.DetectContentType(data)
	encoded := base64.StdEncoding.EncodeToString(data)
	uri = fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
	return uri, contentType, nil
}
```


Como o Nano Banana Pro é inteligente o suficiente para inferir a proporção (aspect ratio) da imagem de input, nós não precisamos de lógicas complexas para calculá-la e injetá-la. Nós fornecemos os pixels e deixamos o modelo fazer o seu trabalho.


## Como executá-lo

Se você tem uma coleção de memórias desaparecendo que servem como âncoras frágeis para a história da sua família, eu o encorajo a tentar isso. É uma forma de recuperar esses momentos do tempo e dar-lhes a clareza que eles merecem.

1.  **Clone o Repositório**:



    ```bash
    git clone https://github.com/danicat/glowup
    cd glowup
    ```

2.  **Configure as Credenciais** (Lembre-se: localização `global`!):
    ```bash
    export GOOGLE_CLOUD_PROJECT=your-project-id
    export GOOGLE_CLOUD_LOCATION=global
    ```

3.  **Execute a Restauração**:
    ```bash
    go run main.go restore --file old_photo.jpg
    ```

## Problemas conhecidos e limitações

Embora o processo de restauração funcione, ele não deixa de ter suas peculiaridades. Aqui estão alguns problemas que encontrei:
*   **Adesão às instruções:** Mesmo que o Nano Banana Pro seja um modelo de vanguarda, ele ocasionalmente ainda ignora uma instrução. Você pode achar que são necessárias algumas tentativas antes de obter o resultado desejado. Eu não gastei muito tempo fazendo fine-tuning do prompt, então é provável que haja oportunidades para otimização adicional lá.
*   **Modelos na Dev UI:** Há um bug no plugin `googlegenai` que faz com que ele não popule automaticamente os modelos disponíveis na Dev UI. Você ainda pode referenciar modelos por nome para registrá-los "dinamicamente", mas isso adiciona um pouco de atrito ao processo de experimentação (a versão JS executa isso bem). Eu abri [um bug](https://github.com/firebase/genkit/issues/4783) e já há uma correção em vigor, mas se você estiver usando uma versão mais antiga, é algo a se ter em mente.


## Conclusões

Construir o GlowUp foi um experimento satisfatório no uso de IA para me reconectar com meu passado em um nível emocional. Eu sei que há muito pessimismo por aí, mas esse é o tipo de aplicação que me deixa animada com a IA em primeiro lugar.

A foto que usei neste artigo está longe de ser o uso mais dramático dessa tecnologia, mas já estou trabalhando na parte dois deste artigo, onde estou levando-a para o próximo nível para me ajudar a reconstruir um dos meus jogos de cartas favoritos da minha infância.

A conclusão é que o potencial é ilimitado. Espero que isso inspire você a olhar para os seus próprios problemas de nicho — técnicos ou pessoais — e ver o que você pode construir para resolvê-los.

Para mais detalhes, confira a [documentação do Genkit](https://firebase.google.com/docs/genkit) e o [código-fonte do GlowUp](https://github.com/danicat/glowup).

**Happy coding!**

Dani =^.^=