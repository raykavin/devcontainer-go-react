# Instruções específicas — APIs e contratos

Complementa `AGENTS.md`.

## Framework e rotas

- Gin é o framework HTTP padrão. `net/http` puro é aceitável apenas para casos simples e isolados (ex.: um webhook receiver com uma única rota).
- Prefixo de rota fixo: `/api/v1`. Não crie `/v2` sem antes registrar uma decisão de versionamento — não há estratégia de depreciação definida por padrão.
- Recursos em kebab-case, plural: `bank-accounts`, `document-types`.
- Leitura por ID: `GET /recurso/:id`.
- Busca/listagem com filtro: `POST /recurso/search` com o filtro no corpo da requisição — **não** use `GET` com querystring para filtros compostos; o filtro se integra diretamente com `gobox/pagination.NewFilterBuilder()`.
- Update: `PATCH` por padrão; use `PUT` apenas quando o recurso inteiro é substituído. Ao editar um endpoint existente, mantenha o verbo já usado.

## Envelope de resposta

`gobox/httpserver/respond` define o formato canônico:
```go
type Response struct {
    Success bool          `json:"success"`
    Message string        `json:"message,omitempty"`
    Data    any           `json:"data,omitempty"`
    Errors  []*APIError   `json:"errors,omitempty"`
}
```
com helpers `OK/Created/Accepted/BadRequest/NotFound/InternalServerError/...`.

Importe `gobox/httpserver/respond` diretamente — não copie o código para o repositório local. Trocar o envelope de resposta de um projeto existente é uma mudança incompatível de contrato e exige aprovação humana.

## Erro de domínio → status HTTP

Ao criar/editar um handler: mapeie explicitamente cada erro de domínio relevante para o status HTTP correto via `errors.Is`/`errors.As`. Não deixe todo erro cair em 500 por omissão. Se encontrar uma anotação Swagger que não bate com o comportamento real do código, corrija um dos dois lados (preferencialmente o código, se o contrato documentado for o esperado pelo consumidor) e sinalize a inconsistência.

## Autenticação e autorização

- Autenticação: sempre via `gobox/oidcauth` (Bearer/cookie OIDC contra o provider OIDC configurado). **Nunca** implemente verificação de JWT/introspecção do zero.
- Autorização humana: RBAC declarativo por rota, permissão nomeada `"recurso:acao"` (ex.: `"invoice:validate"`, `"company:read"`). A checagem sempre acontece no middleware/router, nunca dentro do usecase.
- Autorização M2M (serviço-a-serviço, webhooks): scope OAuth (`RequireScope`) + verificação de `azp` (authorized party) quando o cliente é conhecido.

## Webhooks

Dois mecanismos de autenticação são válidos, dependendo da capacidade do sistema de origem:
- **M2M OIDC completo**: Bearer client_credentials, scope + `azp` obrigatórios. Use quando o sistema de origem fala OIDC nativamente.
- **Token estático**: token fixo em query string, comparado com `subtle.ConstantTimeCompare` sobre SHA-256, **e redigido do log de acesso antes de escrever**. Use apenas quando o sistema de origem não oferece OIDC — sempre implemente a comparação em tempo constante e a redação em log.

## Validação de entrada

- `go-playground/validator` via tags `binding:"required"` do Gin (`ctx.ShouldBindJSON`).
- Trate o erro de bind explicitamente. Não deixe a mensagem crua do validator (inglês, formato `"Key: '...' Error:Field validation..."`) vazar para o campo `details`/`error` da resposta — parseie campo-a-campo.

## Paginação

Sempre via `github.com/raykavin/gobox/pagination`:
```go
filters := pagination.NewFilterBuilder().WhereIf(cond, "coluna", pagination.Eq, valor).Build()
sorts := pagination.NewSortBuilder().OrderBy("created_at", pagination.Desc).Build()
query := pagination.NewQuery(page, filters, sorts)
db.Scopes(pagination.Scope(query, &total)).Find(&resultados)
result := pagination.NewResult(resultados, int(total), query.Page)
```
Entrada de paginação sempre pelo **corpo** da requisição (dentro do filtro de `POST /recurso/search`), nunca por querystring `?page=`. Nota: `pagination.Scope` é específico para `*gorm.DB` — se o projeto usa pgx puro, implemente `LIMIT`/`OFFSET` manual a partir de `Query.Page.Offset()`/`PerPage`, replicando os mesmos nomes de campo por consistência.

## Idempotência de ingestão

Para upload de arquivo, importação de planilha, ou processamento assíncrono de retorno: hash SHA-256 do conteúdo + índice único composto `(escopo_pai_id, file_hash)` + fluxo find-or-create no usecase. Documente explicitamente a janela de corrida aceita (double-upload concorrente do mesmo arquivo).

## Documentação OpenAPI

`swaggo/swag`, gerado via `make swagger` a partir de anotações `@Summary/@Router/@Success/@Failure` nos handlers. Toda tarefa que muda o comportamento de um endpoint deve regenerar a documentação e verificar que as anotações batem com o código real.

## Alterações que exigem aprovação humana

Qualquer mudança que quebre compatibilidade retroativa de um endpoint existente: remover/renomear campo de resposta, trocar o envelope de resposta, trocar o verbo HTTP, mudar o formato de paginação, remover uma permissão exigida. Ver seção "Operações que exigem aprovação" em `AGENTS.md`.
