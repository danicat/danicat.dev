---
title: "Dominando Hooks em Agentes de Codificação"
date: 2026-05-21T11:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["antigravity", "agy-cli", "agentic-coding"]
summary: "Aprenda a usar hooks de agentes para incorporar as melhores práticas de engenharia em seus loops de codificação agentivos."
heroStyle: "big"
---

As capacidades dos agentes de codificação estão avançando rapidamente. Meu primeiro contato com eles foi há cerca de um ano, logo após eu entrar no Google. Naquela época, a grande novidade era o Model Context Protocol (MCP), uma tecnologia flexível que substitui implementações ad-hoc de ferramentas por soluções portáteis (entre outras coisas).

Avançando doze meses, e hoje a maioria das pessoas parece ter migrado dos MCPs para as Agent Skills (Habilidades de Agentes) como a próxima grande tendência. Ao introduzir a "revelação progressiva", as skills permitem um uso mais eficiente do contexto, o que resulta em uma melhor economia de tokens no geral. Com o aumento dos custos de inferência, não é surpresa que as skills tenham se tornado tão populares.

Tanto o MCP quanto as skills têm suas próprias comunidades de fãs, mas há outro padrão de agentes de codificação que surgiu nesse período e não é mencionado com tanta frequência quanto seus equivalentes mais populares: os hooks.

Enquanto MCPs e skills focam em estender as capacidades agentivas (adicionando ferramentas e conhecimento), os hooks operam em um nível diferente, oferecendo mais controle sobre o loop do agente e o processo de desenvolvimento como um todo.

## Hooks de agentes explicados

O nome hook pode não soar familiar à primeira vista, mas os hooks são simplesmente "callbacks" — procedimentos que são chamados em momentos específicos durante o ciclo de vida de processamento do agente.

Os hooks possuem três componentes:
- Um evento gatilho (trigger event): quando o hook será chamado. Normalmente composto por uma fase de execução (pre ou post) e um contexto, como uma chamada de ferramenta ou invocação de modelo. Por exemplo, no Antigravity, temos eventos como `PreToolUse` e `PostInvocation`. Alguns motores também oferecem suporte a um hook `Stop` (na terminação do agente) e um hook `PreCompact` (antes da autocompactação, que é exclusivo para implementações baseadas no Claude).
- Uma condição ou filtro: uma expressão regular baseada no evento gatilho. Por exemplo, em uma chamada de ferramenta, o filtro pode ser o nome da ferramenta e pode incluir seus argumentos. Por exemplo, é possível criar um hook para a chamada de ferramenta `run_command(git)`.
- Um procedimento: o corpo do hook, seja na forma de um script shell ou comando. O procedimento pode ser usado para permitir ou negar uma operação, para sobrescrever completamente as chamadas de modelo ou ferramenta, e para produzir efeitos colaterais como logging ou telemetria.

## Quando usar hooks

Os hooks interceptam o ciclo de vida do agente em momentos específicos para injetar comandos ou scripts personalizados. Ao interceptar no momento certo, você pode controlar o fluxo de operação e adicionar resultados determinísticos ao que, de outra forma, seria um processo majoritariamente não determinístico.

Por exemplo, desenvolvedores frequentemente tentam impor diretrizes de codificação por meio de prompts do sistema ou de um arquivo AGENTS.md (ou similar). No entanto, diretrizes baseadas em prompts não oferecem garantias de execução devido à natureza não determinística dos grandes modelos de linguagem (LLMs): o exato mesmo prompt pode produzir resultados diferentes, e os agentes podem ignorar seletivamente partes do prompt.

Usando hooks em vez de prompts, você pode forçar uma ação específica. Digamos, por exemplo, que você queira garantir que seu agente sempre execute uma ferramenta de análise estática no código (linter) após cada edição, para assegurar que o código esteja sempre limpo e validado. Se você adicionar "sempre execute o linter após edições" aos seus prompts, o agente terá a autonomia de decidir se vai rodar o linter ou não, podendo ignorar a validação completamente se achar que a edição foi "trivial". Mas se você criar um hook — neste caso, um `PostToolUse` filtrando pela ferramenta de edição de arquivos —, você garante de forma determinística que sua ferramenta de análise estática será executada após cada alteração.

Ao interceptar esses ciclos de vida, podemos implementar vários padrões poderosos para controlar o comportamento do agente, coletar métricas ou manter os fluxos de trabalho seguros.

### Direcionar o agente para ferramentas especializadas

Os hooks são úteis em muitos cenários, mas meu caso de uso favorito é colocar barreiras de proteção (guardrails) ao redor do agente, ou, como às vezes gosto de dizer: "limitar a autonomia do agente".

Podemos implementar isso emparelhando um hook `PreToolUse` com um script que nega o acesso à ferramenta e retorna uma "dica de direcionamento" (steering hint) para o agente de codificação. Essa dica conterá as instruções que você deseja que ele execute em vez disso. Por exemplo, suponha que você queira impedir o agente de usar comandos shell para ler arquivos Go; uma dica de direcionamento poderia ser assim: "Tool call blocked - run_command(cat): do not use 'cat' for reading *.go files, use 'smart_read' instead".

