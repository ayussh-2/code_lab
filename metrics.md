---

I have a distributed coding judge platform running locally. I need to measure two latency metrics to compare async vs sync submission processing. Here's what to do:

**Setup check first:**
- Find and start all services (look for a Makefile, docker-compose, or README for instructions)
- Identify the submission API endpoint (look in the routes/handler files)
- Identify where Docker sandbox execution happens (likely in `backend/internal/sandbox/` or similar)

**Measurement 1 — Async end-to-end latency:**
- Write a script (Go or Python, whichever is easier) that:
  1. POSTs a submission to the API with a simple Hello World in Python
  2. Immediately starts a timer
  3. Polls the submission status endpoint every 100ms until verdict is no longer `PENDING`
  4. Stops the timer and records the delta
- Run this 20 times, print min/max/average

**Measurement 2 — API response time (most important):**
- Time only how long the POST /submit endpoint takes to return a response
- This should be near-instant since it's async (just enqueues to NATS)
- Run 20 times, print average

**Measurement 3 — Sync baseline simulation:**
- Find the Docker execution function in the sandbox code
- Write a small benchmark that calls it directly 20 times with a Hello World Python script, timing each call
- Print min/max/average — this simulates what blocking/sync would cost

**Finally print a summary like:**
```
Async API response time (submit endpoint): Xms avg
Sync sandbox execution time (direct Docker call): Xms avg
Improvement: X% reduction in API blocking time
End-to-end verdict latency (async): Xms avg
```


---
