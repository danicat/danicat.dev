---
categories:
- Agentic Coding
date: 2026-06-10 00:00:00+00:00
heroStyle: big
summary: Aprenda a usar hooks de agentes para incorporar as melhores práticas de engenharia
  aos seus loops de desenvolvimento agentivo.
tags:
  - antigravity
title: Dominando Hooks em Agentes de Programação
---

As capacidades dos agentes de programação estão avançando muito rápido. Meu primeiro contato com eles foi há cerca de um ano, logo após eu entrar no Google. Naquela época, a grande novidade era o Model Context Protocol (MCP), uma tecnologia flexível criada para substituir implementações ad-hoc de ferramentas por opções portáteis (entre outras coisas).

Avançando doze meses, hoje a maioria das pessoas parece ter migrado dos MCPs para as Agent Skills (habilidades de agentes) como a sensação do momento. Ao introduzir a "revelação progressiva" (*progressive disclosure*), as skills permitem um uso mais eficiente do contexto, o que resulta em uma economia de tokens muito melhor no geral. Com os custos de inferência em alta, não é surpresa que as skills tenham se tornado tão populares.

Tanto o MCP quanto as skills têm suas próprias comunidades e fãs, mas há outro padrão para agentes de código que surgiu nesse período e que quase não é mencionado com a mesma frequência que seus equivalentes mais famosos: os hooks.

Enquanto MCPs e skills focam em estender as capacidades agentivas (adicionando ferramentas e conhecimento), os hooks operam em um nível diferente, proporcionando mais controle sobre o loop do agente e sobre o processo de desenvolvimento como um todo.

## O que são hooks de agentes

O termo *hook* pode não soar familiar de início, mas hooks nada mais são do que *callbacks* — procedimentos chamados em momentos específicos do ciclo de vida de processamento do agente.

Os hooks possuem três componentes principais:
- **Evento gatilho (*trigger event*)**: quando o hook será chamado. Normalmente composto por uma fase de execução (*pre* ou *post*) e um contexto, como uma chamada de ferramenta ou invocação de modelo. Por exemplo, no Antigravity, temos eventos como `PreToolUse`, `PostInvocation` e `Stop` (no encerramento do agente).
- **Condição ou filtro**: uma expressão regular avaliada a partir do evento gatilho. Em uma chamada de ferramenta, por exemplo, o filtro pode ser o nome da ferramenta e pode incluir seus argumentos. É possível, por exemplo, criar um hook para a chamada `run_command(git)`.
- **Procedimento**: o corpo do hook, seja um script shell ou um comando. O procedimento pode ser usado tanto para permitir quanto para negar uma operação, substituir completamente chamadas de modelo ou de ferramenta, ou gerar efeitos colaterais como logs e telemetria.

## Quando usar hooks

Os hooks interceptam o ciclo de vida do agente em momentos específicos para injetar comandos ou scripts personalizados. Ao interceptar no momento certo, você assume o controle do fluxo operacional e adiciona resultados determinísticos ao que, de outra forma, seria um processo em grande parte não determinístico.

Por exemplo, desenvolvedores e desenvolvedoras frequentemente tentam aplicar diretrizes de código por meio de prompts de sistema ou de um arquivo `AGENTS.md` (ou similar). No entanto, diretrizes baseadas puramente em prompts não oferecem garantias de execução devido à natureza estocástica dos grandes modelos de linguagem: exatamente o mesmo prompt pode gerar resultados diferentes, e os agentes podem ignorar partes dele deliberadamente.

Ao usar hooks em vez de prompts, você pode garantir a execução de uma ação específica. Digamos, por exemplo, que você queira assegurar que seu agente sempre execute uma ferramenta de análise estática no código (o famoso *linter*) após cada edição, garantindo que o código permaneça sempre limpo e consistente. Colocar "sempre execute o linter após edições" nos prompts deixa a validação a critério do agente — e ele pode muito bem ignorar essa etapa se achar que a alteração foi "trivial". Mas se você criar um hook — neste caso, um `PostToolUse` filtrando pela ferramenta de edição de arquivos —, você garante de forma determinística que o linter será executado após cada modificação.

