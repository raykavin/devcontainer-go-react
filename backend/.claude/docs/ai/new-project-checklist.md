# Checklist de início de projeto backend

Use como ponto de partida para um novo serviço backend Go a partir deste template. Itens marcados **[decisão]** exigem confirmação do responsável antes de seguir — não escolha sozinho.

## 1. Decisões arquiteturais

- [ ] Confirmar domínio/finalidade do serviço.
- [ ] **[decisão]** Tipo de ID de entidade: `uint` autoincrement (padrão recomendado) ou `uuid.UUID` — decida com base em se IDs serão expostos/gerados externamente.
- [ ] **[decisão]** Estratégia de schema: `db.AutoMigrate` no boot (mais simples, sem histórico/rollback) vs. migrations SQL versionadas + binário `migrator` (mais seguro para dados críticos).
- [ ] Confirmar se haverá processamento assíncrono — se sim, planejar `cmd/worker` desde o início.
- [ ] DI manual via `internal/app/` (bootstrap/factory) — não introduzir framework de DI (Fx, Wire) salvo decisão explícita em contrário.

## 2. Estrutura inicial

- [ ] Criar árvore: `cmd/api/`, `cmd/worker/` (se aplicável), `internal/{domain,port,usecase,adapter/{inbound/http,outbound},app,config,dto}/`.
- [ ] `Makefile` com alvos `build, deploy, test, swagger` delegando a `scripts/sh/*.sh` — **verificar que os caminhos referenciados nos scripts realmente existem**.
- [ ] `.golangci.yml` (`govet, errcheck, staticcheck, ineffassign, unused, gofmt, goimports`).
- [ ] `.devcontainer/` subindo só os serviços auxiliares realmente usados (Postgres se GORM; não copiar serviços de observabilidade "por precaução", a menos que o novo projeto realmente vá instrumentar telemetria).

## 3. Configuração

- [ ] `configs/api.example.yml` (+ `worker.example.yml` se aplicável) versionado, com placeholders (`"changeme"`), YAML.
- [ ] Arquivo real (`configs/api.yml`) no `.gitignore` desde o primeiro commit — confirmar que a regra do `.gitignore` não tem exceção acidental.
- [ ] Toda variável de configuração validada na inicialização (falha rápido se obrigatória e ausente) e documentada no `.example.yml`.
- [ ] Config carregada via `github.com/raykavin/gobox/config` (`Loader[T]` sobre Viper) — não escrever um loader novo.

## 4. Gestão de segredos

- [ ] Nenhum valor real em nenhum arquivo versionado, nem em commit de "configuração temporária".
- [ ] Segredos de runtime (client secret OIDC, DSN de banco, token de parceiro) só via arquivo de config real (ignorado) ou variável de ambiente injetada no deploy.

## 5. Banco de dados

- [ ] Conexão via `github.com/raykavin/gobox/database/gorm` (factory com retry+pool), não `gorm.Open` direto.
- [ ] `BaseModel` com `ID/CreatedAt/UpdatedAt/DeletedAt(gorm.DeletedAt)` embutido em toda entidade persistida — tipo de ID conforme decisão do item 1.
- [ ] Se optar por migrations versionadas: usar `github.com/raykavin/gobox/database/migrate`, binário `cmd/migrator` dedicado, arquivos `.up.sql`/`.down.sql` reais (não placeholders vazios) desde a primeira migration.
- [ ] Unit-of-Work: implementar `port.TransactionRunner` antes do primeiro usecase multi-etapa que grava em mais de uma tabela.
- [ ] Paginação via `gobox/pagination`, não implementação própria.

## 6. Autenticação e autorização

- [ ] Autenticação via `github.com/raykavin/gobox/oidcauth` — não escrever verificador OIDC próprio.
- [ ] CORS/CSRF via `github.com/raykavin/gobox/httpserver/middlewares` — não reimplementar.
- [ ] Definir a nomenclatura de permissão (`"recurso:acao"`) desde o primeiro endpoint, documentada em um único arquivo (`internal/domain/valueobject/permissions.go` ou equivalente).

## 7. Logging

- [ ] `github.com/raykavin/gobox/logger` — não criar wrapper zerolog local.
- [ ] Definir desde o início quais campos de payload/request são sensíveis (PII, credenciais) e garantir redação antes de logar em nível debug.

## 8. Tratamento de erros

- [ ] Sentinelas de erro locais por pacote (`var ErrX = errors.New(...)`), sem pacote central de exceções.
- [ ] Função/mapeamento explícito de erro de domínio → status HTTP em cada handler, desde o primeiro endpoint (não deixar cair em 500 por omissão).

## 9. Health checks

- [ ] Servidor de health check em porta separada do servidor de negócio.
- [ ] Preferir endpoints distintos `/healthz` (liveness) e `/readyz` (readiness).
- [ ] `scripts/sh/healthcheck.sh` real e testado — confirmar que o `Dockerfile` só faz `COPY` de um script que de fato existe.

## 10. Testes

- [ ] `go test -v -race ./...` funcionando desde o primeiro commit.
- [ ] `.golangci.yml` configurado e `golangci-lint run` passando.
- [ ] Testes de autorização/sucesso/entrada-inválida/não-encontrado para o primeiro endpoint, como modelo para os seguintes.
- [ ] Prefira mocks funcionais escritos à mão em vez de depender de geração automática (`mockgen`) sem testar se ela de fato produz artefatos úteis.

## 11. Qualidade

- [ ] `go build ./... && go vet ./... && golangci-lint run && go test -v -race ./...` como sequência de validação documentada no README.
- [ ] Incluir pipeline de CI básico (lint+build+test) desde o início.

## 12. Documentação

- [ ] `README.md` com as seções: Arquitetura, Configuração, Execução local, Scripts/Makefile, Testes, Documentação da API.
- [ ] Swagger (`swaggo/swag`) desde o primeiro endpoint, gerado via `make swagger`.
- [ ] `AGENTS.md`/`.claude/` do próprio projeto (se o time usar agentes de IA nele) mantido sincronizado com a estrutura real desde o início.

## 13. Containers

- [ ] Dockerfile multi-stage `golang:1.26-alpine3.22` → `alpine:3.22`, usuário non-root, `HEALTHCHECK` via script real.
- [ ] Testar `make build` e `docker build` localmente antes de considerar o projeto pronto.

## 14. Pipeline e deploy

- [ ] CI básico (lint+build+test) — ver item 11.
- [ ] `make deploy` → `scripts/sh/deploy.sh` (`docker push`) como padrão de partida; introduzir Kubernetes/Helm/rollback automatizado apenas sob pedido explícito.

## 15. Critérios de pronto

Ver seção "Definição de pronto" em `../../AGENTS.md`. Resumo: build/vet/lint/test passam; contratos preservados; documentação atualizada; nenhuma operação da lista "exige aprovação" executada sem confirmação.
