# Instruções específicas — Banco de dados e persistência

Complementa `AGENTS.md`.

## ORM e driver

- GORM sobre Postgres é o padrão para domínio relacional geral. Use `github.com/raykavin/gobox/database/gorm` como factory de conexão (retry de conexão embutido, pool configurável via `GormConfig`) — não abra a conexão manualmente com `gorm.Open` direto.
- SQLite é aceitável para desenvolvimento local ou serviços de baixa escala sem infraestrutura de banco compartilhada.
- pgx puro (`jackc/pgx/v5`) sem ORM é aceitável **apenas** para entidades simples (poucas colunas, sem relações profundas) com alta pressão de performance/controle explícito sobre SQL. Para entidades com muitos campos JSONB aninhados, essa abordagem fica visivelmente mais verbosa que GORM — não escolha pgx puro só por preferência estilística quando a entidade tem essa forma; avalie o trade-off real antes.

## Modelo base e ID de entidade

Padrão recomendado: `uint` autoincrement, embutido via `BaseModel`. Use `uuid.UUID` (`gorm:"default:gen_random_uuid()"`) quando o ID precisar ser exposto ou gerado externamente sem coordenação (ex.: sincronização multi-serviço, geração no cliente).

Estrutura recomendada (independentemente do tipo de ID escolhido):
```go
type BaseModel struct {
    ID        <uint ou uuid.UUID>
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```
Embuta em toda entidade persistida. Para entidades de junção/histórico append-only sem necessidade de update/soft-delete, é aceitável omitir `DeletedAt` — mas documente por quê no próprio código.

## Soft delete

- `gorm.DeletedAt` é o mecanismo padrão, embutido via `BaseModel`.
- Se uma coluna precisar de unicidade que ignore linhas soft-deleted, use índice único parcial: `gorm:"uniqueIndex:uq_x,where:deleted_at IS NULL"` — sem essa proteção não é possível recriar um registro com o mesmo valor único após soft-delete.

## Migrations vs. AutoMigrate

- `db.AutoMigrate(...)` chamado uma vez no boot do processo é conveniente para fase inicial/protótipo/domínio de baixo risco, mas: sem histórico de versão, sem `Down`, sem revisão humana antes de alterar produção, e a criação de índice não usa `CONCURRENTLY` (pode bloquear tabela grande).
- Para dados críticos/financeiros: prefira migrations SQL versionadas (`golang-migrate`, formato `NNNNNN_descricao.up.sql`/`.down.sql`) aplicadas por um binário `cmd/migrator` dedicado, **não** rodado automaticamente no boot da API/worker.
- **Regra não-negociável independentemente da decisão acima**: não altere uma migration já aplicada em produção — crie uma nova. Qualquer `DROP COLUMN`/`DROP TABLE`/`ALTER TYPE`/`RENAME` em uma tabela com dados reais exige aprovação humana explícita (ver `AGENTS.md`).

## Transações e Unit of Work

- Implemente uma interface `port.TransactionRunner` com `RunInTransaction(ctx, fn func(ctx) error)`, sobre `gorm.DB.Transaction`, propagada via `context.WithValue` — cada repositório usa um helper `dbFor(ctx, r.db)` que usa a transação ambiente se existir, senão a conexão base.
- Usecases multi-etapa (ex.: dois `Upsert` sequenciais em tabelas diferentes) precisam ser atômicos — não escreva um usecase multi-etapa sem transação.
- Para testes de repositório que envolvem transação real, use SQLite in-memory (`sqlite.Open(":memory:")` ou `file:name?mode=memory&cache=shared`) em vez de mockar cada colaborador.

## N+1 e queries

- Use `Preload("Relação")` do GORM para hidratar associações em métodos de leitura. Encadeie (`Preload("A.B")`) para relações de segundo nível.
- Reserve `.Joins(...)` manual para filtros de busca (ex.: `Joins("LEFT JOIN people ON ...")` dentro de um método `Filter`/`Search`), nunca para carregar associações que deveriam usar `Preload`.

## Paginação na camada de repositório

Via `github.com/raykavin/gobox/pagination` (`Scope(query, &total)` como `gorm.Scopes`). Ver `instructions/api.md` para o padrão completo de uso ponta a ponta.

## Idempotência de escrita

- Padrão preferido para "buscar ou criar" (quando o chamador precisa do registro, existente ou novo): hash de conteúdo + índice único composto + find-or-create no usecase.
- Padrão preferido para "gravar e ignorar duplicata" (quando o chamador não precisa do registro de volta): `clause.OnConflict{DoNothing: true}` ou `{Columns: [...], DoNothing: true}` do GORM.
- Padrão preferido para upsert completo (criar ou sobrescrever): `clause.OnConflict{Columns: [...], UpdateAll: true}`.
- Não use `FirstOrCreate` do GORM sem avaliar condição de corrida — prefira o padrão explícito find-then-create-or-conflict acima.

## Índices, FKs e nomenclatura

- Convenção de nome de índice: `idx_<tabela>_<coluna(s)>`. Únicos compostos: prefira `uq_<tabela>_<colunas>` — evite o prefixo `idx_` para índices únicos.
- FKs via GORM: `gorm:"foreignKey:X;references:ID"`. Documente explicitamente quando uma referência é proposital sem FK ("weak reference" documentado no comentário da migration e no struct Go) — não adicione uma FK "corretiva" sem entender por que ela foi omitida.
- Ao escrever migrations SQL (quando o projeto usa esse mecanismo): agrupe `CREATE INDEX`/`ALTER TABLE ... ADD CONSTRAINT` em blocos idempotentes (`DO $$ IF NOT EXISTS ... END$$;`). Use `CREATE INDEX CONCURRENTLY` ao adicionar índice em tabela que já tem dados em produção.

## Riscos que exigem aprovação humana

`DROP TABLE`, `DROP COLUMN`, `ALTER TYPE`, `RENAME TABLE`/`RENAME COLUMN` em qualquer migration que possa rodar contra dados reais.
