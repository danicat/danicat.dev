---categories:
- Applied GenAI
date: 2026-02-16
heroStyle: big
summary: Aprenda a construir uma ferramenta de restauração de fotos em alta fidelidade
  usando Go, Genkit e os recursos 4K nativos do Nano Banana Pro (Gemini 3 Pro Image).
tags:
  - gemini
  - genkit
  - golang
  - nano-banana
  - tutorial
title: "Trazendo Fotos Antigas à Vida com Genkit e Gemini 3"
slug: "genkit-go-photo-restoration"
aliases:
  - "/pt-br/posts/20260216-genkit-go/"
description: "Tutorial prático para criar uma ferramenta de restauração de fotos 4K em Go usando Genkit e Gemini 3 Pro Image (Nano Banana Pro) com dotprompt e flows."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "Genkit Go >= 0.1.0"
  - "Google Cloud Vertex AI"
---

Como parte do meu trabalho, conheço todo tipo de pessoa, e um tema muito comum de conversa é sobre a minha ascendência. Não apenas tenho um sobrenome com sonoridade evidente do Leste Europeu, como meus traços também lembram a região, então as pessoas frequentemente assumem que sou polonesa ou tcheca. Muita gente fica surpresa quando conto que, na realidade, sou brasileira nata.

Ninguém na família sabe ao certo de onde nossos antepassados vieram, já que sempre fomos realmente **péssimos** em guardar registros históricos. Talvez por termos consciência disso, frequentemente conversamos na família sobre a parte da história que conhecemos e como ela também está se perdendo. Conforme todos vamos envelhecendo, as memórias são as primeiras a ir embora, seguidas pelos documentos e fotos. Há um sentimento natural de tristeza ao perceber que a última vez que vi minha avó foi há 30 anos, e que o rosto dela muitas vezes não passa de um borrão. É por isso que fotografias são tão importantes para mim: elas são a fortaleza para combater a degradação das minhas próprias memórias.

Qualquer coisa registrada nos últimos anos pode ser facilmente duplicada e guardada em quantas cópias redundantes na nuvem eu desejar, mas aqui estamos falando de lembranças de antes da era digital. Mesmo que eu as tivesse digitalizado anos atrás, muitas já acumulam décadas de poeira, mofo, desgaste e ranhuras. Estão congeladas no tempo, mas sem chance de melhora por conta própria.

Graças à evolução da IA generativa, nem tudo está perdido: finalmente posso dar um sopro de ar fresco a essas fotos, não apenas restaurando os danos causados pela passagem do tempo, mas também colorindo e fazendo upscale para trazê-las aos padrões modernos. Foi assim que nasceu um pequeno software chamado "GlowUp".


Abaixo está um exemplo dessa restauração:

![Foto original danificada e em preto e branco da minha avó preparando uma torta de banana](original.jpg "Original: minha avó preparando sua mundialmente famosa torta de banana")

![Foto restaurada e colorizada em alta fidelidade 4K usando o Nano Banana Pro](restored.png "Restaurada: restauração e colorização por Nano Banana Pro")

