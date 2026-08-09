# Padrão de desenvolvimento backend (Go)

Padrão de referência para serviços backend Go construídos a partir deste template.

## Princípios de engenharia

1. Simplicidade e legibilidade acima de abstração especulativa — composição manual e explícita, sem framework de DI.
2. Reuso via `gobox` (`github.com/raykavin/gobox`) é a intenção declarada — antes de escrever logger, config loader, envelope HTTP, paginação, cliente OIDC ou wrapper de banco, verifique se já existe em `gobox`.
3. Menor mudança coerente: os usecases são um arquivo por operação (`create.go`, `update.go`, `find_or_create.go`) — não agrupe múltiplas operações não relacionadas no mesmo arquivo/função.

## Arquitetura padrão

Hexagonal / ports-and-adapters, com esta divisão de pastas em `internal/`:

```
internal/
  domain/    entidades GORM (BaseModel: ID/CreatedAt/UpdatedAt/DeletedAt), value objects, regras de invariante — SEM import de adapter/infra/config
  port/      interfaces: contratos de repositório e de serviços externos, consumidos por usecase
  usecase/   1 arquivo por operação de negócio; depende só de port, nunca de adapter/outbound diretamente
  adapter/
    inbound/http/   handler (bind+validação+delega), middleware, response (envelope), router
    outbound/<x>/    implementação concreta de port (postgres, integrações externas)
  app/       DI manual: bootstrap.go monta tudo e devolve as dependências prontas para cmd/*/main.go
  config/    carregamento e validação de configuração por entrypoint (api, worker)
  dto/       objetos de request/response da camada HTTP
```

**Regras de dependência (proibidas/permitidas):**
- `domain` não importa nada de `adapter`, `infra`, `config` ou `app`.
- `usecase` importa `domain` e `port`; nunca importa `adapter/outbound` diretamente (só a interface em `port`).
- `adapter/outbound/*` implementa as interfaces de `port`; é o único lugar onde GORM/drivers de banco/clientes HTTP de terceiros aparecem.
- `adapter/inbound/http/handler` não implementa regra de negócio — valida entrada, chama 1 usecase, formata a saída.
- A montagem (wiring) de tudo isso é centralizada em `internal/app/` (ou `internal/platform/`), chamada uma única vez em cada `cmd/*/main.go`.

**Injeção de dependência: manual, sem framework.** Não introduza Uber Fx, Wire ou equivalente salvo decisão explícita.

**Múltiplos binários em `cmd/`.** Cada processo com ciclo de vida distinto (API HTTP, worker assíncrono, migrator, CLI de diagnóstico) é um binário próprio em `cmd/<nome>/main.go`, todos consumindo o mesmo `internal/app`. Não coloque lógica de negócio em `main.go` além de: parse de flags, carregar config, chamar bootstrap, iniciar/parar o processo.

**Dois servidores HTTP por serviço.** Servidor de negócio (rotas `/api/v1/...`) e servidor de health check em porta separada, configurável independentemente. Prefira, quando prático, expor `/healthz` (liveness) e `/readyz` (readiness) como endpoints distintos no servidor de health.

## Estrutura recomendada de novo projeto

Ver [new-project-checklist.md](new-project-checklist.md) para o passo a passo completo. Resumo da árvore inicial:
```
cmd/api/main.go
cmd/worker/main.go        # se houver processamento assíncrono
configs/{api,worker}.example.yml
internal/{domain,port,usecase,adapter,app,config,dto}/
scripts/sh/{build,deploy,test,swagger,healthcheck}.sh
Makefile
Dockerfile
.devcontainer/
.golangci.yml
```

## Convenções de código