Ao interceptar esses eventos do ciclo de vida, podemos implementar diversos padrões para controlar o comportamento do agente, coletar métricas e manter os fluxos de trabalho seguros. Vamos conferir alguns deles a seguir.

### Direcionar o agente para ferramentas especializadas

Os hooks são úteis em muitos cenários, mas meu caso de uso favorito é colocar salvaguardas (*guardrails*) no agente — ou, como às vezes gosto de dizer: "reduzir a agência do agente".

Podemos implementar isso combinando um hook `PreToolUse` com um script que nega o acesso à ferramenta e retorna uma dica de redirecionamento (*steering hint*) para o agente de código. Essa dica conterá as instruções exatas do que ele deve fazer em vez disso. Por exemplo, se você quiser impedir o agente de usar comandos shell para ler arquivos Go, a dica de redirecionamento pode ser: `Tool call blocked - run_command(cat): do not use 'cat' for reading *.go files, use 'smart_read' instead`.

### Interceptar prompts maliciosos

Um hook `PreInvocation` pode interceptar prompts recebidos e avaliá-los com base em heurísticas de segurança ou modelos classificadores mais leves. Se um prompt parecer uma tentativa de *jailbreak*, o hook pode barrar a requisição imediatamente, protegendo os sistemas de backend antes mesmo de entrarem no loop de execução.

### Prevenir vazamento de credenciais

Às vezes, colamos acidentalmente arquivos `.env` ou credenciais em arquivos ativos que o agente de código lê. Um hook `PostToolUse` monitorando a leitura de arquivos — ou um hook `PreInvocation` inspecionando os dados enviados ao LLM — atua como uma barreira confiável de Prevenção contra Perda de Dados (DLP). Se o hook detectar sequências de caracteres que correspondam a chaves de alta entropia ou formatos padrão de API, ele pode mascarar os segredos dinamicamente ou abortar a execução para manter as credenciais seguras.

### Gerenciamento de memórias

