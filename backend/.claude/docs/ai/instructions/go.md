# Instruções específicas — Go

Complementa `AGENTS.md` e `backend-development-standard.md`.

## Versão e módulo

- Go 1.26.x é a versão alvo deste template.
- Módulo Go Modules nativo (`go.mod`/`go.sum`); sem vendoring.
- `replace` local para bibliotecas próprias ainda não publicadas é aceitável (ex.: `replace github.com/raykavin/gocnab => ./pkg/gocnab`) — mas prefira publicar como módulo versionado (tag semver) assim que estabilizar.

## gobox — biblioteca de plataforma, verificar antes de reimplementar

Antes de escrever qualquer um destes, confira o subpacote correspondente em `github.com/raykavin/gobox`:

| Preocupação | Subpacote gobox | Observação |
|---|---|---|
| Logging | `logger` | zerolog por baixo, interface tipo logrus |
| Config | `config` | `Loader[T]` genérico sobre Viper, hot-reload via fsnotify |
| Servidor HTTP | `httpserver` | embala Gin; `respond/` = envelope de resposta; `middlewares/` = auth/CORS/CSRF/oidcauth |
| Autenticação OIDC | `oidcauth` | verificação JWT local + introspecção remota, cache plugável |
| Paginação | `pagination` | `Page`/`Result[T]`/`Query`/`Filter` — `Scope()` é específico para `*gorm.DB` |
| Banco (GORM) | `database/gorm` | factory de conexão com retry e pool configurável — **não** define `BaseModel`, nem soft delete, nem transação; isso fica a cargo do projeto consumidor |
| Migrations | `database/migrate` | wrapper sobre `golang-migrate`; só expõe `Migrate()` (Up) e `Populate()` (seeds) |
| SQL genérico | `database/sql` | `Connector[T]` sobre `database/sql` puro, sem pooling extra, sem transação |
| Healthcheck | `healthcheck` | probe concorrente via interface `Pinger` |
| Telemetria | `telemetry` | bootstrap OpenTelemetry (traces OTLP + métricas Prometheus) |
| Retry | `retry` | função pura `Do(ctx, maxAttempts, backoff, shouldRetry, fn)` |
| Cliente HTTP | `httpclient` | wrapper fino sobre `net/http` + descompressão |
| OAuth2 client-credentials | `oauth2` | `TokenManager` com cache de token em memória |

**Regra de contribuição**: se uma preocupação transversal está sendo reimplementada localmente em vez de importar `gobox` (ex.: envelope de resposta HTTP, verificador OIDC, wrapper de logging), prefira migrar para a lib compartilhada como parte da tarefa — **não decida sozinho fazer uma migração ampla não solicitada**; registre a decisão se a migração for maior que o escopo pedido.

## Convenções de código

- Nome de arquivo: `snake_case.go`. Um arquivo por operação em `usecase/` (`create.go`, `update.go`, `get_by_id.go`, `find_or_create.go`, `filter.go`).
- Pacotes: nome curto, minúsculo, sem underscore, um conceito por pacote.
- Erros: sentinelas locais por pacote (`var ErrNotFound = errors.New("...")`), nunca um pacote central de erros/exceções. Wrap com `fmt.Errorf("%w: %w", ErrX, causa)` (double-wrap, Go 1.20+) ou `errors.Join`. No handler HTTP, mapeie explicitamente via `errors.Is`/`errors.As` — nunca deixe cair em 500 por omissão.
- Generics: usados deliberadamente para tipos de infraestrutura reutilizáveis (`pagination.Result[T]`, `database/sql.Connector[T]`, `config.Loader[T]`) — não force generics em código de domínio/negócio onde não há reuso real.
- `context.Context` como primeiro parâmetro em toda operação de I/O/rede.
- Validação de entrada: tags `binding:"required"` do Gin (que usa `go-playground/validator` por baixo). Nunca use a tag `validate:"..."` sozinha fora do fluxo de bind do Gin.
- Comentários: apenas onde explicam uma decisão não-óbvia (ex.: por que uma coluna não tem FK, por que uma race condition específica foi corrigida de um jeito). Não documente o óbvio.

## Estilo de teste

- Framework: stdlib `testing` é o padrão. `testify` (`assert`/`require`/`mock`) é aceitável quando já em uso no projeto — não introduza em um projeto que hoje só usa stdlib sem necessidade concreta.
- Nomenclatura: `TestXxx_Cenário`, com subtests via `t.Run(...)` para variações. Table-driven quando há múltiplos casos de entrada/saída.
- Ver `instructions/testing.md` para comandos, mocks e cobertura.

## Linting

- `.golangci.yml` com `disable-all: true`, habilitando `govet, errcheck, staticcheck, ineffassign, unused, gofmt, goimports`, timeout 3m. Adote exatamente esse conjunto em projetos novos ou ao adicionar lint a um projeto existente.
