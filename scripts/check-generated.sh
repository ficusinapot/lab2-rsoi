#!/usr/bin/env bash
set -euo pipefail

for service in reservation payment loyalty; do
  make -C "services/$service" generate
done
paths=(services/reservation/ent services/payment/ent services/loyalty/ent
       services/reservation/internal/usecase/bookings/repository_mock_test.go
       services/reservation/internal/usecase/hotels/repository_mock_test.go)
git diff --exit-code -- "${paths[@]}"
if [[ -n $(git ls-files --others --exclude-standard -- "${paths[@]}") ]]; then
  echo 'Generation produced untracked files' >&2
  exit 1
fi
