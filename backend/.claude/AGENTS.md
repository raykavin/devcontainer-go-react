# Instruções para agentes

Este arquivo é o ponto de entrada operacional para qualquer agente de IA trabalhando neste backend Go. É deliberadamente curto — o detalhe fica em `docs/ai/`.

## Objetivo do projeto

Este repositório é um serviço backend Go (API REST, tipicamente), construído em arquitetura hexagonal, consumindo a biblioteca compartilhada `github.com/raykavin/gobox` para preocupações transversais (logging, config, resposta HTTP, paginação, OIDC, banco, healthcheck). Confirme o propósito de negócio específico no `README.md` do repositório antes de agir.

## Fontes de verdade (nesta ordem)

1. Requisitos explícitos da tarefa atual.
2. Instruções locais do repositório (`.claude/`, `CLAUDE.md`, `README.md` do próprio projeto) — sempre têm prioridade sobre este documento.
3. Contratos públicos já implementados (rotas HTTP, schema de banco, formato de resposta) — não quebre sem autorização.
4. Este `AGENTS.md`.
5. Documentos específicos em `docs/ai/instructions/` (go, api, database, testing, security).
6. `docs/ai/backend-development-standard.md`.
7. Convenções inferidas do código existente no próprio repositório.
8. Recomendações gerais de engenharia.

Se duas fontes do mesmo nível de prioridade conflitarem, **pare e peça uma decisão** — não escolha silenciosamente.

## Como iniciar uma tarefa

1. Leia o `README.md` do repositório e qualquer `.claude/`/`CLAUDE.md` local — eles têm prioridade sobre este arquivo.
2. Procure uma implementação equivalente já existente no repositório antes de criar uma abstração nova (ex.: antes de escrever paginação, procure `gobox/pagination`; antes de escrever um envelope de resposta HTTP, procure `gobox/httpserver/respond`).
3. Antes de adicionar uma dependência externa nova, verifique se já existe solução equivalente em `github.com/raykavin/gobox` ou em outra dependência já presente no `go.mod`.

## Arquitetura

Hexagonal / ports-and-adapters. Ver `docs/ai/backend-development-standard.md` seção "Arquitetura padrão" para a árvore de diretórios completa.

## Estrutura de diretórios

```
cmd/<binário>/main.go     — só bootstrap: flags, config, chama internal/app, inicia/para o processo
internal/domain/          — entidades e regras, sem I/O
internal/port/            — interfaces (contratos)
internal/usecase/         — 1 arquivo por operação de negócio
internal/adapter/inbound/http/{handler,middleware,response,router}
internal/adapter/outbound/<x>/   — implementações concretas de port
internal/app/ (ou internal/platform/) — DI manual, chamado uma vez por cmd/*/main.go
internal/config/          — carregamento/validação de config
scripts/sh/*.sh           — build/deploy/test/swagger/healthcheck, chamados pelo Makefile
```

## Regras de dependência

- `domain` → nada de `adapter`/`infra`/`config`/`app`.
- `usecase` → só `domain` e `port` (interfaces), nunca `adapter/outbound` diretamente.
- `adapter/outbound/*` → implementa `port`; único lugar com GORM/drivers/clientes de terceiros.
- `adapter/inbound/http/handler` → valida entrada, chama 1 usecase, formata saída; **nunca** implementa regra de negócio.
- Composição/DI → centralizada em `internal/app/` (ou `internal/platform/`), manual, sem framework de DI (não introduza Uber Fx, Wire etc. salvo decisão explícita).

## Convenções de implementação

- 1 arquivo por operação em `usecase/` (`create.go`, `update.go`, `find_or_create.go`), `snake_case.go`.
- Erros: sentinelas locais (`var ErrX = errors.New(...)`) + `fmt.Errorf("%w: ...", ErrX)`. Sem pacote central de erros.
- DTOs validados via `binding:"required"` (Gin/go-playground/validator), nunca `validate:` standalone.
- Logging via `github.com/raykavin/gobox/logger` — nunca escreva um wrapper zerolog novo.
- Paginação via `github.com/raykavin/gobox/pagination` — nunca reimplemente `LIMIT`/`OFFSET` manual em projeto que já usa GORM.

## APIs e contratos

- Framework: Gin. Prefixo fixo `/api/v1`. Recursos em kebab-case plural.
- `GET /recurso/:id` para leitura por ID; `POST /recurso/search` (filtro no body) para busca/listagem.
- Autenticação: **sempre** `github.com/raykavin/gobox/oidcauth` + `gobox/httpserver/middlewares`. Nunca implemente verificação OIDC/JWT do zero.
- Autorização: declarativa por rota (middleware), nunca dentro do usecase. Permissão nomeada `"recurso:acao"`.
- Todo novo endpoint deve ter testes de: autorização (acesso negado sem permissão), sucesso, entrada inválida, e recurso inexistente (quando aplicável).
- Mapeie erro de domínio → status HTTP explicitamente no handler (`errors.Is`/`errors.As`); nunca deixe todo erro cair em 500 por omissão.
- Mantenha as anotações Swagger (`@Success`/`@Failure`) sincronizadas com o comportamento real do handler.

