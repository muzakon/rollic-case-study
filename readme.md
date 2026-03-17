### Rollic Case Study

I implemented all endpoints as described in the case study, with only minor changes and without over-engineering.

### Installation
1. Create your `.env` file in the project root. You can copy the contents of `.env.example` and modify them based on your requirements (e.g., available ports).
2. Run `make build`. I created a `Makefile` for convenience. If you cannot run `make` commands, execute: `docker compose -f docker-compose.yml up -d --build` after creating and updating your `.env` file. This will handle everything for you—creating services, initializing the database, running migrations, etc.
3. For testing, run `make test-e2e` or `docker compose exec server gotestsum --format testdox --hide-summary=skipped -- ./tests/e2e/... -count=1` (make sure the `server` and `postgres` services are healthy).

### Tech Stack
- Go (Fiber v3)
- PostgreSQL (Goose for migrations & GORM with PostgreSQL driver)
- GoCron for cron jobs
- Zerolog for structured logging
- End-to-end tests (no unit tests)
- Docker
- Git workflow configured to run tests in the cloud

---

### Changes I Applied

1. I return `"id"` instead of `"boardId"` in board-related endpoints (e.g., list, create). Example response:
```
{
    "id": "7c61e1c2-3761-4d9d-85bc-7e7ec485af40",
    "name": "test",
    "description": "Global leaderboard for weekly tournament",
    "createdAt": "2026-03-17T15:56:07.084567634Z",
    "schedule": {
        "type": "interval",
        "intervalSeconds": 60
    },
    "nextResetAt": "2026-03-17T15:57:07.082282259Z"
}
```

2. As you might have noticed, the `id` is an auto-generated UUID rather than a custom string (e.g., `"board_{timestamp}"`).

3. I improved validation error messages. For example:
```
{
    "details": [
        {
            "field": "name",
            "message": "field is required"
        }
    ],
    "error": "Validation failed"
}
```

4. I added optional cursor-based pagination to the board listing endpoint. If you provide a `limit` query parameter, pagination is applied when necessary.

```
{
    "data": [
        {
            "id": "8dd690a3-2a67-4c7e-8e7a-5519733e14d6",
            "name": "test",
            "description": "Global leaderboard for weekly tournament"
        },
        {
            "id": "7c61e1c2-3761-4d9d-85bc-7e7ec485af40",
            "name": "test",
            "description": "Global leaderboard for weekly tournament"
        }
    ],
    "totalCount": 2,
    "limit": 0,
    "hasNext": false,
    "cursor": null
}
```

This could be much more improved.

### Limitations & Things I'd change in Production
- Although Go is a great choice for high-performance tasks, this API heavily depends on the PostgreSQL instance. In cloud environments, it is highly recommended to use serverless, horizontally scalable PostgreSQL solutions such as AWS Aurora or CockroachDB (PostgreSQL-compatible, but not identical).
- I would **not** use an in-process cron job in production. The current scheduler runs every minute, queries all boards where `next_reset_at <= now`, and deletes scores synchronously. There are significantly better cloud-native solutions.
- I would prefer deploying on platforms such as AWS Fargate, EKS, or Google Cloud Run.
- If handling millions of records (boards, scores, etc.), I would introduce partitioning and limit batch sizes. Large `DELETE` operations (e.g., boards with 100K+ scores) can lock tables and block reads. Processing thousands of boards in a single cycle can also create significant write spikes.


### How would i design the Scheduler in Production?
1. Remove the `gocron` dependency entirely.  
2. Use AWS EventBridge or Google Cloud Scheduler to trigger an internal endpoint or serverless function.  
3. Implement a dispatcher using AWS Lambda or Google Cloud Functions.  
4. The dispatcher queries boards due for reset using cursor-based pagination:  
   `SELECT ... WHERE next_reset_at <= NOW() AND id > $last_seen_id ORDER BY id ASC LIMIT 500`  
   Each batch of 500 board IDs is published as a single message to SQS or Pub/Sub. For 500K boards, this results in ~1000 messages. With batch publishing, this completes within seconds.  
5. Worker functions (Lambda / Cloud Functions) consume messages with a concurrency limit (e.g., 20), capping database connections and preventing overload regardless of scale.  
6. Instead of deleting scores, introduce a `generation` column in both `boards` and `scores`. Each reset increments the generation:  
   `UPDATE boards SET generation = generation + 1, next_reset_at = previous_reset_at + interval WHERE id = $1`  
   Leaderboard queries filter by `WHERE s.generation = b.generation`, making old scores instantly invisible without costly deletes.  
7. Clean up stale scores (`generation < board.generation`) via a separate scheduled job during off-peak hours (e.g., daily at 03:00). Deletions should be batched (e.g., 10K rows) with small delays between batches to avoid impacting live traffic.  
8. Route all read-heavy leaderboard queries to read replicas (e.g., Aurora Read Replicas on AWS) to isolate reads from write operations.