myapp/
├── cmd/
│   └── api/
│       └── main.go              # entrypoint, wiring semua module
│
├── internal/
│   ├── modules/
│   │   ├── user/
│   │   │   ├── domain/           # entity, value object, business rules
│   │   │   │   └── user.go
│   │   │   ├── repository/       # interface + implementasi (postgres, mysql, dll)
│   │   │   │   ├── repository.go       # interface
│   │   │   │   └── postgres_repository.go
│   │   │   ├── service/          # business logic / use case
│   │   │   │   └── service.go
│   │   │   ├── handler/          # HTTP/gRPC handler
│   │   │   │   └── http_handler.go
│   │   │   ├── dto/              # request/response struct
│   │   │   │   └── dto.go
│   │   │   ├── events/           # event yg di-publish/di-consume module ini
│   │   │   │   └── events.go
│   │   │   └── module.go         # public API module (exported functions/interfaces)
│   │   │
│   │   ├── order/
│   │   │   └── ... (struktur sama)
│   │   │
│   │   └── payment/
│   │       └── ... (struktur sama)
│   │
│   ├── shared/                   # dipakai lintas module, hati2 jangan jadi tong sampah
│   │   ├── database/
│   │   ├── middleware/
│   │   ├── logger/
│   │   ├── config/
│   │   └── eventbus/             # in-process event bus (kalau butuh decouple)
│   │
│   └── platform/                 # infra wiring: DI container, router setup
│       └── server.go
│
├── pkg/                          # kalau ada yg mau di-reuse project lain (opsional)
│
├── migrations/
├── go.mod
└── docker-compose.yml


myapp/
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── modules/
│   │   ├── order/                          # 1 module = 1 Bounded Context
│   │   │   │
│   │   │   ├── domain/                     # === LAYER PALING DALAM ===
│   │   │   │   ├── order.go                # Aggregate Root
│   │   │   │   ├── order_item.go           # Entity (bagian dari aggregate)
│   │   │   │   ├── money.go                # Value Object
│   │   │   │   ├── order_status.go         # Value Object (enum-like)
│   │   │   │   ├── errors.go               # Domain errors
│   │   │   │   ├── events.go               # Domain Events
│   │   │   │   ├── repository.go           # Port: interface repository
│   │   │   │   └── service.go              # Domain Service (logic lintas aggregate)
│   │   │   │
│   │   │   ├── application/                # === USE CASE LAYER ===
│   │   │   │   ├── command/
│   │   │   │   │   ├── place_order.go       # 1 use case = 1 file/struct
│   │   │   │   │   └── cancel_order.go
│   │   │   │   ├── query/
│   │   │   │   │   └── get_order_detail.go
│   │   │   │   ├── dto/
│   │   │   │   │   └── order_dto.go         # request/response, bukan domain object
│   │   │   │   └── ports.go                # Port: interface ke luar (UserFetcher, PaymentGateway, dll)
│   │   │   │
│   │   │   ├── infrastructure/             # === ADAPTER: implementasi teknis ===
│   │   │   │   ├── persistence/
│   │   │   │   │   ├── postgres_repository.go  # implement domain.Repository
│   │   │   │   │   └── model.go                # struct utk mapping DB (beda dari domain entity)
│   │   │   │   ├── messaging/
│   │   │   │   │   └── event_publisher.go      # implement application.EventPublisher
│   │   │   │   └── acl/                         # Anti-Corruption Layer
│   │   │   │       └── user_client.go           # translate data dari module `user` ke bentuk domain `order`
│   │   │   │
│   │   │   ├── interfaces/                 # === ADAPTER: entrypoint dari luar ===
│   │   │   │   ├── http/
│   │   │   │   │   ├── handler.go
│   │   │   │   │   └── router.go
│   │   │   │   └── grpc/
│   │   │   │       └── handler.go
│   │   │   │
│   │   │   └── module.go                   # Wiring & public API module
│   │   │
│   │   ├── user/
│   │   │   └── ... (struktur sama)
│   │   │
│   │   └── payment/
│   │       └── ... (struktur sama)
│   │
│   ├── shared/
│   │   ├── kernel/                         # Shared Kernel (DDD term) - VO/konsep yg bener2 dipakai semua BC
│   │   │   └── id.go                       # misal: type ID string, dsb
│   │   ├── database/
│   │   ├── eventbus/
│   │   ├── logger/
│   │   └── config/
│   │
│   └── platform/
│       └── server.go
│
├── migrations/
├── go.mod
└── docker-compose.yml