## Persistência e migrations

- ORM: GORM sobre Postgres/SQLite é o padrão para domínio relacional geral.
- **Não altere uma migration já aplicada; crie uma nova migration** — salvo se o projeto usar `db.AutoMigrate` sem SQL versionado (aceitável em fase inicial; avalie o risco de rodar em produção — ver `docs/ai/instructions/database.md`).
- Toda nova variável de ambiente/config deve ser validada na inicialização e documentada no arquivo `*.example.yml`.
- Toda alteração de schema que remove ou renomeia uma coluna, ou que roda em uma tabela grande, **exige aprovação humana explícita** — ver seção abaixo.
- Transações multi-etapa: use um `TransactionRunner` propagado por `context.Context`, não escreva usecases multi-etapa sem transação.

## Segurança

Ver `docs/ai/instructions/security.md` para a lista completa. Não-negociável: autenticação sempre via `gobox/oidcauth`; CORS sempre via `gobox/httpserver/middlewares`; CSRF obrigatório com cookie de sessão; segredos apenas via `*.example.yml` versionado + arquivo real no `.gitignore` — nunca commitar valor real.

## Testes e validação

Sequência mínima antes de considerar qualquer tarefa concluída:
```
go build ./...
go vet ./...
golangci-lint run     # se houver .golangci.yml no repositório
go test -v -race ./...
```
Se algum passo não puder ser executado (ex.: `golangci-lint` não instalado no ambiente), explique isso claramente no relatório final — não omita a falha nem finja que passou. Não remova ou reduza um teste apenas para fazer a validação passar. Prefira mocks funcionais escritos à mão ou SQLite in-memory para testes que envolvem GORM/transação, em vez de depender de geração automática de mocks.

## Documentação

Atualize `README.md` (seções: Arquitetura, Configuração, Execução local, Scripts/Makefile, Testes, Documentação da API) e anotações Swagger no mesmo PR que muda o comportamento que elas descrevem. Se o repositório tiver `.claude/`/`CLAUDE.md` próprio, atualize-o também quando a mudança afetar o que ele descreve.

## Operações que exigem aprovação humana explícita

Pare e peça confirmação antes de:
- Qualquer alteração incompatível de API (mudar formato de resposta, remover campo, trocar verbo HTTP de endpoint existente).
- Mudar mecanismo de autenticação/autorização.
- Migration que remove/renomeia coluna, ou `DROP`/`ALTER` em tabela com dados reais.
- Excluir arquivos ou módulos relevantes.
- Trocar framework, banco ou tecnologia estrutural (ex.: sair de GORM, trocar Gin por outro framework).
- Adicionar dependência de grande impacto (novo ORM, novo framework de DI, novo message broker).
- Alterar infraestrutura compartilhada (`gobox`) de um jeito que afete outros consumidores sem coordenar a versão.
- Mudar pipeline de build/deploy em produção.
- Reduzir cobertura de teste ou remover teste existente.
- Qualquer alteração ampla fora do escopo pedido.

## Ações proibidas

- Não invente requisito ou decisão arquitetural não solicitada.
- Não copie/reimplemente localmente algo que já existe em `gobox`.
- Não exponha segredos em código, log, documentação ou resposta — nem mesmo "temporariamente" ou em arquivo de exemplo.
- Não rode `--no-verify`, pule hooks, ou desative lint/testes para fazer uma tarefa "passar".

## Definição de pronto

Uma tarefa está concluída quando: (1) build/vet/lint/test passam (ver seção Testes); (2) contratos públicos preservados ou quebra explicitamente autorizada; (3) documentação relevante atualizada; (4) nenhuma operação da lista "exige aprovação" foi executada sem confirmação; (5) o agente relatou claramente o que mudou, por quê, e como validou.

## Documentos complementares

- [docs/ai/backend-development-standard.md](docs/ai/backend-development-standard.md) — padrão de desenvolvimento detalhado.
- [docs/ai/new-project-checklist.md](docs/ai/new-project-checklist.md) — checklist de início de projeto.
- [docs/ai/feature-template.md](docs/ai/feature-template.md) — fluxo recomendado para implementar uma feature.
- [docs/ai/instructions/](docs/ai/instructions/) — go.md, api.md, database.md, testing.md, security.md (detalhe por tema).