| Tema | Regra |
|---|---|
| Nome de arquivo | `snake_case.go`, um arquivo por operação em `usecase/` |
| DTOs | `binding:"required"` (tags do Gin/validator), nunca `validate:` standalone |
| Erros de domínio | sentinelas locais (`var ErrX = errors.New(...)`) + `fmt.Errorf("%w: ...", ErrX)`; sem pacote central de erros |
| ID de entidade | `uint` autoincrement via `BaseModel` embutido é o padrão; use `uuid.UUID` quando o ID precisar ser gerado/exposto externamente sem coordenação |
| Base de entidade | `BaseModel{ID, CreatedAt, UpdatedAt time.Time, DeletedAt gorm.DeletedAt}` embutido — ver `instructions/database.md` |
| Paginação | `github.com/raykavin/gobox/pagination` — `Page`/`Result[T]`/`Scope` — nunca reimplementar |
| Logging | `github.com/raykavin/gobox/logger` (zerolog por baixo) — nunca criar wrapper local novo |
| Resposta HTTP | `github.com/raykavin/gobox/httpserver/respond` — importar, não copiar |

Detalhe completo por tema em `instructions/go.md`.

## APIs

- Framework: **Gin**. Prefixo de rota fixo `/api/v1`. Recursos em kebab-case, plural (`bank-accounts`, `document-types`).
- Leitura por ID: `GET /recurso/:id`. Busca com filtro: `POST /recurso/search` com o filtro no body — é o padrão a seguir mesmo parecendo não-RESTful, porque é o que `gobox/pagination` espera (filtro complexo não cabe bem em querystring).
- Update: `PATCH` por padrão; use `PUT` apenas quando o recurso inteiro é substituído.
- Autenticação: `github.com/raykavin/gobox/oidcauth` + `gobox/httpserver/middlewares`. **Nunca crie um novo verificador OIDC do zero.**
- Autorização: declarativa por rota/grupo, nunca dentro do usecase. Permissão nomeada `"recurso:acao"` (ex.: `"invoice:validate"`) para usuários humanos; scope OAuth (`"servico:webhook:write"`) para clientes M2M.
- Validação: `go-playground/validator` via tags do Gin. Trate o erro de bind explicitamente — não deixe a mensagem crua do validator (em inglês) vazar para o cliente; parseie campo-a-campo.
- Paginação: sempre via `gobox/pagination`. Nunca implemente `LIMIT`/`OFFSET` manual em código novo se o projeto já usa GORM.
- Idempotência de ingestão (upload de arquivo, importação, webhook): hash SHA-256 do conteúdo + índice único composto com o escopo pai (`(parent_id, file_hash)`) + fluxo find-or-create no usecase. Use `clause.OnConflict{DoNothing:true}` do GORM quando não precisar devolver o registro existente.
- Documentação: `swaggo/swag`, anotações nos handlers, gerado via `make swagger`. **Mantenha as anotações `@Failure`/`@Success` sincronizadas com o `switch`/mapeamento de erro real do handler.**

## Dados

- ORM: GORM sobre Postgres (produção) / SQLite (desenvolvimento) é o padrão para serviços com domínio relacional. pgx puro sem ORM é aceitável apenas quando a entidade é simples (poucas colunas, sem relações) e há forte pressão de performance/controle — não é o padrão para novos serviços de domínio relacional geral.
- Migrations vs. `AutoMigrate`: `db.AutoMigrate` no boot é aceitável em fase inicial/baixo risco, mas tem risco real em produção (sem rollback, sem histórico de versão, `ALTER`/índice não roda `CONCURRENTLY`). Para dados críticos, prefira migrations versionadas (`golang-migrate` + binário `migrator` dedicado).
- Transações: use um `TransactionRunner`/Unit-of-Work propagado por `context.Context`. Não faça usecases multi-etapa sem transação.
- N+1: use `Preload` do GORM para hidratar associações em leituras; reserve `Joins` manual para filtros de busca.
- Soft delete: `gorm.DeletedAt` embutido no `BaseModel`. Se a entidade tiver índice único que precise ignorar linhas soft-deleted, use índice único parcial (`uniqueIndex:...,where:deleted_at IS NULL`).