### Coletar dados de telemetria

O sistema de hooks também é um ótimo lugar para colocar seus coletores de telemetria e geradores de logs, pois oferece uma excelente visibilidade do funcionamento interno do agente.

### Interceptar prompts maliciosos

Um hook `PreInvocation` pode interceptar prompts recebidos e avaliá-los em relação a heurísticas de segurança ou modelos classificadores mais leves. Se um prompt parecer uma tentativa de jailbreak, o hook pode bloquear a requisição imediatamente, protegendo os sistemas de backend antes que cheguem ao loop de execução.

### Prevenir vazamento de credenciais

Desenvolvedores às vezes colam acidentalmente arquivos env ou credenciais em arquivos ativos que o agente de codificação lê. Um hook `PostToolUse` monitorando a leitura de arquivos — ou um hook `PreInvocation` escaneando os payloads enviados para o LLM — atua como uma barreira confiável de Prevenção de Perda de Dados (DLP). Se o hook detectar strings que correspondem a chaves de alta entropia ou formatos padrão de API, ele pode redigir os segredos dinamicamente ou abortar a execução para manter as credenciais seguras.

### Gerenciando memórias

Os agentes são tipicamente stateless (sem estado), a menos que estejam conectados a um sistema de memória externa, como o [Agent Platform Memory Bank](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank) ou o [MemPalace](https://github.com/mempalace/mempalace).

Uma maneira de adicionar capacidades de memória aos agentes é registrando a memorização e a recuperação como ferramentas, mas, ao fazer isso, dependemos de o agente tomar explicitamente a decisão de chamar as ferramentas correspondentes.

O sistema de hooks permite que você automatize a persistência e recuperação de memória. Você pode conectar a [geração de memórias](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/generate-memories#triggering-memory-generation) ao final de uma sessão (usando um hook `Stop`) ou após um certo número de turnos (monitorando o número da etapa ou o número de invocações do modelo).

Por outro lado, a recuperação de memória pode ser feita automaticamente no início da sessão e antes de invocar os modelos (por exemplo, com um hook `PreInvocation`). No Agent Platform Memory Bank, você pode recuperar memórias por [escopo](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#retrieve-all) (por exemplo, o escopo pode ser um ID de usuário) ou [similaridade](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#similarity-search) (baseado em uma consulta). Isso é essencialmente uma geração aumentada por recuperação (RAG) baseada em memória.

## Configurando hooks no Antigravity

Embora diferentes motores de agentes implementem seu próprio vocabulário para callbacks, nesta seção focaremos especificamente no **dialeto Antigravity** de hooks.

Para uma visão completa das especificações, consulte a [Documentação de Hooks do Antigravity](https://antigravity.google/docs/hooks) oficial.

O Antigravity procura por um arquivo `hooks.json` dentro do diretório `.agents/` do seu espaço de trabalho (ou globalmente no diretório do usuário em `~/.gemini/config/hooks.json`).

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

As entradas para esses hooks são entregues via `stdin` como um objeto JSON, contendo contexto como os argumentos da ferramenta (`toolCall.args`), `workspacePaths` ativos e o caminho do arquivo do log da sessão atual (`transcriptPath`). Seus scripts podem avaliar essas informações, realizar computações e exibir uma resposta JSON no `stdout` informando ao Antigravity se deve autorizar (`"allow"`), negar (`"deny"`) ou perguntar (`"ask"`) pela confirmação do usuário.

Por exemplo, veja como você pode escrever um script Python simples (`steer-go-reads.py`) para analisar esse payload recebido e direcionar o agente:

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

Os agentes estão se tornando cada vez mais inteligentes e rápidos, permitindo que produzamos código em uma velocidade sem precedentes. Mas velocidade sem controle é a receita para o desastre. Costumo usar esta analogia em minhas palestras: se você gosta de carros rápidos, a coisa mais importante com a qual deve se preocupar não é o motor, mas os freios. Se os freios forem menos potentes que o motor, você não conseguirá parar e sua segurança estará comprometida.

O mesmo raciocínio deve ser aplicado aos agentes de codificação. Se quiser escrever código rapidamente, você precisa de um sistema de controle robusto que garanta que você não está sacrificando a qualidade e introduzindo bugs — porque se fizer isso, sua aplicação terá grandes problemas mais cedo ou mais tarde.

Os hooks são um excelente lugar para implementar as salvaguardas que podem nos devolver o controle, unindo a autonomia da IA à engenharia de software robusta. Como li recentemente [neste artigo de Joe Bertolami](https://venturebeat.com/technology/agentic-ai-solved-coding-and-exposed-every-other-problem-in-software-engineering): "escrever código nunca foi o limitador de velocidade". Não vamos esquecer décadas de melhores práticas de engenharia; devemos equipar nossos agentes com as ferramentas certas para o trabalho, permitindo que desfrutemos plenamente da moderna experiência de desenvolvimento agentivo.
