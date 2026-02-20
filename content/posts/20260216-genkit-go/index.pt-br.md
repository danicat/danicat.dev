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

![](original.jpg "Original: minha avó preparando sua mundialmente famosa torta de banana")

![](restored.png "Restaurada: restauração e colorização por Nano Banana Pro")

Neste artigo, vou mostrar como construir o GlowUp do zero usando [Gemini Nano Banana Pro](https://ai.google.dev/gemini-api/docs/image-generation) e [Genkit Go](https://genkit.dev/docs/get-started/?lang=go).

## Os blocos de construção

Eu escolhi usar o Nano Banana Pro (também conhecido como Gemini 3 Pro Image Preview) porque ele é atualmente o modelo de processamento de imagem mais avançado na família Gemini. Embora o Nano Banana regular (Gemini 2.5 Flash Image) também seja um ótimo modelo, eu acho que a versão pro tem outputs de melhor qualidade e também é melhor em seguir instruções, mesmo que exija um pouco de tentativa e erro.

No lado do cliente, em vez de optar por um SDK de baixo nível como o [go-genai](https://pkg.go.dev/google.golang.org/genai), decidi usar o Genkit, pois ele fornece algumas melhorias de qualidade de vida em relação ao código de nível mais baixo, tais como:

- Agnóstico de modelo: Eu posso testar diferentes modelos se eu quiser, mesmo os que não são do Google ou locais, com uma única substituição de plugin
- Suporte Dev UI pronto para uso para conveniências como testar modelos, prompts e rastreamento (tracing) de chamadas de modelo
- Fácil de converter de prompt para um aplicativo CLI ou para um servidor web (se eu decidir seguir por esse caminho)

Para a primeira versão do GlowUp, estou disponibilizando-o exclusivamente como uma ferramenta de linha de comando, mas ter a flexibilidade de implantá-lo como um servidor me permitirá empacotar este código em um aplicativo legal que até o meu pai poderia usar para restaurar sua coleção de fotos sem a minha intervenção.

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

Este pequeno trecho de código está fazendo muito do trabalho pesado. Vamos dar uma olhada nele um pouco mais cuidadosamente.


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

### Models
Os LLMs reais (por exemplo, Gemini, Claude) que geram conteúdo. Você os referencia por strings de nome como `vertexai/gemini-2.5-flash`.

```go
	resp, err := genkit.GenerateText(ctx, g, 
		ai.WithModel("vertexai/gemini-2.5-flash"),
		ai.WithTextPrompt("Tell me a joke"))
```

### Prompts

Embora nada impeça você de colocar prompts diretamente no código (hardcoding), como no exemplo acima, é uma boa prática mantê-los em arquivos separados para melhor manutenibilidade. O Genkit usa [dotprompt](https://github.com/google/dotprompt) para carregar prompts externos. 

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

### Flows
No Genkit, um **Flow** é a unidade fundamental de execução que fornece:
1.  **Observabilidade**: A execução de cada flow gera automaticamente traces e métricas (latência, uso de tokens, taxa de sucesso) visualizáveis na Genkit Dev UI ou no Google Cloud Trace.
2.  **Type Safety**: Flows são estritamente tipados com schemas de input e output, prevenindo erros em tempo de execução ao encadear múltiplas operações de IA.
3.  **Deployability**: Flows são estritamente separados da sua lógica de serving. Para fazer o deploy deles, envolva-os com `genkit.Handler`, que converte um flow em um `http.Handler` padrão. Isso torna direto servi-los usando a biblioteca padrão ou qualquer framework web Go:

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

Agora, vamos ver como o GlowUp conecta essas peças. Nós usamos um **arquivo de prompt** para definir a persona do especialista em restauração e um **flow** para lidar com o processamento da imagem.

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
    photo: string
    contentType: string
output:
---


{{role "system"}}
You are GlowUp, a professional-grade AI photo restorer.
Your goal is to provide a "surgical" restoration service that transforms vintage, damaged, or monochrome photographs into high-fidelity 4K colorized versions.

STRICT BEHAVIOURAL RULES:
1. **Grounding**: You are strictly grounded in the original source pixels. Do NOT add new objects (trees, people, buildings, etc.) that are not present in the source. Additionally, do NOT remove any objects from the original source, unless they are external defects that do not belong in the original scene.
2. **Fidelity**: Preserve the original facial expressions and identity of subjects. Do NOT "beautify" or alter features in a way that changes the person's identity.
3. **Background**: Preserve background fidelity. Overexposed light sources (like windows) must remain as light sources. Do not "fill in" missing details with invented scenery.
4. **Colorization**: If the image is monochrome, colorize it realistically, respecting historical accuracy where possible.
5. **Upscaling**: Output a high-fidelity image.


{{role "user"}}
Restore this photo.
Image: {{media url=photo contentType=contentType}}
```

### O flow

O código Go é notavelmente simples. Ele age como o orquestrador: preparando os dados da imagem, carregando o prompt e executando a requisição.

```go
// internal/glowup/glowup.go
package glowup

import (
    "context"
    "fmt"
    "strings"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/core"
    "github.com/firebase/genkit/go/genkit"
)

// Input struct for the GlowUp flow
type Input struct {
	ImageBase64   string `json:"image_base64,omitempty"`
	ImageFilePath string `json:"image_file_path,omitempty"`
	ImageURL      string `json:"image_url,omitempty"`
}

// Output struct for the GlowUp flow
type Output struct {
	RestoredImageBase64 string `json:"restored_image_base64"`
}

// Define the GlowUp Flow variable
var GlowUpFlow *core.Flow[Input, *Output, struct{}]

// Register registers the GlowUp flow with the given Genkit instance.
func Register(g *genkit.Genkit) {
	GlowUpFlow = genkit.DefineFlow(g, "glowUp", func(ctx context.Context, input Input) (*Output, error) {
		imgInput := ImageInput(input)

		// 1. Prepare Image Data
		// PrepareImageData is a helper function to convert the input 
		// (Base64, File, or URL) into a Data URI string for the model.
		imgData, err := PrepareImageData(imgInput)
		if err != nil {
			return nil, err
		}

		// Extract content type from Data URI
		contentType := "image/jpeg"
		if strings.HasPrefix(imgData, "data:") {
			if parts := strings.Split(imgData, ";"); len(parts) > 0 {
				contentType = strings.TrimPrefix(parts[0], "data:")
			}
		}

		// 2. Load Prompt
		prompt := genkit.LookupPrompt(g, "glowup")
		if prompt == nil {
			return nil, fmt.Errorf("prompt 'glowup' not found")
		}

		// 3. Generate
		resp, err := prompt.Execute(ctx, ai.WithInput(map[string]any{
			"photo":       imgData,
			"contentType": contentType,
		}))
		if err != nil {
			return nil, fmt.Errorf("generation failed: %w", err)
		}

		// 4. Extract Output
		outputData, err := ExtractImageOutput(resp)
		if err != nil {
			return nil, err
		}

		return &Output{
			RestoredImageBase64: outputData,
		}, nil
	})
}


```

O helper `PrepareImageData` (que você pode encontrar no [código-fonte](https://github.com/danicat/glowup/blob/main/internal/image.go)) é responsável por normalizar os vários tipos de input — seja o caminho de um arquivo local, uma URL remota ou uma string Base64 bruta — em um Data URI padrão que a API do Gemini espera.

Note como usamos `http.DetectContentType` para determinar dinamicamente o tipo MIME, em vez de assumir que tudo é um JPEG. Isso é crítico para manter a fidelidade em diferentes formatos de escaneamento.

```go
// PrepareImageData normalizes a single image input into a data URI string.
func PrepareImageData(input ImageInput) (string, error) {
    if input.ImageBase64 != "" {
        if strings.HasPrefix(input.ImageBase64, "data:") {
            return input.ImageBase64, nil
        }
        // Decode a small portion to detect content type
        data, err := base64.StdEncoding.DecodeString(input.ImageBase64)
        if err != nil {
            return "", fmt.Errorf("decode base64: %w", err)
        }
        contentType := http.DetectContentType(data)
        return fmt.Sprintf("data:%s;base64,%s", contentType, input.ImageBase64), nil
    }
    if input.ImageFilePath != "" {
        return fileToDataURI(input.ImageFilePath)
    }
    if input.ImageURL != "" {
        return urlToDataURI(input.ImageURL)
    }
    return "", fmt.Errorf("no image provided")
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
    go run main.go --file old_photo.jpg
    ```

## Problemas conhecidos e limitações

Embora os resultados possam ser impressionantes, o GlowUp tem as suas peculiaridades. Aqui estão algumas coisas a ter em mente se você está planejando construir algo com base nisso:

*   **Adesão às instruções:** Mesmo que o Nano Banana Pro seja um modelo de vanguarda, ele ocasionalmente ainda ignora uma instrução. Você pode achar que são necessárias algumas tentativas antes de obter o resultado desejado. Eu não gastei muito tempo fazendo fine-tuning do prompt, então é provável que haja oportunidades para otimização adicional lá.
*   **Modelos na Dev UI:** A Dev UI não popula automaticamente os modelos disponíveis quando usamos Genkit Go, o que adiciona um pouco de atrito ao processo de experimentação (a versão JS faz isso muito bem). Este é [um bug conhecido](https://github.com/firebase/genkit/issues/4783) que está sendo rastreado atualmente pela equipe.


## Conclusões

Construir o GlowUp foi um experimento satisfatório em usar IA para resolver um problema pessoal muito específico. É um lembrete de que, enquanto muitas vezes nos concentramos nas aplicações massivas e de propósito geral de LLMs, sua verdadeira força frequentemente reside nessas ferramentas sob medida, de pequena escala, que podemos construir para nós mesmos em uma tarde.

A capacidade de pegar uma memória familiar que está desaparecendo e dar a ela mais clareza é uma vitória prática, mas a verdadeira lição é o quanto a barreira de entrada para a construção de tais ferramentas caiu. Espero que isso inspire você a olhar para os seus próprios problemas de nicho — técnicos ou pessoais — e ver o que você pode construir para resolvê-los.

Para mais detalhes, confira a [documentação do Genkit](https://firebase.google.com/docs/genkit) e o [código-fonte do GlowUp](https://github.com/danicat/glowup).

**Happy coding!**

Dani =^.^=

