# Instruções específicas — Segurança

Complementa `AGENTS.md`. Nenhum valor de segredo real deve ser reproduzido em nenhum documento ou artefato deste repositório — apenas mecanismos.

## Autenticação

- **Sempre** via `github.com/raykavin/gobox/oidcauth` contra o provider OIDC configurado (Keycloak ou equivalente). Nunca implemente verificação de JWT/introspecção do zero — isso tende a criar forks divergentes que não recebem correções de segurança feitas no `gobox`.
- JWT próprio + bcrypt + RBAC local (ex.: Casbin) é uma exceção aceitável apenas quando o serviço não depende de SSO corporativo/federado (produto standalone, RBAC single-org). Não migre para OIDC sem confirmar que o contexto de implantação mudou.

## Autorização

- RBAC declarativo por rota, permissão nomeada `"recurso:acao"`, ou scope OAuth para M2M. A checagem sempre no middleware/router — nunca dentro do usecase.
- Evite vazar a decisão de autorização para a camada de dados via `ctx.Value` além do necessário para auditoria — isso mistura autorização declarativa com decisão imperativa a jusante e dificulta auditar "quem pode fazer o quê" só olhando o router.

## CORS

- Sempre via `gobox/httpserver/middlewares` (versão atual, restritiva por allowlist de origem). Nunca copie/reimplemente CORS localmente — um fallback permissivo (`Access-Control-Allow-Origin: *` combinado com `Access-Control-Allow-Credentials: true`) é um erro comum em implementações locais e não deve existir em código novo.

## CSRF

- Obrigatório em qualquer rota que aceite cookie de sessão: cookie HttpOnly + middleware CSRF via `gobox/httpserver/middlewares`.
- Se uma rota ler cookie de sessão OIDC como fallback de autenticação mas não tiver middleware CSRF, isso é um gap de segurança — sinalize e trate como mudança que exige aprovação explícita (ver `AGENTS.md`).

## Gestão de segredos

- Padrão: `configs/*.example.yml` versionado (com placeholders tipo `"changeme"`), arquivo real (`configs/*.yml` sem sufixo `.example`) sempre no `.gitignore`.
- **Nunca** commit de arquivo de configuração com valor real, nem "temporariamente", nem em branch de feature.
- Ao encontrar qualquer arquivo que pareça conter uma credencial/token real durante uma tarefa (mesmo não versionado, ex. em `temp/`), **não reproduza o valor em nenhum lugar** (código, log, resposta, commit) — apenas informe a localização ao responsável.

## Webhooks

- Token estático: compare sempre com `subtle.ConstantTimeCompare` sobre hash (nunca `==` de string), e **redija o token de qualquer log de acesso/URL antes de escrever**.
- M2M OIDC: exija scope + `azp` (authorized party) conhecido, não apenas um Bearer token válido genérico.

## Rate limiting

Não é implementado por padrão neste template — é uma recomendação para projetos que expõem endpoints de login e webhook publicamente, não uma garantia existente.

## Dados sensíveis em log

- Não existe mascaramento automático de PII em `gobox/logger` — cada projeto precisa resolver isso pontualmente.
- Se a tarefa envolver logging de payload que contenha CPF/e-mail/telefone/dado bancário, redija explicitamente esses campos antes de logar, especialmente em nível `debug`/`info` — não assuma que existe proteção automática.

## Separação de ambientes

Segregue por arquivo de config diferente por processo (`api.yml` vs `api.dev.yml` vs `api.prod.yml`), não por variável `APP_ENV` lida em runtime. Trate `.gitignore`/scripts de deploy com cuidado — um arquivo de config real acidentalmente versionado é o vetor de vazamento mais comum.

## Checklist de segurança para qualquer novo endpoint

1. Middleware de autenticação (`gobox/oidcauth`) aplicado ao grupo de rota correto.
2. Middleware de autorização com permissão/scope específico da operação, não uma permissão genérica reaproveitada.
3. CORS via `gobox/httpserver/middlewares`, não reimplementado.
4. CSRF se a rota depende de cookie de sessão.
5. Validação de entrada via binding+validator, sem vazar mensagem crua do validator ao cliente.
6. Nenhum dado sensível logado sem redação, mesmo em nível debug.
7. Se o endpoint é um webhook: comparação em tempo constante e redação de token em log.
