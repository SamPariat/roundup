# Roundup — Complete Directory Structure

```
roundup/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                        # lint + test on every PR
│   │   ├── deploy.yml                    # build + push + deploy on main merge
│   │   └── proto.yml                     # regenerate stubs on proto changes
│   ├── CODEOWNERS
│   └── pull_request_template.md
│
├── proto/                                # shared gRPC definitions
│   ├── buf.yaml
│   ├── buf.gen.yaml
│   ├── auth/
│   │   └── auth.proto
│   ├── user/
│   │   └── user.proto
│   ├── billing/
│   │   └── billing.proto
│   ├── payment/
│   │   └── payment.proto
│   └── flags/
│       └── flags.proto
│
├── shared/
│   └── constants/
│       ├── kafka-topics.ts               # topic name constants (imported by NestJS services)
│       ├── kafka-topics.go               # same constants for Go services
│       └── kafka-topics.py               # same constants for Python service
│
├── infra/
│   ├── docker/
│   │   └── prometheus.yml               # Prometheus scrape config
│   ├── k8s/                             # Kubernetes manifests (future)
│   │   └── .gitkeep
│   └── terraform/                       # GCP infra as code (future)
│       └── .gitkeep
│
├── scripts/
│   ├── setup.sh                         # first-time dev environment setup
│   ├── start-all.sh                     # start all services
│   ├── migrate-all.sh                   # run migrations on all DBs
│   └── seed.sh                          # seed dev databases
│
├── docs/
│   ├── ARCHITECTURE.md                  # system overview, service map
│   ├── CONTRIBUTING.md                  # branching, PR, review process
│   ├── RUNBOOK.md                       # ops procedures
│   └── adr/                             # Architecture Decision Records
│       ├── 001-microservices.md
│       ├── 002-kafka-vs-redis-pubsub.md
│       └── 003-language-assignments.md
│
├── services/
│   │
│   ├── gateway/                         # [Go — Axum/Fiber] API Gateway
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   │   └── config.go
│   │   │   ├── proxy/
│   │   │   │   ├── proxy.go             # reverse proxy to upstream services
│   │   │   │   └── upstream.go          # upstream URL registry
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go              # validate token via Auth Service gRPC
│   │   │   │   ├── rate_limit.go        # Redis-backed rate limiter
│   │   │   │   ├── request_id.go        # UUID per request
│   │   │   │   └── logging.go
│   │   │   └── health/
│   │   │       └── health.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── auth-service/                    # [NestJS] Auth Service
│   │   ├── src/
│   │   │   ├── config/
│   │   │   │   ├── app.config.ts
│   │   │   │   ├── firebase.config.ts
│   │   │   │   ├── jwt.config.ts
│   │   │   │   └── index.ts
│   │   │   ├── infrastructure/
│   │   │   │   └── firebase/
│   │   │   │       ├── firebase.module.ts
│   │   │   │       ├── firebase.provider.ts
│   │   │   │       └── firebase-auth.adapter.ts  # implements AuthProviderPort
│   │   │   ├── modules/
│   │   │   │   └── auth/
│   │   │   │       ├── ports/
│   │   │   │       │   └── auth-provider.port.ts
│   │   │   │       ├── strategies/
│   │   │   │       │   ├── auth-strategy.interface.ts
│   │   │   │       │   ├── firebase-auth.strategy.ts
│   │   │   │       │   ├── jwt.strategy.ts
│   │   │   │       │   └── api-key.strategy.ts
│   │   │   │       ├── factories/
│   │   │   │       │   └── auth-strategy.factory.ts
│   │   │   │       ├── dto/
│   │   │   │       │   ├── login.dto.ts
│   │   │   │       │   ├── refresh.dto.ts
│   │   │   │       │   └── auth-user.dto.ts
│   │   │   │       ├── auth.guard.ts
│   │   │   │       ├── auth.controller.ts
│   │   │   │       ├── auth.service.ts
│   │   │   │       └── auth.module.ts
│   │   │   ├── common/
│   │   │   │   ├── decorators/
│   │   │   │   │   ├── public.decorator.ts
│   │   │   │   │   ├── auth-provider.decorator.ts
│   │   │   │   │   └── current-user.decorator.ts
│   │   │   │   ├── filters/
│   │   │   │   │   └── global-exception.filter.ts
│   │   │   │   ├── interceptors/
│   │   │   │   │   ├── transform.interceptor.ts
│   │   │   │   │   └── logging.interceptor.ts
│   │   │   │   ├── pipes/
│   │   │   │   │   └── zod-validation.pipe.ts
│   │   │   │   └── types/
│   │   │   │       └── api-response.type.ts
│   │   │   ├── app.module.ts
│   │   │   └── main.ts
│   │   ├── test/
│   │   │   ├── e2e/
│   │   │   │   └── auth.e2e-spec.ts
│   │   │   └── fixtures/
│   │   │       └── users.fixture.ts
│   │   ├── nest-cli.json
│   │   ├── tsconfig.json
│   │   ├── tsconfig.build.json
│   │   ├── package.json
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── user-service/                    # [Java — Spring Boot 3]
│   │   ├── src/
│   │   │   ├── main/
│   │   │   │   ├── java/in/roundup/users/
│   │   │   │   │   ├── config/
│   │   │   │   │   │   ├── KafkaConfig.java
│   │   │   │   │   │   └── GrpcConfig.java
│   │   │   │   │   ├── domain/
│   │   │   │   │   │   ├── entity/
│   │   │   │   │   │   │   ├── User.java
│   │   │   │   │   │   │   ├── Squad.java
│   │   │   │   │   │   │   ├── SquadMember.java
│   │   │   │   │   │   │   ├── DeviceToken.java
│   │   │   │   │   │   │   ├── UserFaceProfile.java
│   │   │   │   │   │   │   └── FriendConnection.java
│   │   │   │   │   │   └── repository/
│   │   │   │   │   │       ├── UserRepository.java
│   │   │   │   │   │       ├── SquadRepository.java
│   │   │   │   │   │       ├── SquadMemberRepository.java
│   │   │   │   │   │       ├── DeviceTokenRepository.java
│   │   │   │   │   │       └── FriendConnectionRepository.java
│   │   │   │   │   ├── service/
│   │   │   │   │   │   ├── UserService.java
│   │   │   │   │   │   ├── SquadService.java
│   │   │   │   │   │   ├── DeviceTokenService.java
│   │   │   │   │   │   └── FriendService.java
│   │   │   │   │   ├── controller/
│   │   │   │   │   │   ├── UserController.java
│   │   │   │   │   │   └── SquadController.java
│   │   │   │   │   ├── dto/
│   │   │   │   │   │   ├── CreateUserDto.java
│   │   │   │   │   │   ├── UpdateUserDto.java
│   │   │   │   │   │   └── UserResponseDto.java
│   │   │   │   │   ├── grpc/
│   │   │   │   │   │   └── UserGrpcServer.java
│   │   │   │   │   ├── kafka/
│   │   │   │   │   │   ├── UserEventPublisher.java
│   │   │   │   │   │   └── PaymentEventConsumer.java
│   │   │   │   │   └── UserServiceApplication.java
│   │   │   │   └── resources/
│   │   │   │       ├── application.yml
│   │   │   │       ├── application-dev.yml
│   │   │   │       └── db/migration/
│   │   │   │           ├── V1__create_users.sql
│   │   │   │           ├── V2__create_squads.sql
│   │   │   │           ├── V3__create_squad_members.sql
│   │   │   │           ├── V4__create_device_tokens.sql
│   │   │   │           ├── V5__create_user_face_profiles.sql
│   │   │   │           └── V6__create_friend_connections.sql
│   │   │   └── test/
│   │   │       └── java/in/roundup/users/
│   │   │           ├── service/
│   │   │           │   └── UserServiceTest.java
│   │   │           └── e2e/
│   │   │               └── UserControllerTest.java
│   │   ├── build.gradle
│   │   ├── settings.gradle
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── event-service/                   # [Java — Spring Boot 3]
│   │   ├── src/main/java/in/roundup/events/
│   │   │   ├── config/
│   │   │   ├── domain/
│   │   │   │   ├── entity/
│   │   │   │   │   ├── Event.java
│   │   │   │   │   ├── EventMember.java
│   │   │   │   │   ├── Vote.java
│   │   │   │   │   └── InviteLink.java
│   │   │   │   ├── repository/
│   │   │   │   └── statemachine/
│   │   │   │       ├── EventStatus.java
│   │   │   │       └── EventStateMachine.java
│   │   │   ├── service/
│   │   │   │   ├── EventService.java
│   │   │   │   ├── RsvpService.java
│   │   │   │   ├── VotingService.java
│   │   │   │   ├── InviteService.java
│   │   │   │   └── GuestService.java
│   │   │   ├── controller/
│   │   │   │   └── EventController.java
│   │   │   ├── dto/
│   │   │   ├── kafka/
│   │   │   │   ├── EventPublisher.java
│   │   │   │   └── EventConsumer.java
│   │   │   └── EventServiceApplication.java
│   │   ├── src/main/resources/
│   │   │   ├── application.yml
│   │   │   └── db/migration/
│   │   │       ├── V1__create_events.sql
│   │   │       ├── V2__create_event_members.sql
│   │   │       ├── V3__create_votes.sql
│   │   │       └── V4__create_invite_links.sql
│   │   ├── src/test/
│   │   ├── build.gradle
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── billing-service/                 # [Java — Spring Boot 3]
│   │   ├── src/main/java/in/roundup/billing/
│   │   │   ├── config/
│   │   │   ├── domain/
│   │   │   │   ├── entity/
│   │   │   │   │   ├── Bill.java
│   │   │   │   │   ├── BillItem.java
│   │   │   │   │   ├── ItemAssignment.java
│   │   │   │   │   ├── Debt.java
│   │   │   │   │   └── AuditLog.java
│   │   │   │   └── repository/
│   │   │   ├── service/
│   │   │   │   ├── BillService.java
│   │   │   │   └── DebtService.java
│   │   │   ├── patterns/
│   │   │   │   ├── command/
│   │   │   │   │   ├── BillCommand.java          # interface
│   │   │   │   │   ├── AddItemCommand.java
│   │   │   │   │   ├── RemoveItemCommand.java
│   │   │   │   │   ├── AssignItemCommand.java
│   │   │   │   │   ├── CloseBillCommand.java
│   │   │   │   │   └── CommandInvoker.java
│   │   │   │   └── strategy/
│   │   │   │       ├── SplitStrategy.java        # interface
│   │   │   │       ├── EqualSplitStrategy.java
│   │   │   │       ├── ByItemSplitStrategy.java
│   │   │   │       ├── CustomSplitStrategy.java
│   │   │   │       └── SplitStrategyFactory.java
│   │   │   ├── graph/
│   │   │   │   └── DebtGraph.java               # min-cash-flow algorithm
│   │   │   ├── controller/
│   │   │   │   └── BillingController.java
│   │   │   ├── dto/
│   │   │   ├── grpc/
│   │   │   │   └── BillingGrpcServer.java
│   │   │   ├── kafka/
│   │   │   │   ├── BillingPublisher.java
│   │   │   │   └── PaymentConsumer.java
│   │   │   └── BillingServiceApplication.java
│   │   ├── src/main/resources/
│   │   │   └── db/migration/
│   │   │       ├── V1__create_bills.sql
│   │   │       ├── V2__create_bill_items.sql
│   │   │       ├── V3__create_item_assignments.sql
│   │   │       ├── V4__create_debts.sql
│   │   │       ├── V5__create_audit_log.sql
│   │   │       └── V6__create_split_configs.sql
│   │   ├── src/test/
│   │   ├── build.gradle
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── payment-service/                 # [Go]
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   │   └── config.go
│   │   │   ├── juspay/
│   │   │   │   └── client.go            # Hyperswitch REST client
│   │   │   ├── idempotency/
│   │   │   │   └── store.go             # Redis-backed idempotency
│   │   │   ├── webhook/
│   │   │   │   └── handler.go
│   │   │   ├── service/
│   │   │   │   └── payment_service.go
│   │   │   ├── repository/
│   │   │   │   └── payment_repository.go
│   │   │   └── grpc/
│   │   │       └── server.go
│   │   ├── db/
│   │   │   ├── migrations/
│   │   │   │   ├── 0001_create_payment_requests.sql
│   │   │   │   ├── 0002_create_payment_events.sql
│   │   │   │   └── 0003_create_refunds.sql
│   │   │   ├── queries/
│   │   │   │   └── payment.sql          # sqlc input
│   │   │   └── sqlc.yaml
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── venue-service/                   # [Go]
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── places/
│   │   │   │   └── client.go            # Google Maps SDK wrapper
│   │   │   ├── cache/
│   │   │   │   └── venue_cache.go       # Redis cache layer
│   │   │   ├── service/
│   │   │   │   └── venue_service.go
│   │   │   ├── repository/
│   │   │   │   └── venue_repository.go
│   │   │   └── handler/
│   │   │       └── venue_handler.go
│   │   ├── db/
│   │   │   ├── migrations/
│   │   │   │   ├── 0001_create_saved_venues.sql
│   │   │   │   └── 0002_create_venue_visits.sql
│   │   │   └── queries/
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── notification-service/            # [Go]
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── fcm/
│   │   │   │   └── client.go
│   │   │   ├── whatsapp/
│   │   │   │   └── client.go
│   │   │   ├── router/
│   │   │   │   └── notification_router.go  # decides FCM vs WhatsApp
│   │   │   ├── consumer/
│   │   │   │   └── kafka_consumer.go       # all Kafka topic listeners
│   │   │   └── repository/
│   │   │       └── log_repository.go
│   │   ├── db/
│   │   │   └── migrations/
│   │   │       └── 0001_create_notification_log.sql
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── media-service/                   # [Go]
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── cloudinary/
│   │   │   │   └── client.go
│   │   │   ├── vision/
│   │   │   │   └── client.go            # Google Vision API
│   │   │   ├── matching/
│   │   │   │   └── face_matcher.go      # AWS Rekognition CompareFaces
│   │   │   ├── qr/
│   │   │   │   └── qr_service.go        # JWT signing + QR PNG gen
│   │   │   ├── download/
│   │   │   │   └── zip_streamer.go      # stream ZIP without buffering
│   │   │   ├── worker/
│   │   │   │   └── tagging_worker.go    # Kafka consumer → Vision → match
│   │   │   ├── service/
│   │   │   │   └── media_service.go
│   │   │   ├── repository/
│   │   │   │   └── media_repository.go
│   │   │   └── handler/
│   │   │       └── media_handler.go
│   │   ├── db/
│   │   │   └── migrations/
│   │   │       ├── 0001_create_photos.sql
│   │   │       ├── 0002_create_photo_faces.sql
│   │   │       └── 0003_create_download_tokens.sql
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── realtime-service/                # [Go]
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── hub/
│   │   │   │   ├── hub.go               # connection registry (sync.RWMutex)
│   │   │   │   └── connection.go        # single WebSocket connection
│   │   │   ├── presence/
│   │   │   │   └── presence.go          # Redis SET per event
│   │   │   ├── consumer/
│   │   │   │   └── kafka_consumer.go    # fan-out Kafka → WebSocket
│   │   │   ├── message/
│   │   │   │   └── types.go             # WsMessage struct + type constants
│   │   │   └── handler/
│   │   │       └── ws_handler.go        # WebSocket upgrade endpoint
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── reservation-service/             # [Go]
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── dineout/
│   │   │   │   └── client.go            # Dineout / EazyDiner API client
│   │   │   ├── statemachine/
│   │   │   │   └── reservation_sm.go
│   │   │   ├── service/
│   │   │   │   └── reservation_service.go
│   │   │   ├── repository/
│   │   │   │   └── reservation_repository.go
│   │   │   └── handler/
│   │   │       └── reservation_handler.go
│   │   ├── db/
│   │   │   └── migrations/
│   │   │       ├── 0001_create_reservations.sql
│   │   │       └── 0002_create_reservation_events.sql
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── ai-service/                      # [Python — FastAPI + LangChain]
│   │   ├── app/
│   │   │   ├── main.py                  # FastAPI app entry
│   │   │   ├── config.py                # Pydantic settings
│   │   │   ├── llm/
│   │   │   │   └── base.py              # shared LLM + embedding instances
│   │   │   ├── chains/
│   │   │   │   ├── receipt_ocr.py       # photo → line items
│   │   │   │   ├── nudge_generator.py   # personalised nudge text
│   │   │   │   ├── recap_generator.py   # outing summary
│   │   │   │   └── caption_generator.py # photo captions
│   │   │   ├── agents/
│   │   │   │   └── venue_recommender.py # LangChain agent + tools
│   │   │   ├── rag/
│   │   │   │   └── squad_history.py     # pgvector RAG over past outings
│   │   │   ├── models/
│   │   │   │   └── spend_predictor.py
│   │   │   ├── consumers/
│   │   │   │   └── kafka_consumer.py    # listens event.closed → embed
│   │   │   ├── routers/
│   │   │   │   ├── receipt.py
│   │   │   │   ├── venues.py
│   │   │   │   ├── query.py
│   │   │   │   ├── nudges.py
│   │   │   │   └── recap.py
│   │   │   └── schemas/                 # Pydantic request/response models
│   │   │       ├── receipt.py
│   │   │       ├── venue.py
│   │   │       └── nudge.py
│   │   ├── tests/
│   │   │   ├── test_receipt_ocr.py
│   │   │   ├── test_nudge_generator.py
│   │   │   └── test_rag.py
│   │   ├── pyproject.toml
│   │   ├── requirements.txt
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── analytics-service/               # [Java — Spring Boot 3]
│   │   ├── src/main/java/in/roundup/analytics/
│   │   │   ├── config/
│   │   │   ├── domain/
│   │   │   │   ├── entity/
│   │   │   │   │   └── AnalyticsEvent.java
│   │   │   │   └── repository/
│   │   │   │       └── AnalyticsEventRepository.java
│   │   │   ├── service/
│   │   │   │   ├── SquadStatsService.java
│   │   │   │   ├── UserStatsService.java
│   │   │   │   └── WrappedService.java
│   │   │   ├── controller/
│   │   │   │   └── AnalyticsController.java
│   │   │   ├── kafka/
│   │   │   │   └── AnalyticsConsumer.java  # ingests all domain events
│   │   │   └── AnalyticsServiceApplication.java
│   │   ├── src/main/resources/
│   │   │   └── db/migration/
│   │   │       ├── V1__create_analytics_events.sql
│   │   │       ├── V2__squad_stats_view.sql
│   │   │       ├── V3__user_stats_view.sql
│   │   │       ├── V4__venue_stats_view.sql
│   │   │       └── V5__monthly_summary_view.sql
│   │   ├── build.gradle
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   ├── feature-flag-service/            # [Go — or Rust]
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── evaluator/
│   │   │   │   └── flag_evaluator.go    # hash-based rollout
│   │   │   ├── cache/
│   │   │   │   └── flag_cache.go        # Redis cache (60s TTL)
│   │   │   ├── service/
│   │   │   │   └── flag_service.go
│   │   │   ├── repository/
│   │   │   │   └── flag_repository.go
│   │   │   ├── handler/
│   │   │   │   └── flag_handler.go      # REST — admin only
│   │   │   └── grpc/
│   │   │       └── server.go            # EvaluateFlag RPC
│   │   ├── db/
│   │   │   └── migrations/
│   │   │       ├── 0001_create_feature_flags.sql
│   │   │       └── 0002_create_user_overrides.sql
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   └── .env.example
│   │
│   └── admin-service/                   # [NestJS]
│       ├── src/
│       │   ├── config/
│       │   ├── modules/
│       │   │   ├── auth/                # admin-specific JWT auth
│       │   │   │   ├── admin-auth.guard.ts
│       │   │   │   ├── admin-auth.service.ts
│       │   │   │   └── admin-auth.module.ts
│       │   │   ├── users/               # user management
│       │   │   │   ├── admin-users.controller.ts
│       │   │   │   ├── admin-users.service.ts
│       │   │   │   └── admin-users.module.ts
│       │   │   ├── payments/            # payment ops
│       │   │   │   ├── admin-payments.controller.ts
│       │   │   │   ├── admin-payments.service.ts
│       │   │   │   └── admin-payments.module.ts
│       │   │   ├── flags/               # feature flag management
│       │   │   │   └── admin-flags.controller.ts
│       │   │   ├── venues/              # venue partner management
│       │   │   │   └── admin-venues.controller.ts
│       │   │   └── abuse/               # abuse detection
│       │   │       └── admin-abuse.controller.ts
│       │   ├── common/
│       │   │   ├── guards/
│       │   │   │   └── role.guard.ts
│       │   │   └── decorators/
│       │   │       └── roles.decorator.ts
│       │   ├── app.module.ts
│       │   └── main.ts
│       ├── nest-cli.json
│       ├── tsconfig.json
│       ├── package.json
│       ├── Dockerfile
│       └── .env.example
│
├── docker-compose.yml                   # all services + infra
├── docker-compose.override.yml          # local overrides (gitignored)
├── Makefile                             # convenience commands
├── .gitignore
├── .editorconfig
└── README.md
```

---

## Service → Language → Port map

| Service              | Language     | Internal Port | DB Port |
| -------------------- | ------------ | ------------- | ------- |
| gateway              | Go           | 8080          | —       |
| auth-service         | NestJS       | 3001          | 5433    |
| user-service         | Java         | 3002          | 5434    |
| event-service        | Java         | 3003          | 5435    |
| billing-service      | Java         | 3004          | 5436    |
| payment-service      | Go           | 3005          | 5437    |
| venue-service        | Go           | 3006          | 5438    |
| notification-service | Go           | 3007          | —       |
| media-service        | Go           | 3008          | 5439    |
| realtime-service     | Go           | 3009          | —       |
| reservation-service  | Go           | 3011          | 5442    |
| ai-service           | Python       | 3010          | 5441    |
| analytics-service    | Java         | 3012          | 5441    |
| feature-flag-service | Go (or Rust) | 3013          | 5440    |
| admin-service        | NestJS       | 3014          | —       |

_All services share Redis on 6379 and Kafka on 9092._
