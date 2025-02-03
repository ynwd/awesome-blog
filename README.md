# Blog Service

A microservice for managing blog posts, comments, likes and user interactions.

## How to Run

Run the service:
```
go run cmd/main.go
```

The service will start on port 8080.

## How to Test

Run all tests:

```bash
go clean -testcache && go test -v ./...
```

Run specific module tests:

```bash
go test -v ./internal/users/...
go test -v ./internal/posts/...
go test -v ./internal/comments/...
go test -v ./internal/likes/...
go test -v ./internal/summary/...
```

## API Routes

### Authentication
| Method | Endpoint | Module | Description |
|--------|----------|---------|-------------|
| POST | `/register` | Users | Register new user |
| POST | `/login` | Users | User authentication |

### Posts
| Method | Endpoint | Module | Description |
|--------|----------|---------|-------------|
| POST | `/posts` | Posts | Create new post |
| POST | `/posts/pubsub` | Posts | Publish post event |

### Comments
| Method | Endpoint | Module | Description |
|--------|----------|---------|-------------|
| POST | `/comments` | Comments | Create new comment |
| POST | `/comments/pubsub` | Comments | Publish comment event |

### Likes
| Method | Endpoint | Module | Description |
|--------|----------|---------|-------------|
| POST | `/likes` | Likes | Create new like |
| POST | `/likes/pubsub` | Likes | Publish like event |

### Summary
| Method | Endpoint | Module | Description |
|--------|----------|---------|-------------|
| POST | `/summary` | Summary | Get user's yearly activity summary |

## Project Structure

| Directory | Purpose |
|-----------|---------|
| `/cmd` | Main application entry points |
| `/config` | Application configuration |
| `/internal` | Private application code |
| `  /internal/app` | Application bootstrapping and DI |
| `  /internal/users` | User authentication & management |
| `  /internal/posts` | Blog posts management |
| `  /internal/comments` | Comments management |
| `  /internal/likes` | Likes management |
| `  /internal/summary` | Activity summary |
| `/pkg` | Shared packages |
| `  /pkg/database` | Database utilities |
| `  /pkg/middleware` | HTTP middleware |
| `  /pkg/module` | Common interfaces |
| `  /pkg/pubsub` | PubSub utilities |
| `  /pkg/res` | HTTP response helpers |
| `  /pkg/utils` | Common utilities |
| `/tests` | Integration & E2E tests |
| `  /tests/e2e` | End-to-end tests |
| `  /tests/helper` | Test helpers |