Detalhe completo em `instructions/database.md`.

## Erros

- Sentinelas locais por pacote (`var ErrNotFound = errors.New(...)`), nunca um pacote central de exceções.
- Wrap com `fmt.Errorf("%w: %w", ...)` (double-wrap, Go 1.20+) ou `errors.Join` quando aplicável.
- No handler HTTP, mapeie o erro de domínio para status HTTP explicitamente (`errors.Is`/`errors.As`) — nunca deixe todo erro cair em 500 por omissão.

## Logging

- `github.com/raykavin/gobox/logger` (zerolog por baixo), configurável (console/JSON). Nunca escreva um wrapper zerolog novo.
- Nunca logue payload completo de webhook/integração que carregue dado pessoal em nível `debug` sem redigir campos sensíveis.

## Segurança

Ver `instructions/security.md` para a lista completa. Regras não-negociáveis:
- Autenticação sempre via `gobox/oidcauth` — nunca reimplemente verificação de JWT/introspecção.
- CORS sempre via `gobox/httpserver/middlewares` (versão atual, allowlist) — nunca copie/reimplemente CORS localmente.
- CSRF obrigatório em qualquer rota que aceite cookie de sessão.
- Segredos: apenas via `configs/*.example.yml` versionado + arquivo real no `.gitignore`. Nunca commit de arquivo com valor real, nem como "exemplo temporário".

## Testes

- Comando mínimo: `go test -v -race ./...`.
- `go build ./...` e `go vet ./...` antes do teste.
- `.golangci.yml` com o conjunto `govet, errcheck, staticcheck, ineffassign, unused, gofmt, goimports` como ponto de partida em projetos novos.
- Mocks: prefira mocks funcionais escritos à mão (structs com campos `func`) ou SQLite in-memory para testes de repositório que envolvem transação/GORM real, em vez de depender de geração automática de mocks.

Detalhe completo em `instructions/testing.md`.

## Documentação

- README com as seções: Arquitetura, Configuração, Execução local, Scripts/Makefile, Testes, Documentação da API.
- Documentação para agentes de IA (`.claude/`, `AGENTS.md`) deve ser atualizada no mesmo PR que muda a estrutura que ela descreve.

## Observabilidade

- Logging estruturado via `gobox/logger` é a base mínima obrigatória.
- OpenTelemetry+Prometheus via `gobox/telemetry` é recomendado quando o projeto precisar de tracing/métricas, em vez de reimplementar.

## Infraestrutura

- Dockerfile multi-stage: build em `golang:1.26-alpine3.22`, runtime em `alpine:3.22`, usuário non-root, `HEALTHCHECK` via `scripts/sh/healthcheck.sh`. **Confirme que o script referenciado existe** antes de considerar o build pronto.
- `scripts/sh/build.sh` deve apontar para o caminho real do Dockerfile. Teste `make build` antes de considerar um novo projeto pronto.
- `make {build,deploy,test,mocks,swagger}` delegando a `scripts/sh/*.sh` é o padrão de automação.
- **CI/CD**: novo projeto deve incluir um pipeline básico (lint + build + test) desde o início. Ver `new-project-checklist.md`.

## Critérios de conclusão de uma tarefa

Uma alteração só está concluída quando:
1. `go build ./...` sem erro.
2. `go vet ./...` sem erro.
3. `golangci-lint run` sem erro, se o projeto tiver `.golangci.yml` (ou foi adicionado como parte da tarefa).
4. `go test -v -race ./...` sem falha, incluindo testes novos/atualizados para o comportamento alterado.
5. Contratos públicos (rotas HTTP, schema de banco, formato de resposta) preservados, salvo autorização explícita.
6. Documentação (README/Swagger/comentário godoc) atualizada se o comportamento documentado mudou.

## Decisões que exigem aprovação humana

Ver seção "Operações que exigem aprovação" em [../../AGENTS.md](../../AGENTS.md).
