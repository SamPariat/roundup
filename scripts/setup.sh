for service in services/*/; do
  if [ -f "$service/.env.example" ]; then
    cp "$service/.env.example" "$service/.env"
  fi
done
