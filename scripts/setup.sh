for service in services/*/; do
  if [ -f "$service/.env.example" ]; then
    cp "$service/.env.example" "$service/.env"
  fi
done

git config core.hooksPath .githooks
echo "Git hooks path set to .githooks (pre-commit runs gofmt/go vet/golangci-lint on staged Go files)"
