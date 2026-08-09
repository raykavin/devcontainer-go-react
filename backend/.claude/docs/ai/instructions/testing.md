# Instruções específicas — Testes e qualidade

Complementa `AGENTS.md`.

## Comando de validação mínimo

Execute nesta ordem antes de considerar qualquer tarefa concluída:

```sh
cd backend
go build ./...
go vet ./...
golangci-lint run          # apenas se existir .golangci.yml no repositório
go test -v -race ./...
```

- `go test -v -race ./...` é o piso mínimo inegociável, sempre com `-race`.
- Rode `go build`/`go vet` como etapas isoladas mesmo que pareçam redundantes com o teste — são baratas e pegam problemas que o teste sozinho não pega.
- Se o repositório não tiver `.golangci.yml`, não invente um conjunto de linters diferente do recomendado em `instructions/go.md` — ou pule esse passo, ou (se a tarefa incluir "melhorar qualidade") adicione exatamente esse mesmo conjunto.
- Se algum passo não puder rodar no seu ambiente (ferramenta não instalada, sem acesso de rede para baixar dependência), declare isso explicitamente no relatório final da tarefa — nunca omita ou finja que passou.

## Framework de asserção

- stdlib `testing` é o padrão. `testify` (`assert`/`require`/`mock`) é aceitável quando já usado no projeto. Não introduza `testify` em um projeto que hoje só usa stdlib, a menos que a tarefa peça isso explicitamente.

## Nomenclatura e estilo

- `TestXxx_Cenário` com subtests via `t.Run("descrição do caso", func(t *testing.T) {...})`.
- Table-driven para múltiplos casos de entrada/saída (`mutate func(*Input)` por caso é um padrão útil).
- Nomeie o teste pelo comportamento, não pela implementação — o nome do teste deve documentar a condição que ele existe para prevenir.

## Mocks

Prefira um destes três padrões, nessa ordem de preferência para a maioria dos casos:
1. **Structs funcionais escritos à mão** — `type mockSubRepo struct { SearchFn func(...) ...; GetByIDFn func(...) ... }`, implementando a interface de `port` com campos de função substituíveis por teste. É a opção mais simples de manter e a recomendação padrão.
2. **`testify/mock`** — quando `testify` já é usado no projeto.
3. **SQLite in-memory real** — para testar transação/GORM de verdade em vez de mockar cada colaborador. Prefira esta opção quando o teste precisa validar uma transação multi-tabela de verdade.

Não dependa de `make mocks`/`mockgen` sem antes confirmar que ele de fato gera artefatos úteis no projeto — é comum esse alvo existir no Makefile por cópia de outro projeto sem nunca ter sido usado de verdade.

## Cobertura

Não adicione um gate de cobertura mínima arbitrário a menos que a tarefa peça isso explicitamente.

## Testes obrigatórios por tipo de alteração

- **Novo endpoint HTTP**: teste de autorização (sem permissão → 403/401), sucesso, entrada inválida (400), recurso inexistente (404) quando aplicável.
- **Nova regra de domínio/usecase**: teste unitário cobrindo o caminho feliz e pelo menos uma condição de erro de negócio relevante.
- **Nova migration/alteração de schema**: se o projeto usa SQL versionado, teste que a migration aplica sem erro contra um banco limpo (via `cmd/migrator` ou script equivalente); se usa `AutoMigrate`, verifique que `AutoMigrate` não falha ao rodar contra o schema atual.
- **Código concorrente** (filas, workers, workflows): sempre escreva teste, mesmo que o restante do arquivo/pacote tenha baixa cobertura — código concorrente sem teste é onde bugs sutis mais sobrevivem sem serem notados.

## CI/CD

Novos projetos devem incluir um pipeline básico (lint + build + test) desde o início — ver `new-project-checklist.md`. Para repositórios existentes sem CI, adicionar é uma melhoria recomendada, não uma ação a executar sem que a tarefa peça isso explicitamente (mudança em pipeline de build/deploy exige aprovação — ver `AGENTS.md`).
