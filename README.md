# HopRun

A Go-based web service that converts natural language queries into SQL queries using AI, enabling users to query their databases using plain English.

## Overview

HopRun bridges the gap between natural language and SQL by leveraging OpenAI's language models and dynamic database schema introspection. Users can connect their databases, create projects, and execute queries using conversational language instead of writing SQL.

## Key Features

- **Natural Language to SQL**: Convert plain English questions into SQL queries
- **Dynamic Schema Introspection**: Automatically analyzes database structure for accurate query generation
- **Multi-Database Support**: Connect and query multiple user databases
- **User Authentication**: JWT-based authentication and project management
- **Encrypted Connections**: Database credentials are encrypted at rest

## How It Works

### The Introspection Advantage

HopRun uses **runtime database introspection** to provide context to the AI model:

1. **Schema Discovery**: When a query is received, HopRun introspects the target database using PostgreSQL's `information_schema` to extract:
   - All table names in the public schema
   - Column names and data types for each table

2. **Context Generation**: The schema is formatted into a readable structure:
   ```
   Table users:
     id (integer)
     email (character varying)
     password_hash (character varying)

   Table projects:
     id (integer)
     name (text)
     user_id (integer)
   ```

3. **AI-Powered Conversion**: This schema context is sent to OpenAI along with the user's natural language query, enabling the AI to generate accurate, database-specific SQL queries.

**Implementation**: See [`GetDatabaseSchema()`](internal/database/database.go#L34-L68) in [internal/database/database.go](internal/database/database.go)

### Comparison: Runtime Introspection vs. Static Context Files

HopRun's approach is analogous to how AI coding assistants like Claude Code use `CLAUDE.md` files, but with key differences:

| Aspect | HopRun (Schema Introspection) | Claude Code (CLAUDE.md) |
|--------|-------------------------------|-------------------------|
| **Context Source** | Runtime database introspection | Static markdown file |
| **Accuracy** | Always reflects current database state | Requires manual updates to stay accurate |
| **Scope** | Database schema only | Architecture, conventions, TODOs, security notes |
| **Performance** | Schema fetched on each query (can be cached) | Read once, cached in context |
| **Use Case** | Dynamic databases that may change | Codebases with stable patterns and conventions |

**Why Both Approaches Matter:**

- **HopRun's introspection** ensures the AI always works with the current database structure, critical when schemas evolve
- **Claude Code's CLAUDE.md** provides high-level context (architecture patterns, business logic, development workflows) that can't be introspected from code alone

**The Ideal Combination**: For maximum AI accuracy, combine both approaches:
- Use **introspection** for structural discovery (schema, APIs, type definitions)
- Use **context files** for domain knowledge (business rules, conventions, security requirements)

HopRun currently uses runtime introspection, but could benefit from adding a `SCHEMA.md` or per-project context file describing business logic, query patterns, and domain-specific rules to further improve SQL generation accuracy.

## Architecture

```
cmd/
  server/main.go          # Application entry point
internal/
  api/http.go            # HTTP handlers and routing
  auth/service.go        # Authentication and project management
  database/database.go   # Database operations and schema introspection
  database_connection/   # User database connection management with encryption
  nlp/nlp.go            # OpenAI integration for NL→SQL conversion
  query/query.go        # SQL query execution
  middleware/auth.go    # JWT authentication middleware
pkg/
  models/               # Data models (User, Project, DatabaseConnection)
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- OpenAI API key

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/cr34t1ve/hoprun.git
   cd hoprun
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up the database:
   ```bash
   createdb hoprun
   ```

4. Configure the OpenAI API key in [cmd/server/main.go:34](cmd/server/main.go#L34)

5. Update database connection string in [cmd/server/main.go:23](cmd/server/main.go#L23) if needed

### Running the Server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/register` | POST | Create a new user account |
| `/login` | POST | Authenticate and receive JWT token |
| `/project` | POST | Create a new project |
| `/getproject` | POST | List user's projects |
| `/addConnection` | POST | Add a database connection |
| `/getConnections` | POST | List database connections |
| `/query` | POST | Execute a natural language query |

### Example Query Request

```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "query": "show me all users who registered in the last 7 days",
    "connection_id": 1
  }'
```

## Development

**Format code:**
```bash
go fmt ./...
```

**Run linter:**
```bash
go vet ./...
```

**Check for issues:**
```bash
golangci-lint run
```

## Technology Stack

- **Go 1.21+**: Core language
- **GORM**: PostgreSQL ORM
- **Gorilla Mux**: HTTP routing
- **go-openai**: OpenAI API client
- **golang-jwt**: JWT authentication

## Roadmap & Known Limitations

- [ ] **Schema Caching**: Currently fetches schema on every query; implement caching with change detection
- [ ] **Connection Pooling**: Improve database connection lifecycle management
- [ ] **Password Decryption**: Complete decryption implementation for stored database credentials
- [ ] **Environment Variables**: Move API keys and secrets to environment configuration
- [ ] **Query Parameterization**: Add SQL injection protection for user-generated queries
- [ ] **Test Coverage**: Add unit and integration tests
- [ ] **Context Files**: Add per-project context files to improve query accuracy with business logic

## Security Considerations

- Database connection passwords are encrypted using AES-256
- JWT tokens are used for authentication
- OpenAI API key should be moved to environment variables (currently hardcoded)
- Consider parameterized queries or query validation for additional SQL injection protection

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by the need to make database querying accessible to non-technical users
- Built with insights from AI-assisted development patterns like Claude Code's context files
