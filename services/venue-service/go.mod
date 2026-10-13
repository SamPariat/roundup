module github.com/SamPariat/roundup/services/venue-service

go 1.26.1

require (
	github.com/SamPariat/roundup/shared/go v0.0.0
	github.com/gofiber/fiber/v3 v3.6.0
	github.com/jackc/pgx/v5 v5.11.0
	github.com/redis/go-redis/v9 v9.23.0
	github.com/stretchr/testify v1.12.1
	go.uber.org/zap v1.28.0
	googlemaps.github.io/maps v1.7.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gofiber/schema v1.8.8 // indirect
	github.com/gofiber/utils/v2 v2.6.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/molecule-man/go-brrr v1.2.0 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/tinylib/msgp v1.6.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.75.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
)

replace github.com/SamPariat/roundup/shared/go => ../../shared/go
