package repositories

type Dialect string

const (
	DialectSQLite   Dialect = "sqlite"
	DialectPostgres Dialect = "postgres"
)

type SqlLoader interface {
	Load(queryName string) (string, error)
}

type SqlLoaderInstance struct {
	Dialect Dialect // Exported to allow checking dialect in repositories
	queries map[string]string
}

func NewSqlLoader(dialect Dialect) *SqlLoaderInstance {
	return &SqlLoaderInstance{
		Dialect: dialect,
		queries: getQueries(), // build the queries map with dynamic keys
	}
}

func (s *SqlLoaderInstance) Load(queryName string) (string, error) {
	key := buildKey(queryName, s.Dialect)
	if query, exists := s.queries[key]; exists {
		return query, nil
	}
	return "", ErrQueryNotFound
}