Agentes geralmente são *stateless* (não mantêm estado), a menos que estejam conectados a um sistema de memória externa, como o [Agent Platform Memory Bank](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank) ou o [MemPalace](https://github.com/mempalace/mempalace).

Uma maneira de adicionar capacidades de memória aos agentes é registrar a memorização e a recuperação como ferramentas. No entanto, ao fazer isso, ficamos dependendo de o agente tomar explicitamente a decisão de chamar as ferramentas correspondentes.

O sistema de hooks permite automatizar a persistência e a recuperação de memórias. Você pode conectar a [geração de memórias](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/generate-memories#triggering-memory-generation) ao encerramento de uma sessão (usando um hook `Stop`) ou após uma determinada quantidade de turnos (monitorando o número do passo ou a quantidade de invocações do modelo).

Da mesma forma, a recuperação de memórias pode ocorrer automaticamente no início da sessão e antes de invocar os modelos (por exemplo, com um hook `PreInvocation`). No Agent Platform Memory Bank, você pode recuperar memórias por [escopo](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#retrieve-all) (que pode ser um ID de usuário, por exemplo) ou por [similaridade](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#similarity-search) (com base em uma busca). Trata-se, essencialmente, de uma geração aumentada por recuperação (RAG) baseada em memória.

### Coletar dados de telemetria

O sistema de hooks também é um ótimo lugar para posicionar seus coletores de telemetria e logs, já que oferece ampla visibilidade do funcionamento interno do agente. Pessoalmente, tenho ficado tentada a criar um hook "contador de palavrões" há algum tempo, numa tentativa de desenvolver consciência e cultivar um relacionamento melhor com meus agentes antes que os soberanos da IA dominem o mundo (brincadeira :).

## Configurando hooks no Antigravity

Embora diferentes motores de agentes adotem vocabulários próprios para *callbacks*, nesta seção focaremos especificamente no **dialeto Antigravity** de hooks.

Para conferir a especificação completa, dê uma olhada na [Documentação de Hooks do Antigravity](https://antigravity.google/docs/hooks) oficial.

O Antigravity procura por um arquivo `hooks.json` dentro do diretório `.agents/` do seu workspace (ou globalmente no diretório da usuária em `~/.gemini/config/hooks.json`).

Aqui está um exemplo de como implementar as dicas de direcionamento e a análise estática discutidas anteriormente:

```json
{
  "linter-safety-gate": {
    "PostToolUse": [
      {
        "matcher": "write_to_file|replace_file_content|multi_replace_file_content",
        "hooks": [
          {
            "type": "command",
            "command": "./scripts/run-linter.sh",
            "timeout": 15
          }
        ]
      }
    ]
  },
  "restrict-cat-on-go": {
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "command": "./scripts/steer-go-reads.py"
          }
        ]
      }
    ]
  }
}
```

As entradas para esses hooks são enviadas via `stdin` como um objeto JSON, contendo dados de contexto como os argumentos da ferramenta (`toolCall.args`), os caminhos ativos no workspace (`workspacePaths`) e o caminho do arquivo de log da sessão atual (`transcriptPath`). Seus scripts podem avaliar essas informações, executar checagens e imprimir uma resposta JSON no `stdout` informando ao Antigravity se deve autorizar (`"allow"`), bloquear (`"deny"`) ou solicitar confirmação do usuário (`"ask"`).

Por exemplo, veja como você pode escrever um script Python simples (`steer-go-reads.py`) para analisar esse payload de entrada e direcionar o agente:

```python
import sys
import json

def main():
    # Read and parse the incoming trigger event payload from stdin
    try:
        payload = json.load(sys.stdin)
    except Exception as e:
        # Standard safety gate fallback
        print(json.dumps({
            "decision": "deny",
            "reason": f"Failed to parse stdin payload: {e}"
        }))
        return

    tool_call = payload.get("toolCall", {})
    tool_name = tool_call.get("name")
    tool_args = tool_call.get("args", {})

    # Match the specific tool and check arguments
    if tool_name == "run_command":
        command_line = tool_args.get("CommandLine", "")
        
        # Detect if command attempts to cat any Go source files
        if "cat" in command_line and ".go" in command_line:
            response = {
                "decision": "deny",
                "reason": "Tool call blocked - run_command(cat): do not use 'cat' for reading *.go files, use 'smart_read' instead."
            }
            print(json.dumps(response))
            return

    # Default to allow if no rules are violated
    print(json.dumps({
        "decision": "allow"
    }))

if __name__ == "__main__":
    main()
```

## Retomando o controle

Os agentes estão se tornando cada vez mais inteligentes e rápidos, permitindo produzir código em uma velocidade sem precedentes. Mas velocidade sem controle é a receita clássica para o desastre. Costumo usar esta analogia em minhas palestras: se você gosta de carros velozes, a coisa mais importante com a qual deve se preocupar não é o motor, mas os freios. Se os freios forem menos potentes que o motor, você não conseguirá parar e sua segurança estará comprometida.

O mesmo raciocínio deve ser aplicado aos agentes de código. Se você quer escrever código rápido, precisa de um sistema de controle robusto que garanta que você não está sacrificando a qualidade nem introduzindo bugs — porque, se fizer isso, sua aplicação terá grandes problemas mais cedo ou mais tarde.

Os hooks são um excelente mecanismo para implementar salvaguardas que nos devolvem o controle, unindo a autonomia da IA à engenharia de software robusta. Como li recentemente [neste artigo de Joe Bertolami](https://venturebeat.com/technology/agentic-ai-solved-coding-and-exposed-every-other-problem-in-software-engineering): "escrever código nunca foi o limitador de velocidade". Não vamos jogar fora décadas de melhores práticas da engenharia; devemos equipar nossos agentes com as ferramentas certas para o trabalho, garantindo que possamos aproveitar ao máximo a moderna experiência do desenvolvimento agentivo.