Neste artigo, vou mostrar como construir o GlowUp do zero usando o [Gemini Nano Banana Pro](https://ai.google.dev/gemini-api/docs/image-generation) e o [Genkit Go](https://genkit.dev/docs/get-started/?lang=go).

## Os blocos de construção

Optei por usar o Nano Banana Pro (também conhecido como Gemini 3 Pro Image Preview) porque ele é atualmente o modelo de processamento de imagem mais avançado da família Gemini. Embora o Nano Banana tradicional (Gemini 2.5 Flash Image) também seja um ótimo modelo, sinto que a versão Pro entrega outputs de maior qualidade e segue instruções com mais precisão, ainda que demande um pouco de tentativa e erro.

No lado do cliente, em vez de optar por um SDK de baixo nível como o [go-genai](https://pkg.go.dev/google.golang.org/genai), decidi usar o Genkit, pois ele fornece várias melhorias de qualidade de vida em relação ao código de nível mais baixo, tais como:

- Agnóstico de modelo: posso testar diferentes modelos se desejar, mesmo locais ou de terceiros, com uma simples troca de plugin.
- Suporte nativo à Dev UI para conveniências como testar modelos, prompts e rastreamento (tracing) de chamadas de modelo.
- Arquitetura flexível: pode ser empacotado tanto como uma aplicação CLI quanto como um servidor web.

O GlowUp foi construído como um binário unificado que pode rodar como uma ferramenta de linha de comando ou um servidor web. Essa flexibilidade me permite executar restaurações localmente pelo terminal ou implantar o mesmo código como um serviço de nuvem, o que futuramente pode alimentar um aplicativo agradável que até meu pai consiga usar para restaurar o acervo de fotos dele.

## Uma primeira olhada no Genkit Go

O [Genkit](https://firebase.google.com/docs/genkit) é um framework open-source projetado para trazer padrões de produção ao desenvolvimento com IA. Se você é uma pessoa desenvolvedora Go*, pense nele como a "biblioteca padrão" para recursos de IA. _(* E se você **não** desenvolve em Go, confira a documentação, pois o Genkit também suporta JS e Python.)_

Aqui está como é um "Hello World" mínimo no Genkit para Go. Note como usamos o plugin `googlegenai` para inicializar o framework:

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

Este pequeno trecho de código está fazendo muito trabalho pesado. Vamos examiná-lo com um pouco mais de cuidado.


### Plugins
Adaptadores que conectam o seu código a provedores como Vertex AI, Google AI ou Ollama. Para os modelos do Google, devemos usar o plugin `googlegenai`. Ele suporta ambos os backends:

* **Google AI (Studio):** usa uma API Key. Ideal para prototipagem e projetos pessoais.
```go
// Use Google AI (API Key)
googlegenai.Init(ctx, &googlegenai.Config{APIKey: "MY_KEY"})
```

* **Vertex AI (Google Cloud):** usa autenticação IAM do Google Cloud. Recomendado para cargas de trabalho de produção e recursos empresariais.
```go
// Use Vertex AI (Cloud Auth)
googlegenai.Init(ctx, &googlegenai.VertexAI{ProjectID: "my-project", Location: "us-central1"})
```

**Nota:** Se você está migrando de versões anteriores do Genkit, pode estar familiarizada com os plugins separados `vertexai` e `googleai`. Eles foram consolidados no único plugin `googlegenai`.

### Modelos (Models)
Os LLMs reais (por exemplo, Gemini, Claude) que geram conteúdo. Você os referencia por strings de identificação, como `vertexai/gemini-2.5-flash`.

```go
	resp, err := genkit.GenerateText(ctx, g, 
		ai.WithModel("vertexai/gemini-2.5-flash"),
		ai.WithTextPrompt("Tell me a joke"))
```

### Prompts

Embora nada impeça você de fixar prompts diretamente no código (hardcoding), como no exemplo acima, é uma boa prática mantê-los em arquivos separados para facilitar a manutenção. O Genkit usa o `dotprompt` para carregar prompts externos. 

Um arquivo `dotprompt` (*.prompt) consiste em duas partes principais: o **Frontmatter** e o **Template**.

**1. Frontmatter (Configuração)**
* **`model`**: o identificador do modelo (por exemplo, `vertexai/gemini-2.5-flash`).
* **`config`**: parâmetros de geração como `temperature`, `topK` ou configurações específicas do modelo (por exemplo, `imageConfig`).
* **`input`**: um JSON Schema definindo as variáveis esperadas a partir do seu código Go.
* **`output`**: para saídas estruturadas.

**2. Template (Instruções)**
O corpo usa a sintaxe Handlebars para construir o prompt:
* **Variáveis**: marcadores como `{{theme}}` são substituídos pelos valores definidos no seu schema de entrada.
* **Roles**: os helpers `{{role "system"}}` e `{{role "user"}}` estruturam a conversa, separando as instruções do sistema das perguntas da pessoa usuária.
* **Mídia**: o helper `{{media url=myImage}}` injeta dados multimodais (imagens, vídeos) diretamente no contexto do modelo.

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
1.  **Observabilidade**: a execução de cada flow gera automaticamente traces e métricas (latência, uso de tokens, taxa de sucesso) visualizáveis na Genkit Dev UI ou no Google Cloud Trace.
2.  **Segurança de Tipos (Type Safety)**: flows são estritamente tipados com schemas de entrada e saída, prevenindo erros em tempo de execução ao encadear múltiplas operações de IA.
3.  **Facilidade de Deploy (Deployability)**: flows são estritamente desacoplados da lógica de serving. Para fazer o deploy, basta envolvê-los com `genkit.Handler`, que converte um flow em um `http.Handler` padrão. Isso torna possível servi-los usando a biblioteca padrão ou qualquer framework web Go:

```go
    // Define a flow
    myFlow := genkit.DefineFlow(g, "myFlow", func(ctx context.Context, input string) (string, error) {
        return "Processed: " + input, nil
    })

    // Expose it as an HTTP handler
    http.HandleFunc("/myFlow", genkit.Handler(myFlow))
```



## Nano Banana Pro

O motor por trás da nossa restauração é o **Gemini 3 Pro Image**, carinhosamente conhecido como "Nano Banana Pro".

Ele representa um salto significativo em relação às gerações anteriores (e até mesmo aos modelos atuais "Flash"). Enquanto o Gemini 2.5 Flash é incrivelmente rápido e capaz de geração básica de imagens (`gemini-2.5-flash-image`), o **Nano Banana Pro** (`gemini-3-pro-image-preview`) foi construído para raciocínio multimodal profundo.

Ele não apenas "enxerga" pixels; ele compreende o contexto semântico. Sabe diferenciar um "arranhão no papel" de uma "cicatriz no rosto". Ele reconhece que uma cozinha dos anos 1950 provavelmente tinha piso de linóleo, e não madeira de lei moderna.

### Principais diferenças

*   **Flash (gemini-2.5-flash-image)**: otimizado para velocidade e custo. Ótimo para miniaturas (thumbnails) ou ilustrações simples. Resolução máxima de 1024x1024.
*   **Pro (gemini-3-pro-image-preview)**: otimizado para fidelidade e raciocínio. Suporta geração em **resolução 4K** nativa (até 4096px), o que é indispensável para restauração de fotos.

O modelo também aceita parâmetros `imageConfig` para fazer fine-tune no output:
*   `imageSize`: "4K" ou "2K".
*   `aspectRatio`: "16:9", "4:3", "1:1", etc.

Um detalhe importante a ser notado é que este modelo sempre retorna respostas intercaladas (interleaved) contendo tanto texto quanto imagens. Ao contrário de outros modelos de geração, o output exclusivo de imagem não é suportado. É por isso que nossa lógica de extração (que veremos abaixo) precisa ser flexível o suficiente para encontrar os dados da imagem dentro da mensagem de resposta multipartes.

**Nota:** No momento da publicação deste artigo, este modelo está disponível apenas na localização `global` no Vertex AI. Você deve configurar seu cliente Vertex AI de acordo.


## Conectando as partes

Agora, vamos ver como o GlowUp conecta essas peças. Usamos um **arquivo de prompt** para definir a persona de especialista em restauração e um **fluxo** (flow) para lidar com o processamento da imagem.

### O prompt

Usamos um arquivo `.prompt` para definir a configuração do nosso modelo e as instruções. Note como forçamos a resolução `4K` diretamente aqui, mantendo nosso código limpo.

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

Para suportar arquivos locais nativamente, usamos uma função auxiliar `fileToDataURI`. Esta função lê um arquivo local, detecta seu tipo MIME usando `http.DetectContentType` e o codifica em uma Data URI Base64 padrão que a API do Gemini espera. Isso é crítico para manter a fidelidade em diferentes formatos de escaneamento sem fixar extensões no código.

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


Como o Nano Banana Pro é inteligente o suficiente para inferir a proporção de tela (aspect ratio) da imagem de entrada, nós não precisamos de lógicas complexas para calculá-la e injetá-la. Nós apenas fornecemos os pixels e deixamos o modelo fazer o seu trabalho.


## Como executar

Se você tem um acervo de memórias desbotadas que servem como âncoras frágeis para a história da sua família, encorajo você a testar isso. É uma forma de recuperar esses momentos do tempo e dar a eles a clareza que merecem.

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

Embora o processo de restauração funcione, ele tem suas peculiaridades. Aqui estão alguns pontos que observei:
*   **Adesão às instruções:** Mesmo que o Nano Banana Pro seja um modelo de ponta, ele ocasionalmente ainda ignora uma instrução. Você pode achar que são necessárias algumas tentativas antes de obter o resultado desejado. Como não passei muito tempo refinando o prompt, certamente há oportunidades para otimizações adicionais lá.
*   **Modelos na Dev UI:** Há um bug no plugin `googlegenai` que faz com que ele não liste automaticamente os modelos disponíveis na Dev UI. Você ainda pode referenciar modelos por nome para registrá-los "dinamicamente", mas isso adiciona um pouco de atrito ao processo de experimentação (a versão em JS lida bem com isso). Eu abri [uma issue](https://github.com/firebase/genkit/issues/4783) e já há uma correção em vigor, mas se você estiver usando uma versão mais antiga, é algo a se ter em mente.


## Conclusões

Construir o GlowUp foi um experimento gratificante no uso de IA para me reconectar com meu passado em um nível emocional. Sei que há muito pessimismo por aí, mas esse é o tipo de aplicação que me deixa animada com a IA em primeiro lugar.

A foto que usei neste artigo está longe de ser o caso de uso mais dramático dessa tecnologia, mas já estou trabalhando na segunda parte deste artigo, onde vou levá-la para o próximo nível para me ajudar a reconstruir um dos meus jogos de cartas favoritos da infância.

A conclusão é que o potencial é ilimitado. Espero que isso inspire você a olhar para os seus próprios problemas de nicho — técnicos ou pessoais — e ver o que você pode construir para resolvê-los.

**Quer tentar construir o seu próprio?** Eu preparei um [codelab passo a passo](https://codelabs.developers.google.com/cloud-genkit-go-nano-banana?hl=en#0) onde você pode construir este exato app de restauração de fotos do zero.

Para mais detalhes, você também pode conferir a [documentação do Genkit](https://firebase.google.com/docs/genkit) e o [código-fonte do GlowUp](https://github.com/danicat/glowup).

**Bons códigos!**

Dani =^.^=
