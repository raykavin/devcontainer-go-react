# Template para criação de funcionalidades

Fluxo recomendado para implementar uma feature neste backend, do entendimento do requisito à validação final.

## 1. Entender o comportamento esperado

- Releia o requisito da tarefa. Se algo estiver ambíguo de um jeito que mudaria a implementação (ex.: verbo HTTP, formato de resposta, se precisa de transação), pergunte antes de escrever código — não assuma.
- Identifique em qual camada a mudança realmente pertence: `domain` (regra/invariante), `usecase` (orquestração), `adapter/inbound/http` (contrato de API), `adapter/outbound` (integração externa/banco).

## 2. Identificar contratos afetados

- A mudança altera uma rota HTTP existente (formato de request/response, status code, verbo)? Se sim, é uma mudança potencialmente incompatível — ver "Operações que exigem aprovação" em `AGENTS.md` antes de prosseguir.
- A mudança altera schema de banco (nova coluna, novo índice, novo relacionamento)? Verifique se o projeto usa `AutoMigrate` ou migrations versionadas (`instructions/database.md`) e siga o mecanismo já em uso.
- A mudança afeta um contrato consumido por outro serviço (ex.: payload de webhook, formato aceito por integração externa)? Trate como mudança entre serviços — exige mais cautela e possivelmente aprovação.

## 3. Localizar exemplos existentes

Antes de escrever qualquer código novo, procure um caso já implementado no mesmo repositório:
- Novo endpoint de listagem com filtro → veja um `Search`/`List` existente em `internal/usecase/` + `internal/adapter/outbound/postgres/`.
- Nova operação de ingestão idempotente (upload, importação) → veja se já existe um padrão find-or-create com hash de conteúdo no projeto.
- Novo usecase multi-etapa que grava em mais de uma tabela → veja se já existe um `TransactionRunner`/Unit-of-Work.
- Novo endpoint autenticado → veja como `gobox/oidcauth`+`middlewares` já são usados no mesmo repositório antes de tocar em autenticação.
- Antes de criar uma abstração nova (helper genérico, wrapper, cliente HTTP): verifique se já existe em `github.com/raykavin/gobox` (ver tabela em `instructions/go.md`).
- Antes de adicionar uma dependência nova ao `go.mod`: verifique se `gobox` ou uma dependência já presente resolve o mesmo problema.

## 4. Definir regras de domínio

- Regras de negócio/invariantes vivem em `internal/domain` (ou como método de entidade). Não coloque regra de negócio em `handler` ou em `adapter/outbound`.
- Se a regra depende de estado externo (outro serviço, config), isso é orquestração de `usecase`, não regra pura de `domain`.

## 5. Avaliar segurança e autorização

- Toda rota nova precisa de: middleware de autenticação (`gobox/oidcauth`) + middleware de autorização com permissão específica da operação (não reaproveite uma permissão genérica de outro endpoint).
- Se a operação processa dado sensível (CPF, dado bancário, PII), confirme que não será logado sem redação, especialmente em nível debug.
- Ver checklist completo em `instructions/security.md`.

## 6. Avaliar persistência e migrations

- Nova entidade/coluna: siga o mecanismo de schema já usado no repositório (`AutoMigrate` ou migration SQL versionada — `instructions/database.md`).
- Nova operação de escrita multi-tabela: use (ou crie, se ainda não existir no repositório) um `TransactionRunner`.
- Nova consulta com relacionamento: use `Preload`, não `Joins` manual (reserve `Joins` para filtro).
- Se a alteração toca em uma migration já aplicada em produção: **não a edite** — crie uma nova.

## 7. Implementar a menor mudança coerente

- Um usecase por operação (`internal/usecase/<contexto>/<operação>.go`).
- Não refatore código não relacionado à tarefa "enquanto está ali" — se notar uma oportunidade de melhoria fora do escopo, relate ao final em vez de executar sem pedir.
- Reaproveite o envelope de resposta, paginação, logger e autenticação já em uso no repositório.

## 8. Criar ou atualizar testes

- Endpoint novo: teste de autorização, sucesso, entrada inválida, recurso inexistente.
- Usecase novo: caminho feliz + pelo menos uma condição de erro de negócio.
- Código concorrente (fila, worker, workflow): sempre teste, mesmo que pareça simples.
- Use mock funcional escrito à mão ou SQLite in-memory para transação real (ver `instructions/testing.md`).

## 9. Executar validações

```sh
go build ./...
go vet ./...
golangci-lint run     # se .golangci.yml existir no repositório
go test -v -race ./...
```
Se algum passo não puder rodar, declare isso explicitamente no relatório final.

## 10. Atualizar documentação

- Anotações Swagger (`@Summary/@Router/@Success/@Failure`) sincronizadas com o comportamento real — regenere via `make swagger`.
- README (seção relevante: Arquitetura/Configuração/API) se o comportamento documentado mudou.
- `.claude/`/`AGENTS.md` local do repositório, se existir e se a mudança afetar o que ele descreve.

## 11. Relatar

No relatório final da tarefa, inclua: o que mudou (arquivos/rotas/schema), por que (ligação com o requisito), quais decisões você tomou e por quê (ex.: qual verbo HTTP escolheu e por quê, se seguiu `AutoMigrate` ou migration versionada), quais riscos permanecem, e como validou (comandos executados e resultado — inclusive falhas ou passos pulados). Se algo tocou em uma das operações que exigem aprovação (`AGENTS.md`), confirme que a aprovação foi obtida antes de executar, não depois.
